package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

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
	//empty message means keep-alive
	if m == nil {
		return make([]byte, 4)
	}
	msgBuff := make([]byte, 5+len(m.Payload))
	length := make([]byte, 4)
	length = binary.BigEndian.AppendUint32(length, uint32(len(m.Payload)))
	copy(msgBuff[:4], length[:])
	msgBuff[4] = byte(m.Id)
	copy(msgBuff[5:], m.Payload[:])

	return msgBuff
}

// <length prefix><message ID><payload>
// length prefix is a four byte big-endian value
// message ID is a single decimal byte
// payload is message dependent.
func Read(r io.Reader) *Message {
	lengthBuff := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuff)
	if err != nil {
		fmt.Printf("Error while reading the prefix length: %v", err.Error())
		return nil
	}

	length := binary.BigEndian.Uint32(lengthBuff)
	message := make([]byte, length)
	_, err = io.ReadFull(r, message)
	if err != nil {
		fmt.Printf("Error while reading the message: %v", err.Error())
		return nil
	}

	return &Message{
		Id:      messageId(message[0]),
		Payload: message[1:],
	}
}

func (m *Message) KeepAliveMsg() []byte {
	msg := make([]byte, 4)
	binary.BigEndian.PutUint32(msg, 0)
	return msg
}

func (m *Message) GetMessageName() string {
	switch m.Id {
	case ChokeMsg:
		return "ChokeMsg"
	case UnchokeMsg:
		return "UnchokeMsg"
	case InterestedMsg:
		return "InterestedMsg"
	case NotInterestedMsg:
		return "NotInterestedMsg"
	case HaveMsg:
		return "HaveMsg"
	case BitfieldMsg:
		return "BitfieldMsg"
	case RequestMsg:
		return "RequestMsg"
	case PieceMsg:
		return "PieceMsg"
	case CancelMsg:
		return "CancelMsg"
	case PortMsg:
		return "PortMsg"
	default:
		return "UnknownMsg"
	}
}

func (m *Message) onBitfieldMsg() {

}
