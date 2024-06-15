package http_tracker

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"swiftpeer/client/bencode"
	"swiftpeer/client/common"
	"swiftpeer/client/torrent"
	"swiftpeer/client/tracker"
)

type TrackerResponse struct {
	FailureReason  string
	WarningMessage string
	Interval       int
	MinInterval    int
	TrackerID      string
	Complete       int
	Incomplete     int
	Peers          []tracker.Peer
}

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

func announceTracker(t *torrent.Torrent, port int) (string, error) {
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

func ParseTrackerResponse(t *torrent.Torrent) (TrackerResponse, error) {
	response, err := announceTracker(t, 6899)
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
			var peers []tracker.Peer
			for _, peer := range peersList {
				if _, ok := peer.(map[string]any)["peer_id"].(string); !ok {
					peers = append(peers, tracker.Peer{
						IP:     peer.(map[string]any)["ip"].(string),
						Port:   peer.(map[string]any)["port"].(int),
						PeerId: "",
					})
				} else {
					peers = append(peers, tracker.Peer{
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
