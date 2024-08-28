package tracker

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/url"
	"swiftpeer/client/peer"
	"time"
)

const (
	connectAction  = 0
	announceAction = 1

	protocolID           = 0x41727101980
	connPacketSize       = 16
	announceReqSize      = 98
	maxRetries           = 3
	connIDExpirationTime = 60 * time.Second
)

type UdpTracker struct {
	url      *url.URL
	conn     *net.UDPConn
	connID   uint64
	connTime time.Time
}

func NewUdpTracker(trackerURL string) (Tracker, error) {
	u, err := url.Parse(trackerURL)
	if err != nil {
		return nil, fmt.Errorf("invalid tracker URL: %w", err)
	}
	return &UdpTracker{url: u}, nil
}

func (t *UdpTracker) Announce(infoHash [20]byte, peerID [20]byte, port int) ([]peer.Peer, error) {
	if err := t.ensureConnection(); err != nil {
		return nil, fmt.Errorf("failed to establish connection: %w", err)
	}

	response, err := t.sendAnnounceRequest(infoHash, peerID, port)
	if err != nil {
		return nil, fmt.Errorf("failed to send announce request: %w", err)
	}

	return response.Peers, nil
}

func (t *UdpTracker) ensureConnection() error {
	if t.conn == nil || time.Since(t.connTime) > connIDExpirationTime {
		if err := t.connect(); err != nil {
			return err
		}
	}
	return nil
}

func (t *UdpTracker) connect() error {
	udpAddr, err := net.ResolveUDPAddr("udp", t.url.Host)
	if err != nil {
		return fmt.Errorf("[ERROR]couldn't resolve UDP address: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return fmt.Errorf("failed to dial UDP: %w", err)
	}

	if err := conn.SetReadBuffer(2048); err != nil {
		conn.Close()
		return fmt.Errorf("failed to set UDP read buffer: %w", err)
	}

	t.conn = conn

	transactionID := uint32(time.Now().UnixNano())
	req := t.buildConnectRequest(transactionID)

	for attempt := 0; attempt <= maxRetries; attempt++ {
		t.conn.SetDeadline(time.Now().Add(2 * time.Second * (1 << attempt)))
		fmt.Printf("[INFO] attempt %v to connect %v failed, retrying...\n", attempt, udpAddr)
		if err := t.sendAndReceive(req, connPacketSize, func(resp []byte) error {
			return t.handleConnectResponse(resp, transactionID)
		}); err == nil {
			t.connTime = time.Now()
			return nil
		}
	}

	return fmt.Errorf("[INFO]failed to connect after %d attempts", maxRetries+1)
}

func (t *UdpTracker) buildConnectRequest(transactionID uint32) []byte {
	req := make([]byte, connPacketSize)
	binary.BigEndian.PutUint64(req[:8], protocolID)
	binary.BigEndian.PutUint32(req[8:12], connectAction)
	binary.BigEndian.PutUint32(req[12:16], transactionID)
	return req
}

func (t *UdpTracker) handleConnectResponse(resp []byte, transactionID uint32) error {
	if len(resp) != connPacketSize {
		return fmt.Errorf("invalid packet size, expected %v, got %v", connPacketSize, len(resp))
	}

	if action := binary.BigEndian.Uint32(resp[:4]); action != connectAction {
		return fmt.Errorf("invalid action: expected %d, got %d", connectAction, action)
	}

	if respTransactionID := binary.BigEndian.Uint32(resp[4:8]); respTransactionID != transactionID {
		return fmt.Errorf("transaction ID mismatch: expected %d, got %d", transactionID, respTransactionID)
	}

	t.connID = binary.BigEndian.Uint64(resp[8:])
	return nil
}

func (t *UdpTracker) sendAnnounceRequest(infoHash [20]byte, peerID [20]byte, port int) (*UdpResponse, error) {
	transactionID := uint32(time.Now().UnixNano())
	req := t.buildAnnounceRequest(transactionID, infoHash, peerID, port)

	var response UdpResponse
	err := t.sendAndReceive(req, 20, func(resp []byte) error {
		return t.handleAnnounceResponse(resp, transactionID, &response)
	})

	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (t *UdpTracker) buildAnnounceRequest(transactionID uint32, infoHash [20]byte, peerID [20]byte, port int) []byte {
	req := make([]byte, announceReqSize)
	binary.BigEndian.PutUint64(req[:8], t.connID)
	binary.BigEndian.PutUint32(req[8:12], announceAction)
	binary.BigEndian.PutUint32(req[12:16], transactionID)
	copy(req[16:36], infoHash[:])
	copy(req[36:56], peerID[:])
	binary.BigEndian.PutUint64(req[56:64], 0)          // downloaded
	binary.BigEndian.PutUint64(req[64:72], 0)          // left
	binary.BigEndian.PutUint64(req[72:80], 0)          // uploaded
	binary.BigEndian.PutUint32(req[80:84], 0)          // event
	binary.BigEndian.PutUint32(req[84:88], 0)          // IP address
	binary.BigEndian.PutUint32(req[88:92], 0)          // key
	binary.BigEndian.PutUint32(req[92:96], 0xFFFFFFFF) // num_want
	binary.BigEndian.PutUint16(req[96:98], uint16(port))
	return req
}

func (t *UdpTracker) handleAnnounceResponse(resp []byte, transactionID uint32, response *UdpResponse) error {
	if len(resp) < 20 {
		return fmt.Errorf("response too short: %d bytes", len(resp))
	}

	if action := binary.BigEndian.Uint32(resp[0:4]); action != announceAction {
		return fmt.Errorf("invalid action: expected %d, got %d", announceAction, action)
	}

	if respTransactionID := binary.BigEndian.Uint32(resp[4:8]); respTransactionID != transactionID {
		return fmt.Errorf("transaction ID mismatch: expected %d, got %d", transactionID, respTransactionID)
	}

	response.Interval = int(binary.BigEndian.Uint32(resp[8:12]))
	response.Leechers = int(binary.BigEndian.Uint32(resp[12:16]))

	response.Peers = t.parsePeers(resp[20:])
	return nil
}

func (t *UdpTracker) parsePeers(data []byte) []peer.Peer {
	peerCount := len(data) / 6
	peers := make([]peer.Peer, 0, peerCount)

	for i := 0; i < peerCount; i++ {
		offset := i * 6
		peers = append(peers, peer.Peer{
			IP:   fmt.Sprintf("%d.%d.%d.%d", data[offset], data[offset+1], data[offset+2], data[offset+3]),
			Port: int(binary.BigEndian.Uint16(data[offset+4 : offset+6])),
		})
	}

	return peers
}

func (t *UdpTracker) sendAndReceive(req []byte, minRespSize int, handler func([]byte) error) error {
	if _, err := t.conn.Write(req); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	resp := make([]byte, 2048)
	n, err := t.conn.Read(resp)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if n < minRespSize {
		return fmt.Errorf("response too short: got %d bytes, expected at least %d", n, minRespSize)
	}

	return handler(resp[:n])
}

func (t *UdpTracker) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}
