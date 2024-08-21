package message

import (
	"bytes"
	"reflect"
	"testing"
)

func TestMessage_Name(t *testing.T) {
	tests := []struct {
		name   string
		fields messageId
		want   string
	}{
		{"ChokeMsg", ChokeMsg, "ChokeMsg"},
		{"UnchokeMsg", UnchokeMsg, "UnchokeMsg"},
		{"InterestedMsg", InterestedMsg, "InterestedMsg"},
		{"NotInterestedMsg", NotInterestedMsg, "NotInterestedMsg"},
		{"HaveMsg", HaveMsg, "HaveMsg"},
		{"BitfieldMsg", BitfieldMsg, "BitfieldMsg"},
		{"RequestMsg", RequestMsg, "RequestMsg"},
		{"PieceMsg", PieceMsg, "PieceMsg"},
		{"CancelMsg", CancelMsg, "CancelMsg"},
		{"PortMsg", PortMsg, "PortMsg"},
		{"UnknownMsg", 255, "UnknownMsg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Id: tt.fields,
			}
			if got := m.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_ProcessHaveMsg(t *testing.T) {
	tests := []struct {
		name    string
		fields  messageId
		payload []byte
		want    int
		wantErr bool
	}{
		{"ValidHaveMsg", HaveMsg, []byte{0x00, 0x00, 0x00, 0x01}, 1, false},
		{"InvalidId", InterestedMsg, []byte{0x00, 0x00, 0x00, 0x01}, 0, true},
		{"MalformedPayload", HaveMsg, []byte{0x00, 0x00, 0x00}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Id:      tt.fields,
				Payload: tt.payload,
			}
			got, err := m.ProcessHaveMsg()
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessHaveMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ProcessHaveMsg() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_ProcessPieceMsg(t *testing.T) {
	tests := []struct {
		name    string
		fields  messageId
		payload []byte
		args    struct {
			index int
			data  []byte
		}
		want    int
		wantErr bool
	}{
		{
			name:    "ValidPieceMsg",
			fields:  PieceMsg,
			payload: append([]byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02}, []byte("test")...),
			args: struct {
				index int
				data  []byte
			}{1, make([]byte, 10)},
			want:    4,
			wantErr: false,
		},
		{
			name:    "InvalidPieceIndex",
			fields:  PieceMsg,
			payload: append([]byte{0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x02}, []byte("test")...),
			args: struct {
				index int
				data  []byte
			}{1, make([]byte, 10)},
			want:    0,
			wantErr: true,
		},
		{
			name:    "MalformedPayload",
			fields:  PieceMsg,
			payload: []byte{0x00, 0x00, 0x00, 0x01},
			args: struct {
				index int
				data  []byte
			}{1, make([]byte, 10)},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Id:      tt.fields,
				Payload: tt.payload,
			}
			got, err := m.ProcessPieceMsg(tt.args.index, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessPieceMsg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ProcessPieceMsg() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_Serialize(t *testing.T) {
	tests := []struct {
		name    string
		id      messageId
		payload []byte
		want    []byte
	}{
		{
			name:    "NilMessage",
			id:      0,
			payload: nil,
			want:    []byte{0x00, 0x00, 0x00, 0x00},
		},
		{
			name:    "ChokeMessage",
			id:      ChokeMsg,
			payload: nil,
			want:    []byte{0x00, 0x00, 0x00, 0x01, ChokeMsg},
		},
		{
			name:    "UnchokeMessage",
			id:      UnchokeMsg,
			payload: nil,
			want:    []byte{0x00, 0x00, 0x00, 0x01, UnchokeMsg},
		},
		{
			name:    "InterestedMessage",
			id:      InterestedMsg,
			payload: nil,
			want:    []byte{0x00, 0x00, 0x00, 0x01, InterestedMsg},
		},
		{
			name:    "NotInterestedMessage",
			id:      NotInterestedMsg,
			payload: nil,
			want:    []byte{0x00, 0x00, 0x00, 0x01, NotInterestedMsg},
		},
		{
			name:    "HaveMessage",
			id:      HaveMsg,
			payload: []byte{0x00, 0x00, 0x00, 0x01},
			want:    []byte{0x00, 0x00, 0x00, 0x05, HaveMsg, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name:    "BitfieldMessage",
			id:      BitfieldMsg,
			payload: []byte{0b10101010, 0b11001100},
			want:    []byte{0x00, 0x00, 0x00, 0x03, BitfieldMsg, 0b10101010, 0b11001100},
		},
		{
			name:    "RequestMessage",
			id:      RequestMsg,
			payload: []byte{0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x08},
			want:    []byte{0x00, 0x00, 0x00, 0x0d, RequestMsg, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x08},
		},
		{
			name:    "PieceMessage",
			id:      PieceMsg,
			payload: []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x01, 0x02, 0x03, 0x04},
			want:    []byte{0x00, 0x00, 0x00, 0x0d, PieceMsg, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x01, 0x02, 0x03, 0x04},
		},
		{
			name:    "CancelMessage",
			id:      CancelMsg,
			payload: []byte{0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x07},
			want:    []byte{0x00, 0x00, 0x00, 0x0d, CancelMsg, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x07},
		},
		{
			name:    "PortMessage",
			id:      PortMsg,
			payload: []byte{0x12, 0x34},
			want:    []byte{0x00, 0x00, 0x00, 0x03, PortMsg, 0x12, 0x34},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m *Message = nil
			if tt.name != "NilMessage" {
				m = &Message{
					Id:      tt.id,
					Payload: tt.payload,
				}
			}

			if got := m.Serialize(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Serialize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewCancel(t *testing.T) {
	tests := []struct {
		name       string
		pieceIndex int
		offset     int
		length     int
		want       *Message
	}{
		{"ValidCancel", 1, 0, 16384, &Message{Id: CancelMsg, Payload: []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x00}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCancel(tt.pieceIndex, tt.offset, tt.length); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCancel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewChoke(t *testing.T) {
	tests := []struct {
		name string
		want *Message
	}{
		{"ValidChoke", &Message{Id: ChokeMsg}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewChoke(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewChoke() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewHave(t *testing.T) {
	tests := []struct {
		name       string
		pieceIndex int
		want       *Message
	}{
		{"ValidHave", 1, &Message{Id: HaveMsg, Payload: []byte{0x00, 0x00, 0x00, 0x01}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHave(tt.pieceIndex); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHave() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewInterested(t *testing.T) {
	tests := []struct {
		name string
		want *Message
	}{
		{"ValidInterested", &Message{Id: InterestedMsg}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewInterested(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewInterested() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewKeepAlive(t *testing.T) {
	tests := []struct {
		name string
		want *Message
	}{
		{"ValidKeepAlive", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewKeepAlive(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewKeepAlive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewNotInterested(t *testing.T) {
	tests := []struct {
		name string
		want *Message
	}{
		{"ValidNotInterested", &Message{Id: NotInterestedMsg}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNotInterested(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNotInterested() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPiece(t *testing.T) {
	tests := []struct {
		name       string
		pieceIndex int
		begin      int
		block      []byte
		want       *Message
	}{
		{
			"ValidPiece",
			1,
			0,
			[]byte("test block"),
			&Message{
				Id:      PieceMsg,
				Payload: append([]byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}, []byte("test block")...),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPiece(tt.pieceIndex, tt.begin, tt.block); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPiece() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewRequest(t *testing.T) {
	tests := []struct {
		name       string
		pieceIndex int
		offset     int
		length     int
		want       *Message
		wantErr    bool
	}{
		{
			"ValidRequest",
			1,
			0,
			16384,
			&Message{
				Id:      RequestMsg,
				Payload: []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x40, 0x00},
			},
			false,
		},
		{
			"InvalidRequestNegativeLength",
			1,
			0,
			-1,
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRequest(tt.pieceIndex, tt.offset, tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewUnchoke(t *testing.T) {
	tests := []struct {
		name string
		want *Message
	}{
		{"ValidUnchoke", &Message{Id: UnchokeMsg}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUnchoke(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUnchoke() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRead(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    *Message
		wantErr bool
	}{
		{
			name:    "ValidHaveMessage",
			input:   []byte{0x00, 0x00, 0x00, 0x05, HaveMsg, 0x00, 0x00, 0x00, 0x01},
			want:    &Message{Id: HaveMsg, Payload: []byte{0x00, 0x00, 0x00, 0x01}},
			wantErr: false,
		},
		{
			name:    "ValidChokeMessage",
			input:   []byte{0x00, 0x00, 0x00, 0x01, ChokeMsg},
			want:    &Message{Id: ChokeMsg, Payload: nil},
			wantErr: false,
		},
		{
			name:    "ValidUnchokeMessage",
			input:   []byte{0x00, 0x00, 0x00, 0x01, UnchokeMsg},
			want:    &Message{Id: UnchokeMsg, Payload: nil},
			wantErr: false,
		},
		{
			name:    "ValidInterestedMessage",
			input:   []byte{0x00, 0x00, 0x00, 0x01, InterestedMsg},
			want:    &Message{Id: InterestedMsg, Payload: nil},
			wantErr: false,
		},
		{
			name:    "ValidNotInterestedMessage",
			input:   []byte{0x00, 0x00, 0x00, 0x01, NotInterestedMsg},
			want:    &Message{Id: NotInterestedMsg, Payload: nil},
			wantErr: false,
		},
		{
			name:    "ValidBitfieldMessage",
			input:   []byte{0x00, 0x00, 0x00, 0x03, BitfieldMsg, 0b10101010, 0b11001100},
			want:    &Message{Id: BitfieldMsg, Payload: []byte{0b10101010, 0b11001100}},
			wantErr: false,
		},
		{
			name:    "ValidRequestMessage",
			input:   []byte{0x00, 0x00, 0x00, 0x0d, RequestMsg, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x08},
			want:    &Message{Id: RequestMsg, Payload: []byte{0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x08}},
			wantErr: false,
		},
		{
			name:    "ValidPieceMessage",
			input:   []byte{0x00, 0x00, 0x00, 0x0d, PieceMsg, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x01, 0x02, 0x03, 0x04},
			want:    &Message{Id: PieceMsg, Payload: []byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x01, 0x02, 0x03, 0x04}},
			wantErr: false,
		},
		{
			name:    "ValidCancelMessage",
			input:   []byte{0x00, 0x00, 0x00, 0x0d, CancelMsg, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x07},
			want:    &Message{Id: CancelMsg, Payload: []byte{0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x07}},
			wantErr: false,
		},
		{
			name:    "ValidPortMessage",
			input:   []byte{0x00, 0x00, 0x00, 0x03, PortMsg, 0x12, 0x34},
			want:    &Message{Id: PortMsg, Payload: []byte{0x12, 0x34}},
			wantErr: false,
		},
		{
			name:    "EmptyMessage-KeepAlive",
			input:   []byte{0x00, 0x00, 0x00, 0x00},
			want:    nil,
			wantErr: false,
		},
		{
			name:    "MalformedMessage-LengthTooShort",
			input:   []byte{0x00, 0x00, 0x00, 0x05, HaveMsg}, // Missing payload bytes
			want:    nil,
			wantErr: true,
		},
		{
			name:    "MalformedMessage-LengthTooLong",
			input:   []byte{0x00, 0x00, 0x00, 0x10, HaveMsg, 0x00, 0x00, 0x00, 0x01}, // Declared length is longer than provided data
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.input)
			got, err := Read(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Read() got = %v, want %v", got, tt.want)
			}
		})
	}
}
