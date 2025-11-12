package ra

import (
	"context"
	"log"
	"sync"
	"time"

	pb "example.com/ra/proto"
)

type Node struct {
	pb.UnimplementedMutexServer

	id, addr string

	peersMu sync.RWMutex
	peers   map[string]pb.MutexClient
	addrs   []string

	mu       sync.Mutex
	st       state
	clock    *lamport
	replyCh  chan struct{}
	deferQ   map[string]bool
	pending  map[string]bool
	csHold   time.Duration
	stopWork context.CancelFunc
}

func (n *Node) peerCount() int {
	n.peersMu.RLock()
	defer n.peersMu.RUnlock()
	return len(n.peers)
}

func (n *Node) criticalSectionWork() {
	log.Printf("[%s] *** CRITICAL SECTION START ***", n.id)
	hold := n.csHold
	if hold <= 0 {
		hold = 800 * time.Millisecond
	}
	time.Sleep(hold)
	log.Printf("[%s] safely updated shared resource at T=%d", n.id, n.clock.now())
	log.Printf("[%s] *** CRITICAL SECTION END ***", n.id)
}

