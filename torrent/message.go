package torrent

import "encoding/binary"

type messageId byte

const (
	ChokeMsg = iota
	UnchokeMsg
	InterestedMsg
	NotInterestedMsg
	HaveMsg
	BitfieldMsg
	RequestMsg
	PieceMsg
	CancelMsg
	PortMsg // only for DHT
)

type Message struct {
	Id      messageId
	Payload []byte
}

func (m *Message) Serialize() []byte {

	//4 bytes for length + 1 for id
	//not all messages have a payload
	msgBuff := make([]byte, 5+len(m.Payload))
	length := make([]byte, 4)
	length = binary.BigEndian.AppendUint32(length, uint32(len(m.Payload)))
	copy(msgBuff[:4], length[:])
	msgBuff[4] = byte(m.Id)
	copy(msgBuff[5:], m.Payload[:])

	return msgBuff
}
