package handshake

import (
	"fmt"
	"io"
)

const (
	pstr         = "BitTorrent protocol"
	ClientId     = "-SP1011-IJLasf24lqrI"
	handshakeLen = 49 + len(pstr) // hash_info + peer_id + 1(header byte for the length)
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

func (h *Handshake) Serialize() []byte {

	buff := make([]byte, handshakeLen)
	buff[0] = byte(len(pstr))
	idx := 1
	idx += copy(buff[idx:], h.Pstr)
	idx += copy(buff[idx:], make([]byte, 8)) //8 reserved bytes
	idx += copy(buff[idx:], h.InfoHash[:])
	idx += copy(buff[idx:], h.PeerId[:])
	return buff
}

func (h *Handshake) Deserialize(r io.Reader) (*Handshake, error) {
	headerByte := make([]byte, 1)
	_, err := io.ReadFull(r, headerByte)
	if err != nil {
		return nil, fmt.Errorf("failed to read handshake headerByte: %v", err)
	}
	pstrLen := int(headerByte[0])

	if pstrLen == 0 {
		return nil, fmt.Errorf("invalid header: handshake length cannot be zero")
	}

	bodyBuff := make([]byte, pstrLen+48) //handshake size (in bytes) - 1 byte from the header, already consumed
	_, err = io.ReadFull(r, bodyBuff)
	if err != nil {
		return nil, fmt.Errorf("failed to read handshake body: %v\", err")
	}

	//reserved := bodyBuff[pstrLen : pstrLen+8]
	var infoHash, peerId [20]byte
	copy(infoHash[:], bodyBuff[pstrLen+8:pstrLen+28])
	copy(peerId[:], bodyBuff[pstrLen+28:])

	return &Handshake{
		Pstr:     string(bodyBuff[0:pstrLen]),
		InfoHash: infoHash,
		PeerId:   peerId,
	}, nil
}
