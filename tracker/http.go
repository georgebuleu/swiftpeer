package tracker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
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

	fmt.Println(string(response))

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
	var originalResp OriginalResponse
	if err := decoder.Decode(&originalResp); err != nil {
		return nil, fmt.Errorf("failed to decode original response: %w", err)
	}

	return t.unmarshalPeersFromBytes(originalResp.Peers)

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
	var compactResponse CompactResponse
	if err := decoder.Decode(&compactResponse); err != nil {
		return nil, fmt.Errorf("failed to decode compact response: %w", err)
	}
	return t.unmarshalPeersFromBytes(compactResponse.Peers)
}

func (t *HTTPTracker) unmarshalPeersFromBytes(data []byte) ([]peer.Peer, error) {
	var peers []peer.Peer
	const peerSize = 6 // 4 bytes for IP, 2 for Port
	if len(data)%peerSize != 0 {
		return nil, fmt.Errorf("malformed http tracker response, len(compactResponse) does not divide to 6 = 0")
	}

	for i := 0; i < len(data); i += peerSize {
		addr := net.TCPAddr{IP: data[i : i+4], Port: int(binary.BigEndian.Uint16(data[i+4 : i+6]))}

		peers = append(peers, peer.Peer{IP: addr.IP.String(), Port: addr.Port})
	}

	return peers, nil

}
