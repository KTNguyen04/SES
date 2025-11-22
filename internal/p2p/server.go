package p2p

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"

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
	peerPort := md.Get("port")
	if len(peerPort) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "missing peer port in metadata")
	}
	id := md.Get("id")
	if len(id) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "missing peer id in metadata")
	}
	peerId, err := strconv.Atoi(id[0])
	if err != nil {
		log.Printf("invalid peer id: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid peer id")
	}
	peer := Peer{
		Id:      peerId,
		Address: p.Addr.(*net.TCPAddr).IP.String(),
		Port:    peerPort[0],
	}
	if len(peerPort) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "missing peer port in metadata")
	}
	log.Println("Received ping request from", p.Addr)
	log.Printf("Trying to connect back to peer %s:%s", p.Addr.(*net.TCPAddr).IP.String(), peerPort[0])
	if host.IfConnectedToPeer(peer) {
		log.Printf("Already connected to peer %s:%s", p.Addr.(*net.TCPAddr).IP.String(), peerPort[0])
		return &emptypb.Empty{}, nil
	}
	err = host.DialToPeer(peer)
	if err != nil {
		log.Printf("Cannot connect back to peer %s:%s: %v", p.Addr.(*net.TCPAddr).IP.String(), peerPort[0], err)
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (host *Host) Disconnect(ctx context.Context, req *comm.Request) (*emptypb.Empty, error) {
	id, err := strconv.Atoi(req.Id)
	if err != nil {
		log.Printf("invalid peer id: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid peer id")
	}
	host.ClosePeerConnection(id)
	return &emptypb.Empty{}, nil
}
func (host *Host) Send(ctx context.Context, msg *comm.Message) (*emptypb.Empty, error) {
	log.Printf("Received message from peer %s: %s", msg.From, msg.Content)
	peerId, err := strconv.Atoi(msg.From)
	if err != nil {
		log.Printf("invalid peer id in message: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid peer id in message")
	}

	host.mu.Lock()
	host.Vvt.V[host.Id].T[host.Id]++
	host.mu.Unlock()
	canDeliver := host.checkCanDeliver(msg)

	if canDeliver {
		host.deliver(msg)
		host.checkBufferedMessages()
		host.mergeVvector(msg.Vvt, peerId)
	} else {
		host.mu.Lock()
		host.Vvt.V[host.Id].T[host.Id]--
		host.BufferedMessages = append(host.BufferedMessages, msg)
		log.Printf("Buffered message from peer %s: %s", msg.From, msg.Content)
		host.mu.Unlock()
	}

	return &emptypb.Empty{}, nil
}

func (host *Host) checkCanDeliver(msg *comm.Message) bool {
	host.mu.RLock()
	defer host.mu.RUnlock()
	log.Printf("--------------------------------\n")
	log.Printf("Checking if can deliver message from peer %s: %s\n", msg.From, msg.Content)
	log.Printf("Message Vvector:\n")
	for i, v := range msg.Vvt.V {
		line := fmt.Sprintf("V[%d]:", i)
		for _, t := range v.T {
			line += fmt.Sprintf(" %d", t)
		}
		log.Println(line)
	}
	log.Printf("Host Vvector:\n")
	for i, v := range host.Vvt.V {
		line := fmt.Sprintf("V[%d]:", i)
		for _, t := range v.T {
			line += fmt.Sprintf(" %d", t)
		}
		log.Println(line)
	}
	log.Printf("End\n")
	log.Printf("--------------------------------\n")

	if msg.Vvt.V[host.Id].T == nil {
		return true
	}
	for i := range msg.Vvt.V[host.Id].T {
		if host.Vvt.V[host.Id].T[i] < msg.Vvt.V[host.Id].T[i] {
			return false
		}
	}
	return true
}

func (host *Host) deliver(msg *comm.Message) {
	log.Printf("Delivered message from peer %s: %s", msg.From, msg.Content)
}
func (host *Host) checkBufferedMessages() {

	host.mu.RLock()
	defer host.mu.RUnlock()
	if len(host.BufferedMessages) == 0 {
		return
	}
	newBuf := host.BufferedMessages[:0]
	for _, msg := range host.BufferedMessages {
		if host.checkCanDeliver(msg) {
			log.Printf("deliver buffered message from peer %s: %s", msg.From, msg.Content)
			host.deliver(msg)
		} else {
			newBuf = append(newBuf, msg)
		}
	}
	host.BufferedMessages = newBuf
}

func (host *Host) mergeVvector(msgVvt *comm.Vvector, peerId int) {
	host.mu.Lock()
	defer host.mu.Unlock()
	log.Printf("--------------------------------\n")
	log.Printf("Merging Vvectors after receiving message\n")
	log.Printf("host Vvector before merge:\n")
	for i, v := range host.Vvt.V {
		line := fmt.Sprintf("V[%d]:", i)
		for _, t := range v.T {
			line += fmt.Sprintf(" %d", t)
		}
		log.Println(line)
	}
	log.Printf("message Vvector:\n")
	for i, v := range msgVvt.V {
		line := fmt.Sprintf("V[%d]:", i)
		for _, t := range v.T {
			line += fmt.Sprintf(" %d", t)
		}
		log.Println(line)
	}
	// merge self timestamp vector
	if host.Vvt.V[host.Id].T == nil {
		host.Vvt.V[host.Id].T = make([]int64, len(msgVvt.V[peerId].T))
		copy(host.Vvt.V[host.Id].T, msgVvt.V[peerId].T)
	} else {
		for j := range host.Vvt.V[host.Id].T {
			if msgVvt.V[peerId].T[j] > host.Vvt.V[host.Id].T[j] {
				host.Vvt.V[host.Id].T[j] = msgVvt.V[peerId].T[j]
			}
		}
	}
	// merge other processes' timestamp vectors
	for i := range msgVvt.V {
		if i == host.Id || i == peerId {
			continue
		}
		if msgVvt.V[i].T == nil {
			continue
		}
		if host.Vvt.V[i].T == nil && msgVvt.V[i].T != nil {
			host.Vvt.V[i].T = make([]int64, len(msgVvt.V[i].T))
			copy(host.Vvt.V[i].T, msgVvt.V[i].T)
			continue
		}

		for j := range msgVvt.V[i].T {
			if msgVvt.V[i].T[j] > host.Vvt.V[i].T[j] {
				host.Vvt.V[i].T[j] = msgVvt.V[i].T[j]
			}
		}
	}
	log.Printf("host Vvector after merge:\n")
	for i, v := range host.Vvt.V {
		line := fmt.Sprintf("V[%d]:", i)
		for _, t := range v.T {
			line += fmt.Sprintf(" %d", t)
		}
		log.Println(line)
	}
	log.Printf("End\n")
	log.Printf("--------------------------------\n")
}
