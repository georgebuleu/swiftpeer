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

type HTTPTracker struct {
	baseUrl string
}

func NewHTTPTracker(baseUrl string) *HTTPTracker {
	return &HTTPTracker{baseUrl: baseUrl}
}

// it can return OriginalResponse type(original response type) or compactResponse
func (t *HTTPTracker) Announce(infoHash [20]byte, peerID [20]byte, port int) ([]peer.Peer, error) {
	announceURL, err := t.buildAnnounceURL(infoHash, peerID, port)
	if err != nil {
		return nil, fmt.Errorf("failed to build announce URL: %w", err)
	}

	response, err := t.sendAnnounceRequest(announceURL)
	if err != nil {
		return nil, fmt.Errorf("failed to send announce request: %w", err)
	}

	peers, err := t.extractPeersFromResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to extract peers from response: %w", err)
	}

	return peers, nil
}

func (t *HTTPTracker) buildAnnounceURL(infoHash [20]byte, peerID [20]byte, port int) (string, error) {
	base, err := url.Parse(t.baseUrl)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	params := url.Values{
		"info_hash":  []string{string(infoHash[:])},
		"peer_id":    []string{common.PeerIdToString(peerID)},
		"port":       []string{fmt.Sprintf("%d", port)},
		"compact":    []string{"1"},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":       []string{"0"},
	}
	base.RawQuery = params.Encode()

	return base.String(), nil
}

func (t *HTTPTracker) sendAnnounceRequest(announceURL string) ([]byte, error) {
	resp, err := http.Get(announceURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP GET request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

func (t *HTTPTracker) handleOriginalFormat(decoder *bencode.Decoder) ([]peer.Peer, error) {
	var originalResp struct {
		Peers []struct {
			IP   string `bencode:"ip"`
			Port int    `bencode:"port"`
		} `bencode:"peers"`
	}
	if err := decoder.Decode(&originalResp); err != nil {
		return nil, fmt.Errorf("failed to decode original response: %w", err)
	}

	peers := make([]peer.Peer, len(originalResp.Peers))
	for i, p := range originalResp.Peers {
		peers[i] = peer.Peer{IP: p.IP, Port: p.Port}
	}
	return peers, nil
}

func (t *HTTPTracker) isCompactResponse(response []byte) bool {
	return bytes.Contains(response[:50], []byte(`5:peers`))
}

func (t *HTTPTracker) extractPeersFromResponse(response []byte) ([]peer.Peer, error) {
	decoder := bencode.NewDecoder(bufio.NewReader(bytes.NewReader(response)))

	if t.isCompactResponse(response) {
		return t.handleCompactFormat(decoder)
	}
	return t.handleOriginalFormat(decoder)
}

func (t *HTTPTracker) handleCompactFormat(decoder *bencode.Decoder) ([]peer.Peer, error) {
	var compactResp struct {
		Peers string `bencode:"peers"`
	}
	if err := decoder.Decode(&compactResp); err != nil {
		return nil, fmt.Errorf("failed to decode compact response: %w", err)
	}
	return t.parseCompactPeers(compactResp.Peers)
}

func (t *HTTPTracker) parseCompactPeers(compactPeers string) ([]peer.Peer, error) {
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
