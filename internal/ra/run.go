package ra

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Run(ctx context.Context, id, addr string, peersCSV string, logdir string, interval time.Duration, csHold time.Duration, every int) error {
	if logdir != "" {
		if err := os.MkdirAll(logdir, 0o755); err != nil { return fmt.Errorf("make logdir: %w", err) }
		f, err := os.OpenFile(filepath.Join(logdir, id+".log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil { return fmt.Errorf("open log file: %w", err) }
		log.SetOutput(io.MultiWriter(os.Stdout, f))
	}

	var addrs []string
	if peersCSV != "" {
		for _, a := range strings.Split(peersCSV, ",") {
			addrs = append(addrs, strings.TrimSpace(a))
		}
	} else {
		addrs = []string{addr}
	}

	n := &Node{
		id:      id,
		addr:    addr,
		addrs:   addrs,
		st:      RELEASED,
		clock:   &lamport{id: id},
		deferQ:  map[string]bool{},
		replyCh: make(chan struct{}, 1024),
		csHold:  csHold,
	}

	go func() {
		if err := n.serve(); err != nil { log.Printf("serve error: %v", err) }
	}()

	// Retry dialing peers with backoff, then start workload
	backoff := 250 * time.Millisecond
	for {
		if err := n.dialPeers(); err != nil {
			log.Printf("peer dial failed: %v; retrying in %v", err, backoff)
			time.Sleep(backoff)
			if backoff < 5*time.Second { backoff *= 2 }
			continue
		}
		log.Printf("[%s] peers connected: %d", n.id, n.peerCount()-1)
		n.startWorkload(interval, every)
		break
	}

	<-ctx.Done()
	return nil
}
