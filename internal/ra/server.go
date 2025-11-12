package ra

import (
	"net"

	"google.golang.org/grpc"

	pb "example.com/ra/proto"
)

func (n *Node) serve() error {
	s := grpc.NewServer()
	pb.RegisterMutexServer(s, n)
	lis, err := net.Listen("tcp", n.addr)
	if err != nil {
		return err
	}
	return s.Serve(lis)
}
