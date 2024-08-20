package peerconn

import (
	"fmt"
	"io"
	"net"
	"swiftpeer/client/bitfield"
	"swiftpeer/client/common"
	"swiftpeer/client/handshake"
	"swiftpeer/client/message"
	"time"
)

type PeerConn struct {
	Conn     net.Conn
	Addr     string
	InfoHash [20]byte
	IsChoked bool
	Pieces   bitfield.Bitfield
}

func NewPeerConn(addr string, infoHash [20]byte) (*PeerConn, error) {
	//address, err := addr.FormatAddress()
	//if err != nil {
	//	return nil, err
	//}
	conn, err := net.DialTimeout("tcp", addr, time.Second*3)

	if err != nil {

		return nil, fmt.Errorf("Failed to connect to  %v. %v\n", addr, err.Error())
	}

	pc := &PeerConn{
		Conn:     conn,
		Addr:     addr,
		InfoHash: infoHash,
		IsChoked: true,
		Pieces:   bitfield.Bitfield{},
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
	hs := handshake.NewHandshake(common.GeneratePeerId(), pc.InfoHash)
	pc.Conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer pc.Conn.SetDeadline(time.Time{})
	_, err := pc.Conn.Write(hs.Serialize())
	if err != nil {
		return fmt.Errorf("failed to send handshake with %v : %v", pc.Conn.RemoteAddr(), err)
	}
	response, err := hs.Deserialize(io.Reader(pc.Conn))
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
	fmt.Printf("Successfuly connected to: %v\n", pc.Conn.LocalAddr())
	return nil
}

func (pc *PeerConn) receiveBitfield() error {
	pc.Conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer pc.Conn.SetDeadline(time.Time{})

	msg, err := message.Read(pc.Conn)
	if err != nil {
		return err
	}

	if msg.Id != message.BitfieldMsg {
		return fmt.Errorf("expected bitdfield msg but got: %v", msg.Name())
	}

	pc.Pieces = msg.Payload
	return nil
}

func (pc *PeerConn) SendRequestMsg(pieceIndex, offset, length int) error {

	m, err := message.NewRequest(pieceIndex, offset, length)
	if err != nil {
		return err
	}
	_, err = pc.Conn.Write(m.Serialize())
	return err
}

func (pc *PeerConn) SendInterested() error {
	_, err := pc.Conn.Write(message.NewInterested().Serialize())
	return err
}

func (pc *PeerConn) SendNotInterested() error {
	_, err := pc.Conn.Write(message.NewNotInterested().Serialize())
	return err
}

func (pc *PeerConn) SendUnchoke() error {
	_, err := pc.Conn.Write(message.NewUnchoke().Serialize())
	return err
}

func (pc *PeerConn) SendHave(index int) error {
	_, err := pc.Conn.Write(message.NewHave(index).Serialize())
	return err
}

func (pc *PeerConn) Read() (*message.Message, error) {
	if pc == nil {
		return nil, fmt.Errorf("error:connection closed")
	}
	return message.Read(pc.Conn)
}
