package tracker

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	return parseTrackerResponse(b)
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
func parseTrackerResponse(response []byte) (interface{}, error) {

	fmt.Println(response)
	decoder := bencode.NewDecoder(bytes.NewReader(response))
	r := bufio.NewReader(bytes.NewReader(response))

	start, err := r.Peek(50)

	if err != nil {
		return nil, err
	}

	compact := bytes.Contains(start, []byte(`5:peers`))

	if compact {
		// Compact format
		var resDict map[string]interface{}
		err := decoder.Decode(resDict)

		if err != nil {
			return nil, err
		}

		return parseCompactFormat(resDict)

	}

	if !compact {
		// Original format
		originalResponse := new(OriginalResponse)
		err := decoder.Decode(originalResponse)
		return originalResponse.Peers, err
	}

	return nil, fmt.Errorf("invalid peers field format")
}

func parseCompactFormat(resDict map[string]interface{}) (CompactResponse, error) {

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
	}, err
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
