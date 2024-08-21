package bitfield

import (
	"reflect"
	"testing"
)

func TestBitfield_Count(t *testing.T) {
	tests := []struct {
		name string
		bf   Bitfield
		want int
	}{
		{
			name: "Empty bitfield (16 bits)",
			bf:   NewBitfield(16),
			want: 0,
		},
		{
			name: "All bits unset (32 bits)",
			bf:   NewBitfield(32),
			want: 0,
		},
		{
			name: "All bits set (16 bits)",
			bf:   Bitfield{0xFF, 0xFF},
			want: 16,
		},
		{
			name: "Some bits set (24 bits)",
			bf:   Bitfield{0xAA, 0x55, 0xAA}, // 10101010 01010101 10101010 in binary
			want: 12,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bf.Count(); got != tt.want {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBitfield_HasPiece(t *testing.T) {
	type args struct {
		pieceNo int
	}
	tests := []struct {
		name string
		bf   Bitfield
		args args
		want bool
	}{
		{
			name: "Piece available (16 bits)",
			bf:   Bitfield{0x80, 0x00}, // 10000000 00000000 in binary
			args: args{pieceNo: 0},
			want: true,
		},
		{
			name: "Piece unavailable (32 bits)",
			bf:   Bitfield{0x00, 0x00, 0x00, 0x00}, // All bits unset
			args: args{pieceNo: 10},
			want: false,
		},
		{
			name: "Piece out of range (24 bits)",
			bf:   Bitfield{0xFF, 0xFF, 0xFF},
			args: args{pieceNo: 25},
			want: false,
		},
		{
			name: "Middle bit set (16 bits)",
			bf:   Bitfield{0x00, 0x10}, // 00000000 00010000 in binary
			args: args{pieceNo: 11},
			want: true,
		},
		{
			name: "Piece in second byte",
			bf:   Bitfield{0x00, 0x80}, // 00000000 10000000 in binary
			args: args{pieceNo: 8},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bf.HasPiece(tt.args.pieceNo); got != tt.want {
				t.Errorf("got= %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBitfield_SetPiece(t *testing.T) {
	type args struct {
		pieceNo int
	}
	tests := []struct {
		name string
		bf   Bitfield
		args args
		want Bitfield
	}{
		{
			name: "Set piece in empty bitfield (16 bits)",
			bf:   NewBitfield(16),
			args: args{pieceNo: 10},
			want: Bitfield{0x00, 0x20}, // 00000000 00100000 in binary
		},
		{
			name: "Set piece already set (24 bits)",
			bf:   Bitfield{0x00, 0x80, 0x00},
			args: args{pieceNo: 8},
			want: Bitfield{0x00, 0x80, 0x00}, // 00000000 10000000 00000000 in binary
		},
		{
			name: "Set piece in middle (16 bits)",
			bf:   Bitfield{0x00, 0x00},
			args: args{pieceNo: 7},
			want: Bitfield{0x01, 0x00}, // 00000001 00000000 in binary
		},
		{
			name: "Set piece in second byte",
			bf:   Bitfield{0x00, 0x00},
			args: args{pieceNo: 8},
			want: Bitfield{0x00, 0x80}, // 00000000 10000000 in binary
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.bf.SetPiece(tt.args.pieceNo)
			if !reflect.DeepEqual(tt.bf, tt.want) {
				t.Errorf("got = %v, want %v", tt.bf, tt.want)
			}
		})
	}
}

func TestBitfield_UnsetPiece(t *testing.T) {
	type args struct {
		pieceNo int
	}
	tests := []struct {
		name string
		bf   Bitfield
		args args
		want Bitfield
	}{
		{
			name: "Unset piece in full bitfield (16 bits)",
			bf:   Bitfield{0xFF, 0xFF}, // 11111111 11111111 in binary
			args: args{pieceNo: 12},
			want: Bitfield{0xFF, 0xEF}, // 11111111 11101111 in binary
		},
		{
			name: "Unset piece already unset (16 bits)",
			bf:   Bitfield{0x00, 0x00}, // 00000000 00000000 in binary
			args: args{pieceNo: 10},
			want: Bitfield{0x00, 0x00}, // 00000000 00000000 in binary
		},
		{
			name: "Unset piece in middle (24 bits)",
			bf:   Bitfield{0xFF, 0xFF, 0xFF}, // 11111111 11111111 11111111 in binary
			args: args{pieceNo: 8},
			want: Bitfield{0xFF, 0xFE, 0xFF}, // 11111111 01111111 11111111 in binary
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.bf.UnsetPiece(tt.args.pieceNo)
			if !reflect.DeepEqual(tt.bf, tt.want) {
				t.Errorf("got = %v, want %v", tt.bf, tt.want)
			}
		})
	}
}

func TestNewBitfield(t *testing.T) {
	type args struct {
		numPieces int
	}
	tests := []struct {
		name string
		args args
		want Bitfield
	}{
		{
			name: "Zero pieces",
			args: args{numPieces: 0},
			want: Bitfield{},
		},
		{
			name: "One piece (1 byte)",
			args: args{numPieces: 1},
			want: Bitfield{0x00},
		},
		{
			name: "Eight pieces (1 byte)",
			args: args{numPieces: 8},
			want: Bitfield{0x00},
		},
		{
			name: "Nine pieces (2 bytes)",
			args: args{numPieces: 9},
			want: Bitfield{0x00, 0x00},
		},
		{
			name: "Sixteen pieces (2 bytes)",
			args: args{numPieces: 16},
			want: Bitfield{0x00, 0x00},
		},
		{
			name: "Thirty-two pieces (4 bytes)",
			args: args{numPieces: 32},
			want: Bitfield{0x00, 0x00, 0x00, 0x00},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewBitfield(tt.args.numPieces); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got = %v, want %v", got, tt.want)
			}
		})
	}
}
