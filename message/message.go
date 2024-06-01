package message

import (
	"encoding/binary"
	"fmt"
	"io"
	"swiftpeer/client/common"
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
func Read(r io.Reader) (*Message, error) {
	lengthBuff := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuff)
	if err != nil {
		fmt.Printf("Error while reading the prefix length in message: %v\n", err.Error())
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuff)

	if length == 0 {
		return nil, nil
	}

	message := make([]byte, length)
	_, err = io.ReadFull(r, message)
	if err != nil {
		fmt.Printf("Error while reading the message: %v\n", err.Error())
		return nil, err
	}

	m := Message{
		Id:      messageId(message[0]),
		Payload: message[1:],
	}

	//fmt.Printf("received message type: %v\n", m.Name())

	return &m, nil
}

// keep-alive is 0 length, the id is not serialized
func NewKeepAlive() *Message {
	return nil
}

func NewRequest(pieceIndex, offset, length int) (*Message, error) {
	if pieceIndex < 0 || offset < 0 || length <= 0 || length > common.BlockSize {
		return nil, fmt.Errorf("invalid parameters: pieceIndex=%d, offset=%d, length=%d\n", pieceIndex, offset, length)
	}
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(pieceIndex))
	binary.BigEndian.PutUint32(payload[4:8], uint32(offset))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{
		Id:      RequestMsg,
		Payload: payload,
	}, nil
}

func NewInterested() *Message {
	return &Message{
		Id:      InterestedMsg,
		Payload: nil,
	}
}

func NewNotInterested() *Message {
	return &Message{
		Id:      NotInterestedMsg,
		Payload: nil,
	}
}

func NewHave(pieceIndex int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(pieceIndex))
	return &Message{
		Id:      HaveMsg,
		Payload: payload,
	}
}

func NewBitfield(bitfield []byte) *Message {
	return &Message{
		Id:      BitfieldMsg,
		Payload: bitfield,
	}
}

func NewPiece(pieceIndex, begin int, block []byte) *Message {
	payload := make([]byte, 8+len(block))
	binary.BigEndian.PutUint32(payload[0:4], uint32(pieceIndex))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	copy(payload[8:], block)
	return &Message{
		Id:      PieceMsg,
		Payload: payload,
	}
}

func NewCancel(pieceIndex, offset, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(pieceIndex))
	binary.BigEndian.PutUint32(payload[4:8], uint32(offset))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Message{
		Id:      CancelMsg,
		Payload: payload,
	}
}

func NewChoke() *Message {
	return &Message{
		Id:      ChokeMsg,
		Payload: nil, // Choke messages have no payload
	}
}

func NewUnchoke() *Message {
	return &Message{
		Id:      UnchokeMsg,
		Payload: nil, // Unchoke messages have no payload
	}
}

func (m *Message) ProcessHaveMsg() (int, error) {
	if m.Id != HaveMsg {
		return 0, fmt.Errorf("expected HAVE message (Id %d), received Id %d", HaveMsg, m.Id)
	}
	if len(m.Payload) != 4 {
		return 0, fmt.Errorf("malformed paylod, length %v\n", len(m.Payload))
	}
	return int(binary.BigEndian.Uint32(m.Payload)), nil
}

func (m *Message) ProcessPieceMsg(index int, data []byte) (int, error) {
	if m.Id != PieceMsg {
		return 0, fmt.Errorf("expected PIECE (Id %d), got Id %d", PieceMsg, m.Id)
	}
	if len(m.Payload) < 8 {
		return 0, fmt.Errorf("invalid payload size\n")
	}
	pieceIndex := int(binary.BigEndian.Uint32(m.Payload[0:4]))
	if pieceIndex != index {
		return 0, fmt.Errorf("different piece index. expected: %v received: %v\n", index, pieceIndex)
	}
	begin := int(binary.BigEndian.Uint32(m.Payload[4:8]))
	if begin > len(data) {
		return 0, fmt.Errorf("begin offset is out of bounds: %v\n", begin)
	}
	block := m.Payload[8:]

	if begin+len(block) > len(data) {
		return 0, fmt.Errorf("block + offset (%v) bigger than expected(%v)\n\n", len(block)+begin, len(data))
	}
	copy(data[begin:], block)

	return len(block), nil
}

func (m *Message) Name() string {
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
