package tracker

import (
	"fmt"
	"net/url"
	"swiftpeer/client/peer"
	"sync"
)

type Tracker interface {
	Announce(infoHash [20]byte, peerId [20]byte, port int) ([]peer.Peer, error)
}

type OriginalResponse struct {
	FailureReason  string
	WarningMessage string
	Interval       int
	MinInterval    int
	TrackerID      string
	Complete       int
	Incomplete     int
	Event          string
	Peers          []byte
}

type CompactResponse struct {
	Interval int
	Peers    []byte
}

type UdpResponse struct {
	Interval int
	Leechers int
	Peers    []peer.Peer
}

func NewTracker(trackeUrl string) (Tracker, error) {
	parsedURL, err := url.Parse(trackeUrl)
	if err != nil {
		return nil, err
	}

	switch parsedURL.Scheme {
	case "http", "https":
		return NewHTTPTracker(trackeUrl), nil
	case "udp":
		return NewUdpTracker(trackeUrl)
	default:
		return nil, fmt.Errorf("unsupported tracker scheme: %s", parsedURL.Scheme)
	}
}

func GetTorrentData(announceUrl string, announceList [][]string, port int, infoHash [20]byte, peerId [20]byte, peerAddrs peer.AddrSet) error {
	var urls []string
	if len(announceList) > 0 {
		for _, tier := range announceList {
			urls = append(urls, tier...)
		}
		fmt.Printf("[INFO] number of trackers: %v\n", len(urls))
	} else {
		urls = append(urls, announceUrl)
	}

	var wg sync.WaitGroup
	peerChan := make(chan []peer.Peer, len(urls))

	for _, trackerUrl := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			tracker, err := NewTracker(url)
			if err != nil {
				fmt.Printf("[INFO] failed to create tracker instance for %s: %v\n", url, err)
				return
			}

			peers, err := tracker.Announce(infoHash, peerId, port)
			if err != nil {
				fmt.Printf("[INFO]Error announcing to %s: %v\n", url, err)
				return
			}

			peerChan <- peers
			fmt.Printf("[INFO] Successfully announced to %s and received %d peers\n", url, len(peers))
		}(trackerUrl)
	}

	go func() {
		wg.Wait()
		close(peerChan)
	}()

	for peers := range peerChan {
		for _, p := range peers {
			address, err := p.FormatAddress()
			if err != nil {
				fmt.Printf("[ERROR] error formatting address for peer: %v\n", err)
				continue
			}
			peerAddrs[address] = struct{}{}
		}
	}

	if len(peerAddrs) == 0 {
		return fmt.Errorf("unable to get peers from any tracker")
	}

	fmt.Printf("[INFO] total peers: %d\n", len(peerAddrs))
	return nil
}
