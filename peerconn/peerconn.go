package peerconn

import (
	"fmt"
	"io"
	"net"
	"swiftpeer/client/bitfield"
	"swiftpeer/client/handshake"
	"swiftpeer/client/message"
	"swiftpeer/client/tracker"
	"swiftpeer/client/utils"
	"time"
)

type PeerConn struct {
	conn     net.Conn
	peer     tracker.Peer
	infoHash [20]byte
	isChoked bool
	pieces   bitfield.Bitfield
}

func NewPeerConn(peer tracker.Peer, infoHash [20]byte) (*PeerConn, error) {
	address, err := peer.FormatAddress()
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout("tcp", address, time.Second*20)

	if err != nil {

		return nil, fmt.Errorf("Failed to connect to  %v. %v\n", address, err.Error())
	}

	pc := &PeerConn{
		conn:     conn,
		peer:     peer,
		infoHash: infoHash,
		isChoked: true,
		pieces:   bitfield.Bitfield{},
	}

	err = pc.doHandshake()
	if err != nil {
		conn.Close()
		return nil, err
	}

	err = pc.receiveBitfield()
	if err != nil {
		conn.Close()
		return nil, err
	}

	return pc, nil
}

func (pc *PeerConn) doHandshake() error {
	hs := handshake.NewHandshake(utils.GetPeerIdAsBytes(pc.peer.PeerId), pc.infoHash)
	_, err := pc.conn.Write(hs.Serialize())
	if err != nil {
		return fmt.Errorf("failed to send handshake with %v : %v", pc.conn.RemoteAddr(), err)
	}
	response, err := hs.Deserialize(io.Reader(pc.conn))
	if err != nil {
		if err != io.EOF {
			return fmt.Errorf("Failed to READ: %v\n", err.Error())
		} else {
			return fmt.Errorf("peer closed the connection")
		}
	}
	if hs.InfoHash != response.InfoHash {

		return fmt.Errorf("different info_hash during handshake")
	}
	fmt.Printf("Successfuly connected to: %v\n", pc.conn.LocalAddr())
	return nil
}

func (pc *PeerConn) receiveBitfield() error {
	pc.conn.SetDeadline(time.Now().Add(8 * time.Second))
	defer pc.conn.SetDeadline(time.Time{})

	msg, err := message.Read(pc.conn)
	if err != nil {
		return err
	}

	if msg.Id != message.BitfieldMsg {
		return fmt.Errorf("expected bitdfield msg but got: %v", msg.Name())
	}

	pc.pieces = msg.Payload
	return nil
}

func (pc *PeerConn) SendRequestMsg(pieceIndex, offset, length int) error {

	m, err := message.NewRequest(pieceIndex, offset, length)
	if err != nil {
		return err
	}
	_, err = pc.conn.Write(m.Serialize())
	return err
}

func (pc *PeerConn) SendInterested() error {
	_, err := pc.conn.Write(message.NewInterested().Serialize())
	return err
}

func (pc *PeerConn) SendNotInterested() error {
	_, err := pc.conn.Write(message.NewNotInterested().Serialize())
	return err
}

func (pc *PeerConn) SendUnchoke() error {
	_, err := pc.conn.Write(message.NewUnchoke().Serialize())
	return err
}

//TODO have message

func (pc *PeerConn) Read() (*message.Message, error) {
	return message.Read(pc.conn)
}
