package p2p

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	comm "github.com/KTNguyen04/SES/communication"
	ses "github.com/KTNguyen04/SES/internal/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func RunServer(host *ses.Host) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%v:%v", host.Address, host.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Listening on port %d", host.Port)
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	comm.RegisterCommunicationServer(grpcServer, host)
	if err = grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// Connect to peer
func RunClient(targetAddr, targetPort string) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	serverAddr := fmt.Sprintf("%s:%s", targetAddr, targetPort)
	conn, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	chatClient := comm.NewCommunicationClient(conn)

	//Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = chatClient.Ping(ctx, &emptypb.Empty{}, grpc.WaitForReady(true))
	if err != nil {
		log.Fatalf("Cannot connect to server (%s): %v", serverAddr, err)
	}
	log.Printf("Connected to peer %s", serverAddr)
}
