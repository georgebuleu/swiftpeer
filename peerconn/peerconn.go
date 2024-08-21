package peerconn

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"swiftpeer/client/bitfield"
	"swiftpeer/client/handshake"
	"swiftpeer/client/message"
	"time"
)

const (
	writeTimeout = 3 * time.Second
	readTimeout  = 30 * time.Second
	dialTimeout  = 5 * time.Second
	maxTimeouts  = 8
)

type PeerConn struct {
	conn           net.Conn
	reader         *bufio.Reader
	writer         *bufio.Writer
	Choked         bool
	Interested     bool
	PeerChoked     bool
	PeerInterested bool
	Bitfield       bitfield.Bitfield
	Addr           string
	PeerID         [20]byte //remote peer id
	timeoutCount   int
}

func NewPeerConn(addr string, infoHash [20]byte, clientID [20]byte) (*PeerConn, error) {
	conn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer %s: %w", addr, err)
	}

	p := &PeerConn{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
		Choked: true,
		Addr:   addr,
	}

	err = p.performHandshake(infoHash, clientID)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("handshake failed with peer %s: %w", addr, err)
	}

	return p, nil
}

func (pc *PeerConn) performHandshake(infoHash [20]byte, ourPeerID [20]byte) error {
	hs := handshake.NewHandshake(infoHash, ourPeerID)
	err := pc.writeMessage(hs.Serialize())
	if err != nil {
		return fmt.Errorf("failed to send handshake: %w", err)
	}

	pc.conn.SetReadDeadline(time.Now().Add(readTimeout))
	resp, err := hs.Deserialize(pc.reader)
	if err != nil {
		return fmt.Errorf("failed to receive handshake: %w", err)
	}

	if resp.InfoHash != infoHash {
		return fmt.Errorf("infohash mismatch")
	}

	pc.PeerID = resp.PeerId
	return nil
}

func (pc *PeerConn) ReadMessage() (*message.Message, error) {
	pc.conn.SetReadDeadline(time.Now().Add(readTimeout))
	msg, err := message.Read(pc.reader)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			pc.timeoutCount++
			if pc.timeoutCount > maxTimeouts {
				return nil, fmt.Errorf("too many consecutive timeouts")
			}
			return nil, err
		}
		// Reset timeout count for non-timeout errors
		pc.timeoutCount = 0
		return nil, err
	}
	// Reset timeout count due to successful read
	pc.timeoutCount = 0
	return msg, nil
}

func (pc *PeerConn) writeMessage(data []byte) error {
	pc.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	_, err := pc.writer.Write(data)
	if err != nil {
		return err
	}
	return pc.writer.Flush()
}

func (pc *PeerConn) SendRequest(index, begin, length int) error {
	req, err := message.NewRequest(index, begin, length)
	if err != nil {
		return err
	}
	return pc.writeMessage(req.Serialize())
}

func (pc *PeerConn) SendInterested() error {
	msg := message.NewInterested()
	err := pc.writeMessage(msg.Serialize())
	if err != nil {
		return err
	}
	pc.Interested = true
	return nil
}

func (pc *PeerConn) SendNotInterested() error {
	msg := message.NewNotInterested()
	err := pc.writeMessage(msg.Serialize())
	if err != nil {
		return err
	}
	pc.Interested = false
	return nil
}

func (pc *PeerConn) SendHave(index int) error {
	msg := message.NewHave(index)
	return pc.writeMessage(msg.Serialize())
}

func (pc *PeerConn) SendBitfield(bitfield []byte) error {
	msg := message.NewBitfield(bitfield)
	return pc.writeMessage(msg.Serialize())
}

func (pc *PeerConn) SendKeepAlive() error {
	return pc.writeMessage(message.NewKeepAlive().Serialize())
}

func (pc *PeerConn) Close() error {
	return pc.conn.Close()
}
