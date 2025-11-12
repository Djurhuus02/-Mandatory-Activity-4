package ra

import (
	"context"
	"log"
	"time"
)

func (n *Node) startWorkload(pace time.Duration, every int) {
	if every < 1 {
		every = 1
	}
	if pace <= 0 {
		pace = 500 * time.Millisecond
	}
	ctx, cancel := context.WithCancel(context.Background())
	n.stopWork = cancel

	t := time.NewTicker(pace)
	defer t.Stop()

	i := 0
	log.Printf("[%s] workload: tick every=%v, request every=%d", n.id, pace, every)
	for {
		select {
		case <-t.C:
			i++
			if i%every == 0 {
				n.requestCS()
			}
		case <-ctx.Done():
			return
		}
	}
}

func (n *Node) requestCS() {
	n.mu.Lock()
	if n.st != RELEASED {
		n.mu.Unlock()
		return
	}
	n.st = WANTED
	ts := n.clock.tick()
	// stash our request timestamp to compare in Request handler
	n.clock.req = ts
	n.replyCh = make(chan struct{}, 1024)
	total := n.peerCount()
	log.Printf("[%s] WANTS CS (reqTs=%d), waiting for %d replies", n.id, ts, total-1)
	n.mu.Unlock()

	n.sendRequestToAll(ts)

	for i := 0; i < total-1; i++ {
		<-n.replyCh
	}

	n.mu.Lock()
	n.st = HELD
	log.Printf("[%s] ENTER CS", n.id)
	n.mu.Unlock()

	n.criticalSectionWork()

	n.mu.Lock()
	n.st = RELEASED
	deferCount := 0
	for pid := range n.deferQ {
		deferCount++
		go n.sendReply(pid)
	}
	n.deferQ = map[string]bool{}
	log.Printf("[%s] EXIT CS -> flushed %d deferred replies", n.id, deferCount)
	n.mu.Unlock()
}
