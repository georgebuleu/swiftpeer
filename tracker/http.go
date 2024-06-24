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
		return nil, err
	}

	data, ok := decodedResponse.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid format")
	}

	if _, ok := data["peers"].(string); ok {
		// Compact format
		return parseCompactFormat(response)
	}

	if _, ok := data["peers"].([]interface{}); ok {
		// Original format
		return parseOriginalFormat(data)
	}

	return nil, fmt.Errorf("invalid peers field format")
}

func parseOriginalFormat(data map[string]interface{}) (OriginalResponse, error) {
	var trackerResponse OriginalResponse

	if interval, ok := data["interval"].(int); ok {
		trackerResponse.Interval = interval
	} else {
		return trackerResponse, fmt.Errorf("invalid interval field")
	}

	if complete, ok := data["complete"].(int); ok {
		trackerResponse.Complete = complete
	}

	if incomplete, ok := data["incomplete"].(int); ok {
		trackerResponse.Incomplete = incomplete
	}

	if peersList, ok := data["peers"].([]interface{}); ok {
		var peers []peer.Peer
		for _, p := range peersList {
			peerMap, ok := p.(map[string]interface{})
			if !ok {
				return trackerResponse, fmt.Errorf("invalid peer entry")
			}

			ip, ipOk := peerMap["ip"].(string)
			port, portOk := peerMap["port"].(int)
			peerID, peerIDOk := peerMap["peer_id"].(string)

			if !ipOk || !portOk {
				return trackerResponse, fmt.Errorf("invalid peer entry fields")
			}

			if !peerIDOk {
				peerID = ""
			}

			peers = append(peers, peer.Peer{
				IP:     ip,
				Port:   port,
				PeerId: peerID,
			})
		}
		trackerResponse.Peers = peers
	} else {
		return trackerResponse, fmt.Errorf("invalid peers field")
	}

	return trackerResponse, nil
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
