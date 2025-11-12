package ra

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "example.com/ra/proto"
)

func (n *Node) dialPeers() error {
	n.peersMu.Lock()
	n.peers = map[string]pb.MutexClient{}
	n.peers[n.id] = nil

	flush := make([]string, 0)

	for _, addr := range n.addrs {
		if addr == "" || addr == n.addr {
			continue
		}
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			n.peersMu.Unlock()
			return fmt.Errorf("dial %s: %w", addr, err)
		}
		c := pb.NewMutexClient(conn)
		// ping to determine peer id
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		resp, err := c.Ping(ctx, &pb.PingReq{})
		cancel()
		if err != nil {
			n.peersMu.Unlock()
			return fmt.Errorf("ping %s: %w", addr, err)
		}
		n.peers[resp.Id] = c
		flush = append(flush, resp.Id)
	}
	n.peersMu.Unlock()

	// Send any pending replies for peers we couldn't reach earlier
	for _, pid := range flush {
		n.mu.Lock()
		p := n.pending != nil && n.pending[pid]
		n.mu.Unlock()
		if p {
			n.sendReply(pid)
			n.mu.Lock()
			delete(n.pending, pid)
			n.mu.Unlock()
		}
	}
	return nil
}

func (n *Node) sendRequestToAll(ts int64) {
	n.peersMu.RLock()
	defer n.peersMu.RUnlock()
	for pid, c := range n.peers {
		if pid == n.id || c == nil {
			continue
		}
		go func(pid string, c pb.MutexClient) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_, err := c.Request(ctx, &pb.RequestMsg{Timestamp: ts, From: n.id})
			if err != nil {
				log.Printf("[%s] ERROR REQUEST -> %s: %v", n.id, pid, err)
			} else {
				log.Printf("[%s] sent REQUEST(ts=%d) -> %s", n.id, ts, pid)
			}
		}(pid, c)
	}
}

func (n *Node) sendReply(pid string) {
	n.peersMu.RLock()
	c, ok := n.peers[pid]
	n.peersMu.RUnlock()
	if !ok || c == nil {
		log.Printf("[%s] cannot REPLY: unknown peer %s (queue and retry later)", n.id, pid)
		n.mu.Lock()
		if n.pending == nil {
			n.pending = map[string]bool{}
		}
		n.pending[pid] = true
		n.mu.Unlock()
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := c.Reply(ctx, &pb.ReplyMsg{Timestamp: n.clock.tick(), From: n.id})
	if err != nil {
		log.Printf("[%s] ERROR REPLY -> %s: %v", n.id, pid, err)
	}
}
