package tracker

import (
	"fmt"
	"net/url"
	"swiftpeer/client/peer"
)

func GetTorrentData(trackerUrl string, port int, infoHash [20]byte) (interface{}, error) {
	u, err := url.Parse(trackerUrl)
	if err != nil {
		return []peer.Peer{}, err
	}

	switch u.Scheme {
	case "udp":
		// handle udp tracker
		res, err := getPeersFromUDPTracker(u, infoHash, port)
		for _, p := range res.Peers {
			address, err := p.FormatAddress()
			if err != nil {
				fmt.Println("Error formatting address:", err)
				continue
			}
			fmt.Println("Peer Address:", address)
		}
		fmt.Println("Interval:", res.Interval)
		if err != nil {
			return nil, err
		}
		return res, nil
	case "http", "https":
		//handle http tracker
		return requestPeers(u.String(), infoHash, port)

	default:
		return []peer.Peer{}, fmt.Errorf("invalid tracker url scheme")
	}
}
