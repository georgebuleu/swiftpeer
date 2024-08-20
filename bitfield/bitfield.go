package bitfield

// A Bitfield represents the pieces that a peer has
type Bitfield []byte

// pieceNo = 9 => byteIdx = 9 / 8 = 1 and  offset = 1
// mask = 00000010 => mask = byte(1 <<7 - 1)
// bf = 11111111.11111111
//7 is used in the mask computing part because of the endianness
//the bitfield is and array so the most significant bit is at index 0 so is little endian, but i needed big endian

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
func (bf Bitfield) SetPiece(pieceNo int) {
	byteIdx := pieceNo / 8
	bitOffset := pieceNo % 8
	mask := byte(1 << (7 - bitOffset))
	if byteIdx < 0 || byteIdx >= len(bf) {
		return
	}
	bf[byteIdx] = bf[byteIdx] | mask

}
