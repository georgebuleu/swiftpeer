package tracker

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"swiftpeer/client/bencode"
	"swiftpeer/client/common"
	"swiftpeer/client/torrent"
)

type TrackerResponse struct {
	FailureReason  string
	WarningMessage string
	Interval       int
	MinInterval    int
	TrackerID      string
	Complete       int
	Incomplete     int
	Peers          []Peer
}

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

// ip and event are optional, but that might change
func constructURL(t *torrent.Torrent, port int) (string, error) {
	if t == nil {
		return "", fmt.Errorf("Tracker: failed to create a new torrent")
	}
	domain, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{common.PeerId[:]},
		"port":       []string{fmt.Sprintf("%d", port)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":       []string{fmt.Sprintf("%d", t.Length)},
	}
	domain.RawQuery = params.Encode()
	fmt.Println("\nURL: " + domain.String() + "\n")
	return domain.String(), nil
}

func AnnounceTracker(t *torrent.Torrent, port int) (string, error) {
	u, err := constructURL(t, port)
	if err != nil {
		return u, err
	}
	resp, err := http.Get(u)
	if err != nil {
		fmt.Println("error during get request: " + err.Error())
		return "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// TO DO: peer id not consumed
func ParseTrackerResponse(t *torrent.Torrent) (TrackerResponse, error) {
	response, err := AnnounceTracker(t, 6889)
	if err != nil {
		return TrackerResponse{}, err
	}
	fmt.Println(response)
	decodedResponse, err := bencode.NewDecoder(bufio.NewReader(strings.NewReader(response))).Decode()
	if err != nil {
		return TrackerResponse{}, err
	}

	var trackerResponse TrackerResponse

	switch data := decodedResponse.(type) {
	case map[string]interface{}:
		//complete is optional
		if complete, ok := data["complete"].(int); ok {
			trackerResponse.Complete = complete
		} else {
			trackerResponse.Complete = 0
		}
		//incomplete is optional
		if incomplete, ok := data["incomplete"].(int); ok {
			trackerResponse.Incomplete = incomplete
		} else {
			trackerResponse.Incomplete = 0
		}
		if interval, ok := data["interval"].(int); ok {
			trackerResponse.Interval = interval
		} else {
			return TrackerResponse{}, fmt.Errorf("invalid interval field")
		}
		if peersList, ok := data["peers"].([]interface{}); ok {
			var peers []Peer
			for _, peer := range peersList {
				if _, ok := peer.(map[string]any)["peer_id"].(string); !ok {
					peers = append(peers, Peer{
						IP:     peer.(map[string]any)["ip"].(string),
						Port:   peer.(map[string]any)["port"].(int),
						PeerId: "",
					})
				} else {
					peers = append(peers, Peer{
						IP:     peer.(map[string]any)["ip"].(string),
						Port:   peer.(map[string]any)["port"].(int),
						PeerId: peer.(map[string]any)["peer_id"].(string),
					})
				}

			}
			trackerResponse.Peers = peers
		} else {
			return TrackerResponse{}, fmt.Errorf("invalid peer field")
		}

	default:
		return TrackerResponse{}, fmt.Errorf("invalid format")
	}
	return trackerResponse, err
}
