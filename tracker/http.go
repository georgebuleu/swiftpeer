package tracker

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"swiftpeer/client/bencode"
	"swiftpeer/client/common"
	"swiftpeer/client/peer"
)

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

// it can return OriginalResponse type(original response type) or compactResponse
func requestPeers(url string, infoHash [20]byte, port int) (interface{}, error) {
	u, err := constructURL(url, infoHash, port)
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
	return parseTrackerResponse(string(b))
}

func constructURL(trackerUrl string, infoHash [20]byte, port int) (string, error) {

	domain, err := url.Parse(trackerUrl)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(infoHash[:])},
		"peer_id":    []string{common.PeerId[:]},
		"port":       []string{fmt.Sprintf("%d", port)},
		"compact":    []string{"1"},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":       []string{"0"},
	}
	domain.RawQuery = params.Encode()
	fmt.Println("\nURL: " + domain.String() + "\n")
	return domain.String(), nil
}

// it can return OriginalResponse type(original response type) or compactResponse
func parseTrackerResponse(response string) (interface{}, error) {

	fmt.Println(response)
	decodedResponse, err := bencode.NewDecoder(bufio.NewReader(strings.NewReader(response))).Decode()
	if err != nil {
		return OriginalResponse{}, err
	}
	var trackerResponse OriginalResponse

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
			return OriginalResponse{}, fmt.Errorf("invalid interval field")
		}

		//check format and parse based on it

		switch peers := data["peers"].(type) {
		case string:
			res, err := parseCompactFormat(peers)
			if err != nil {
				return CompactResponse{}, err
			}
			return res, nil

		case []interface{}:
			if peersList, ok := data["peers"].([]interface{}); ok {
				var peers []peer.Peer
				for _, p := range peersList {
					if _, ok := p.(map[string]any)["peer_id"].(string); !ok {
						peers = append(peers, peer.Peer{
							IP:     p.(map[string]any)["ip"].(string),
							Port:   p.(map[string]any)["port"].(int),
							PeerId: "",
						})
					} else {
						peers = append(peers, peer.Peer{
							IP:     p.(map[string]any)["ip"].(string),
							Port:   p.(map[string]any)["port"].(int),
							PeerId: p.(map[string]any)["peer_id"].(string),
						})
					}

				}
				trackerResponse.Peers = peers
			} else {
				return OriginalResponse{}, fmt.Errorf("invalid peer field")
			}
		}

	default:
		return OriginalResponse{}, fmt.Errorf("invalid format")
	}
	return trackerResponse, err
}

func parseCompactFormat(response string) (CompactResponse, error) {

	decodedRes, err := bencode.NewDecoder(bufio.NewReader(strings.NewReader(response))).Decode()
	if err != nil {
		return CompactResponse{}, err
	}

	resDict, ok := decodedRes.(map[string]interface{})
	if !ok {
		return CompactResponse{}, fmt.Errorf("invalid compact response format")
	}

	interval, ok := resDict["interval"].(int)
	if !ok {
		return CompactResponse{}, fmt.Errorf("invalid or missing interval field in compact response")
	}
	compactPeers, ok := resDict["peers"].(string)
	if !ok {
		return CompactResponse{}, fmt.Errorf("invalid or missing peers field in compact response")
	}

	peers, err := parseCompactPeers(compactPeers)

	return CompactResponse{
		Peers:    peers,
		Interval: interval,
	}, nil
}

func parseCompactPeers(compactPeers string) ([]peer.Peer, error) {
	if len(compactPeers)%6 != 0 {
		return nil, fmt.Errorf("invalid compact peers format")
	}
	var peers []peer.Peer

	for i := 0; i < len(compactPeers); i += 6 {
		ip := fmt.Sprintf("%d.%d.%d.%d", compactPeers[i], compactPeers[i+1], compactPeers[i+2], compactPeers[i+3])
		port := int(compactPeers[4])<<8 + int(compactPeers[5])
		peers = append(peers, peer.Peer{IP: ip, Port: port})
	}

	return peers, nil
}
