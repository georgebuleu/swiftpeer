package tracker

import (
	"fmt"
	"net/url"
	"swiftpeer/client/peer"
)

func GetTorrentData(trackerUrl string, announceList [][]string, port int, infoHash [20]byte) (interface{}, error) {

	url, err := url.Parse(trackerUrl)
	if err != nil {
		return []peer.Peer{}, err
	}
	var urls []string
	if len(announceList) > 0 {
		// use the announce-list if it's not empty
		for _, tier := range announceList {
			urls = append(urls, tier[0])
		}
	} else {
		urls = append(urls, trackerUrl)
	}

	for _, trackerUrl := range urls {
		u, err := url.Parse(trackerUrl)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		switch u.Scheme {
		case "udp":
			// handle udp tracker
			fmt.Printf("url: %v\n", u)
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
				fmt.Println(err.Error())
				continue
			}
			return res, nil
		case "http", "https":
			//handle http tracker
			return requestPeers(u.String(), infoHash, port)

		default:
			return []peer.Peer{}, fmt.Errorf("invalid tracker url scheme")
		}
	}
	return []peer.Peer{}, fmt.Errorf("unable to connect to any tracker")
}
