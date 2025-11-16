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

func (host *Host) DialToPeer(peer Peer) error {
	peerAddr := fmt.Sprintf("%v:%v", peer.Address, peer.Port)
	log.Printf("Trying to dial to peer %s", peerAddr)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.NewClient(peerAddr, opts...)
	if err != nil {
		log.Printf("Cannot dial to peer %s: %v", peerAddr, err)
		return err
	}
	// defer conn.Close()
	if host.IfConnectedToPeer(peer) {
		log.Printf("Already connected to peer %s:%s", peer.Address, peer.Port)
		return nil
	}
	host.SelfConnsToPeer[peer.Id] = conn
	host.SelfClientsToPeer[peer.Id] = comm.NewCommunicationClient(conn)

	log.Printf("Dialing to peer %s", peerAddr)
	return nil

}
func (host *Host) ClosePeerConnection(id int) {
	conn, ok := host.SelfConnsToPeer[id]
	if !ok {
		return
	}
	_ = conn.Close()
	delete(host.SelfConnsToPeer, id)
	delete(host.SelfClientsToPeer, id)
	log.Printf("Closed connection to peer %d", id)
}

func (host *Host) Pinging(peer Peer) error {
	//Ping for connection
	targetAddr := peer.Address
	targetPort := peer.Port
	if !host.IfConnectedToPeer(peer) {
		log.Printf("Not connected to peer %s:%s yet", targetAddr, targetPort)
		return fmt.Errorf("not connected to peer %s:%s yet", targetAddr, targetPort)
	}
	peerAddr := fmt.Sprintf("%v:%v", targetAddr, targetPort)
	md := metadata.New(map[string]string{
		"id": fmt.Sprintf("%v", host.Id),
		// "addr": host.Address,
		"port": host.Port,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, err := host.SelfClientsToPeer[peer.Id].Ping(ctx, &emptypb.Empty{}, grpc.WaitForReady(true))
	if err != nil {
		log.Printf("Not connected to peer %s: %v", peerAddr, err)
		return err
	}
	host.ActivePeers = append(host.ActivePeers, Peer{
		Address: targetAddr,
		Port:    targetPort,
	})
	log.Printf("Ping to peer %s successfully", peerAddr)

	return nil
}
