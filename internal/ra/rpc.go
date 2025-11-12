package ra

import (
	"context"
	"log"

	pb "example.com/ra/proto"
)

func (n *Node) Ping(ctx context.Context, _ *pb.PingReq) (*pb.PingResp, error) {
	return &pb.PingResp{Id: n.id}, nil
}

func lessTuple(aTs int64, aID string, bTs int64, bID string) bool {
	if aTs != bTs {
		return aTs < bTs
	}
	return aID < bID
}

func (n *Node) Request(ctx context.Context, r *pb.RequestMsg) (*pb.Empty, error) {
	n.clock.merge(r.Timestamp)
	n.mu.Lock()
	defer n.mu.Unlock()

	shouldDefer := n.st == HELD || (n.st == WANTED && lessTuple(n.clock.req, n.id, r.Timestamp, r.From))
	if shouldDefer {
		if n.deferQ == nil {
			n.deferQ = map[string]bool{}
		}
		n.deferQ[r.From] = true
		log.Printf("[%s] recv REQUEST(ts=%d, from=%s) -> DEFER (state=%s, myReqTs=%d)", n.id, r.Timestamp, r.From, n.st, n.clock.req)
	} else {
		log.Printf("[%s] recv REQUEST(ts=%d, from=%s) -> REPLY", n.id, r.Timestamp, r.From)
		go n.sendReply(r.From)
	}
	return &pb.Empty{}, nil
}

func (n *Node) Reply(ctx context.Context, r *pb.ReplyMsg) (*pb.Empty, error) {
	n.clock.merge(r.Timestamp)
	select {
	case n.replyCh <- struct{}{}:
	default:
	}
	log.Printf("[%s] recv REPLY(from=%s, ts=%d)", n.id, r.From, r.Timestamp)
	return &pb.Empty{}, nil
}
