package tracker

import (
	"fmt"
	"net/url"
	"swiftpeer/client/peer"
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
	Peers          []peer.Peer
}

type CompactResponse struct {
	Interval int
	Peers    []peer.Peer
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
		fmt.Printf("number of trackers: %v\n", len(urls))
	} else {
		urls = append(urls, announceUrl)
	}

	for _, trackerUrl := range urls {
		tracker, err := NewTracker(trackerUrl)
		if err != nil {
			fmt.Printf("Error creating tracker for %s: %v\n", trackerUrl, err)
			continue
		}

		peers, err := tracker.Announce(infoHash, peerId, port)
		if err != nil {
			fmt.Printf("Error announcing to %s: %v\n", trackerUrl, err)
			continue
		}

		for _, p := range peers {
			address, err := p.FormatAddress()
			if err != nil {
				fmt.Printf("Error formatting address for peer from %s: %v\n", trackerUrl, err)
				continue
			}
			peerAddrs[address] = struct{}{}
		}

		fmt.Printf("Successfully announced to %s and received %d peers\n", trackerUrl, len(peers))
		return nil // Successfully connected to a tracker
	}

	return fmt.Errorf("unable to connect to any tracker")
}
