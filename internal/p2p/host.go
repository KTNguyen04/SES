package p2p

import (
	"context"
	"fmt"
	"log"

	comm "github.com/KTNguyen04/SES/communication"
	"google.golang.org/grpc"
)

type Host struct {
	Id                                    int
	Address                               string
	Port                                  string
	Vvt                                   comm.Vvector
	SelfClientsToPeer                     map[int]comm.CommunicationClient
	SelfConnsToPeer                       map[int]*grpc.ClientConn
	ActivePeers                           []Peer
	comm.UnimplementedCommunicationServer // implement interface
	BufferedMessages                      []*comm.Message
}

type Peer struct {
	Id      int
	Address string
	Port    string
}

func NewHost(id int, addr string, port string) *Host {
	return &Host{
		Vvt:               comm.Vvector{},
		Id:                id,
		Address:           addr,
		Port:              port,
		SelfClientsToPeer: make(map[int]comm.CommunicationClient),
		SelfConnsToPeer:   make(map[int]*grpc.ClientConn),
	}
}

func (host *Host) IfConnectedToPeer(peer Peer) bool {
	if _, ok := host.SelfConnsToPeer[peer.Id]; ok {
		return true
	}
	return false
}

func (host *Host) SESSendMessage(to int, content string) error {
	if !host.IfConnectedToPeer(Peer{Id: to}) {
		log.Printf("Not connected to peer %d yet", to)
		return fmt.Errorf("not connected to peer %d yet", to)
	}

	host.Vvt.V[host.Id].T[host.Id]++

	_, err := host.SelfClientsToPeer[to].Send(context.Background(), &comm.Message{
		From:    fmt.Sprintf("%d", host.Id),
		To:      fmt.Sprintf("%d", to),
		Content: content,
		Vvt:     &host.Vvt,
	})
	if err != nil {
		log.Printf("Failed to send message to peer %d: %v", to, err)
		return err
	}
	log.Printf("Sent message to peer %d: %s", to, content)
	log.Printf("host Vvector after sending message to %d:\n", to)
	for i, v := range host.Vvt.V {
		line := fmt.Sprintf("V[%d]:", i)
		for _, t := range v.T {
			line += fmt.Sprintf(" %d", t)
		}
		log.Println(line)
	}
	if host.Vvt.V[to].T == nil {
		host.Vvt.V[to].T = make([]int64, len(host.Vvt.V))
		copy(host.Vvt.V[to].T, host.Vvt.V[host.Id].T)
	} else {
		for i := range host.Vvt.V[to].T {
			if host.Vvt.V[host.Id].T[i] > host.Vvt.V[to].T[i] {
				host.Vvt.V[to].T[i] = host.Vvt.V[host.Id].T[i]
			}
		}
	}

	return err
}
