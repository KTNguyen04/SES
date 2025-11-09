package ses

type Peers struct {
	Peers []Host
}

func (p *Peers) AddPeer(id int, address string, port int) {
	p.Peers = append(p.Peers, Host{
		Id:      id,
		Address: address,
		Port:    port,
	})
}
