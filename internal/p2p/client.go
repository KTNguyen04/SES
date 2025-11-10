package p2p

import (
	"context"
	"fmt"
	"log"
	"time"

	comm "github.com/KTNguyen04/SES/communication"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (host *Host) DialToPeer(targetAddr string, targetPort string) error {
	peerAddr := fmt.Sprintf("%v:%v", targetAddr, targetPort)
	log.Printf("Trying to dial to peer %s", peerAddr)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.NewClient(peerAddr, opts...)
	if err != nil {
		log.Printf("Cannot dial to peer %s: %v", peerAddr, err)
		return err
	}
	// defer conn.Close()
	host.selfConn = conn
	host.selfClient = comm.NewCommunicationClient(conn)
	log.Printf("Dialing to peer %s", peerAddr)
	return nil

}
func (host *Host) ClosePeerConnection() {
	if host.selfConn != nil {
		host.selfConn.Close()
	}
}

func (host *Host) Inform(targetAddr string, targetPort string) error {
	//Ping for connection
	if host.IfConnectedToPeer(targetAddr, targetPort) {
		log.Printf("Already connected to peer %s:%s", targetAddr, targetPort)
		return nil
	}
	peerAddr := fmt.Sprintf("%v:%v", targetAddr, targetPort)
	md := metadata.Pairs("self-port", host.Port)
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, err := host.selfClient.Ping(ctx, &emptypb.Empty{}, grpc.WaitForReady(true))
	if err != nil {
		log.Printf("Cannot connect to peer %s: %v", peerAddr, err)
		return err
	}
	host.ActivePeers = append(host.ActivePeers, Peer{
		Address: targetAddr,
		Port:    targetPort,
	})
	log.Printf("Connected to peer %s successfully", peerAddr)

	return nil
}
