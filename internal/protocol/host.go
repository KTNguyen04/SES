package ses

import comm "github.com/KTNguyen04/SES/communication"

type Host struct {
	Id                                    int
	Address                               string
	Port                                  int
	Vvt                                   comm.Vvector
	comm.UnimplementedCommunicationServer // implement interface
}

func NewHost(id int, addr string, port int) *Host {
	return &Host{
		Vvt:     comm.Vvector{},
		Id:      id,
		Address: addr,
		Port:    port,
	}
}
