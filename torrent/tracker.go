package torrent

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"swiftpeer/client/torrent/bencode"
	"swiftpeer/client/torrent/parser"
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
	IP   string
	Port int
}

// ip and event are optional, but that might change
func constructURL(peerID [20]byte, port int) (string, error) {
	m, err := parser.ParseMetadata()
	if err != nil {
		return "", err

	}
	domain, err := url.Parse(m.Announce)
	if err != nil {
		return "", err
	}
	t, err := toTorrent(m)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{fmt.Sprintf("%d", port)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":       []string{fmt.Sprintf("%d", t.Length)},
	}
	domain.RawQuery = params.Encode()
	fmt.Println("\nURL: " + domain.String() + "\n")
	return domain.String(), nil
}

func AnnounceTracker(peerID [20]byte, port int) (string, error) {
	u, err := constructURL(peerID, port)
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
func ParseTrackerResponse() (TrackerResponse, error) {
	response, err := AnnounceTracker([20]byte([]byte("abcdefghijabcdefghij")), 6889)
	if err != nil {
		return TrackerResponse{}, err
	}

	decoded_response, err := bencode.NewDecoder(bufio.NewReader(strings.NewReader(response))).Decode()
	if err != nil {
		return TrackerResponse{}, err
	}

	var trackerResponse TrackerResponse

	switch data := decoded_response.(type) {
	case map[string]interface{}:
		if complete, ok := data["complete"].(int); ok {
			trackerResponse.Complete = complete
		} else {
			return TrackerResponse{}, fmt.Errorf("invalid complete field")
		}

		if incomplete, ok := data["incomplete"].(int); ok {
			trackerResponse.Incomplete = incomplete
		} else {
			return TrackerResponse{}, fmt.Errorf("invalid incomplete field")
		}
		if interval, ok := data["interval"].(int); ok {
			trackerResponse.Interval = interval
		} else {
			return TrackerResponse{}, fmt.Errorf("invalid interval field")
		}
		if peersList, ok := data["peers"].([]interface{}); ok {
			var peers []Peer
			for _, peer := range peersList {
				peers = append(peers, Peer{
					IP:   peer.(map[string]any)["ip"].(string),
					Port: peer.(map[string]any)["port"].(int),
				})
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
