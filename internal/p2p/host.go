package p2p

import (
	comm "github.com/KTNguyen04/SES/communication"
	"google.golang.org/grpc"
)

type Host struct {
	Id                                    int
	Address                               string
	Port                                  string
	Vvt                                   comm.Vvector
	selfClient                            comm.CommunicationClient
	selfConn                              *grpc.ClientConn
	ActivePeers                           []Peer
	comm.UnimplementedCommunicationServer // implement interface
}

type Peer struct {
	Id      int
	Address string
	Port    string
}

func NewHost(id int, addr string, port string) *Host {
	return &Host{
		Vvt:        comm.Vvector{},
		Id:         id,
		Address:    addr,
		Port:       port,
		selfClient: nil,
	}
}

func (host *Host) IfConnectedToPeer(peerAddr string, peerPort string) bool {
	for _, peer := range host.ActivePeers {
		if peer.Address == peerAddr && peer.Port == peerPort {
			return true
		}
	}
	return false
}
