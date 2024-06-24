package tracker

import (
	"fmt"
	"net/url"
	"swiftpeer/client/peer"
)

func GetTorrentData(announceUrl string, announceList [][]string, port int, infoHash [20]byte, peerAddrs peer.AddrSet) error {

	var urls []string
	if len(announceList) > 0 {
		c := 0
		// use the announce-list if it's not empty
		for _, tier := range announceList {
			for _, u := range tier {
				urls = append(urls, u)
				c++
			}
		}
		fmt.Printf("number of trackers: %v\n", c)
	} else {
		urls = append(urls, announceUrl)
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
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			for _, p := range res.Peers {
				address, err := p.FormatAddress()
				if err != nil {
					fmt.Println("Error formatting address:", err)
					continue
				}
				peerAddrs[address] = struct{}{}
				fmt.Println("Peer Address:", address)
			}
			fmt.Println("Interval:", res.Interval)

			return nil
		case "http", "https":
			//handle http tracker
			res, err := requestPeers(u.String(), infoHash, port)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			if r, ok := res.(CompactResponse); ok {
				for _, p := range r.Peers {
					address, err := p.FormatAddress()
					if err != nil {
						fmt.Println("Error formatting address:", err)
						continue
					}
					peerAddrs[address] = struct{}{}
				}
			} else if r, ok := res.(OriginalResponse); ok {
				for _, p := range r.Peers {
					address, err := p.FormatAddress()
					if err != nil {
						fmt.Println("Error formatting address:", err)
						continue
					}
					peerAddrs[address] = struct{}{}
				}
			}

		default:
			return fmt.Errorf("invalid tracker url scheme")
		}
	}
	return fmt.Errorf("unable to connect to any tracker")
}
