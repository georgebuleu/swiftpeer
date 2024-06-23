package peer

import (
	"fmt"
	"net"
)

type Peer struct {
	IP     string
	Port   int
	PeerId string
}

func (p Peer) FormatAddress() (string, error) {
	ip := net.ParseIP(p.IP)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", p.IP)
	}

	var address string

	if ip.To4() != nil {
		address = fmt.Sprintf("%v:%v", p.IP, p.Port)
	} else {
		address = fmt.Sprintf("[%v]:%v", p.IP, p.Port)
	}
	return address, nil
}
