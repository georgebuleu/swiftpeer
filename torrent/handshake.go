package torrent

import (
	"fmt"
	"io"
	"net"
	"sync"
)

//type Handshake struct {
//	Pstr      string
//	HashInfo [20]byte
//	Reserved  byte
//	Client_Id   [20]byte
//}

const (
	pstr         = "BitTorrent protocol"
	ClientId     = "-SP1011-IJLasf24lqrI"
	handshakeLen = 49 + len(pstr)
)

type Handshake struct {
	PeerId   [20]byte
	Pstr     string
	InfoHash [20]byte
}

func NewHandshake(peerId, infoHash [20]byte) *Handshake {
	return &Handshake{
		PeerId:   peerId,
		Pstr:     pstr,
		InfoHash: infoHash,
	}
}

func (h *Handshake) serialize() []byte {

	buff := make([]byte, handshakeLen)
	buff[0] = byte(len(pstr))
	idx := 1
	idx += copy(buff[idx:], h.Pstr)
	idx += copy(buff[idx:], make([]byte, 8)) //8 reserved bytes
	idx += copy(buff[idx:], h.InfoHash[:])
	idx += copy(buff[idx:], h.PeerId[:])
	return buff
}

func (h *Handshake) deserialize(r io.Reader) (*Handshake, error) {
	headerByte := make([]byte, 1)
	_, err := io.ReadFull(r, headerByte)
	if err != nil {
		return nil, fmt.Errorf("failed to read handshake headerByte: %v", err)
	}
	pstrLen := int(headerByte[0])

	if pstrLen == 0 {
		return nil, fmt.Errorf("invalid header: handshake length cannot be zero")
	}

	bodyBuff := make([]byte, pstrLen)
	_, err = io.ReadFull(r, bodyBuff)
	if err != nil {
		return nil, fmt.Errorf("failed to read handshake body: %v\", err")
	}

}

func GetPeerIdAsBytes(peerId string) []byte {
	id := make([]byte, 20)
	_ = copy(id, peerId[:])
	return id
}
