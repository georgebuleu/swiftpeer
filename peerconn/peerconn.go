package peerconn

import (
	"fmt"
	"io"
	"net"
	"swiftpeer/client/bitfield"
	"swiftpeer/client/handshake"
	"swiftpeer/client/message"
	"swiftpeer/client/torrent"
	"swiftpeer/client/tracker"
	"swiftpeer/client/utils"
	"sync"
	"time"
)

type PeerConn struct {
	conn     net.Conn
	peer     tracker.Peer
	infoHash [20]byte
	isChoked bool
	pieces   bitfield.Bitfield
}

func (pc *PeerConn) initHandshake() error {
	hs := handshake.NewHandshake(utils.GetPeerIdAsBytes(pc.peer.PeerId), pc.infoHash)
	_, err := pc.conn.Write(hs.Serialize())
	if err != nil {
		return fmt.Errorf("failed to send handshake with %v : %v", pc.conn.RemoteAddr(), err)
	}
	answer, err := hs.Deserialize(io.Reader(pc.conn))
	if err != nil {
		if err != io.EOF {
			return fmt.Errorf("Failed to READ: %v\n", err.Error())
		} else {
			return fmt.Errorf("peer closed the connection")
		}
	}
	if hs.InfoHash != answer.InfoHash {

		return fmt.Errorf("different info_hash during handshake")
	}

	return nil
}

func HandlePeers(peers []tracker.Peer, wg *sync.WaitGroup) {
	for _, peer := range peers {
		wg.Add(1)
		go func(p tracker.Peer) {
			_, err := NewPeerConn(p)
			if err != nil {
				fmt.Println(err)
			}
			wg.Done()
		}(peer)
	}
}

func NewPeerConn(peer tracker.Peer) (*PeerConn, error) {
	address, err := peer.FormatAddress()
	if err != nil {
		return nil, err
	}
	hash, _ := torrent.HashInfo()
	conn, err := net.DialTimeout("tcp", address, time.Second*20)

	if err != nil {

		return nil, fmt.Errorf("Failed to connect to  %v. %v\n", address, err.Error())
	}

	pc := &PeerConn{
		conn:     conn,
		peer:     peer,
		infoHash: hash,
		isChoked: true,
		pieces:   bitfield.Bitfield{},
	}

	err = pc.initHandshake()
	if err != nil {
		conn.Close()
		return nil, err
	}
	msg := message.Read(io.Reader(conn))
	fmt.Printf("First Message: %v: %v\n", msg.GetMessageName(), msg.Payload)
	if msg.Id == message.BitfieldMsg {
		pc.pieces = msg.Payload
	}

	return pc, nil
}
