package p2p

import (
	"context"
	"fmt"
	"log"
	"net"

	comm "github.com/KTNguyen04/SES/communication"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (host *Host) RunServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf("%v:%v", host.Address, host.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Listening for other peers on port %v", host.Port)
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	comm.RegisterCommunicationServer(grpcServer, host)
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (host *Host) Ping(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	p, _ := peer.FromContext(ctx)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "missing metadata")
	}
	peerPort := md.Get("self-port")
	if len(peerPort) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "missing peer port in metadata")
	}
	log.Println("Received ping request from", p.Addr)
	log.Printf("Trying to connect back to peer %s:%s", p.Addr.(*net.TCPAddr).IP.String(), peerPort[0])
	if host.IfConnectedToPeer(p.Addr.(*net.TCPAddr).IP.String(), peerPort[0]) {
		log.Printf("Already connected to peer %s:%s", p.Addr.(*net.TCPAddr).IP.String(), peerPort[0])
		return &emptypb.Empty{}, nil
	}
	err := host.DialToPeer(p.Addr.(*net.TCPAddr).IP.String(), peerPort[0])
	if err != nil {
		log.Printf("Cannot connect back to peer %s:%s: %v", p.Addr.(*net.TCPAddr).IP.String(), peerPort[0], err)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
