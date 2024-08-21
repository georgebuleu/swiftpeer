package bitfield

import "math/bits"

// A Bitfield represents the pieces that a peer has
type Bitfield []byte

// NewBitfield creates a new Bitfield with the given number of pieces
func NewBitfield(numPieces int) Bitfield {
	return make(Bitfield, (numPieces+7)/8)
}

// pieceNo = 9 => byteIdx = 9 / 8 = 1 and  offset = 1
// mask = 00000010 => mask = byte(1 <<7 - 1)
// bf = 11111111.11111111
//7 is used in the mask computing part because of the endianness
//the bitfield is and array so the most significant bit is at index 0 so is little endian, but i needed big endian

// HasPiece returns true if the peer has the piece at the given index.
func (bf Bitfield) HasPiece(pieceNo int) bool {
	byteIdx := pieceNo / 8
	bitOffset := pieceNo % 8
	mask := byte(1 << (7 - bitOffset))

	if byteIdx < 0 || byteIdx >= len(bf) {
		return false
	}

	return (bf[byteIdx] & mask) != 0
}

// pieceNo = 9 => byteIdx = 9 / 8 = 1 and  offset = 1
// mask = 00000010 => mask = byte(1 << 1)
// bf = 11111111.11111111

// SetPiece sets a piece as available
func (bf Bitfield) SetPiece(pieceNo int) {
	byteIdx := pieceNo / 8
	bitOffset := pieceNo % 8
	mask := byte(1 << (7 - bitOffset))
	if byteIdx < 0 || byteIdx >= len(bf) {
		return
	}
	bf[byteIdx] = bf[byteIdx] | mask
}

// UnsetPiece sets a piece as unavailable
func (bf Bitfield) UnsetPiece(pieceNo int) {
	byteIdx, bitOffset := pieceNo/8, uint(pieceNo%8)
	if byteIdx < len(bf) {
		bf[byteIdx] &^= 1 << bitOffset
	}
}

// Count returns the number of available pieces
func (bf Bitfield) Count() int {
	count := 0
	for _, element := range bf {
		count += bits.OnesCount8(element)
	}
	return count
}
