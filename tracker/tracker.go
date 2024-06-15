package tracker

import (
	"fmt"
	"net"
	"net/url"
)

type Peer struct {
	IP     string
	Port   int
	PeerId string
}

func GetPeers(trackerUrl string, port int, infoHash, clientId [20]byte) ([]Peer, error) {
	u, err := url.Parse(trackerUrl)
	if err != nil {
		return []Peer{}, err
	}

	switch u.Scheme {
	case "udp":
		// handle udp tracker
	case "http", "https":
		//handle http tracker
	default:
		return []Peer{}, fmt.Errorf("invalid tracker url scheme")
	}
	return []Peer{}, fmt.Errorf("invalid tracker url scheme")
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
