package bitfield

type Bitfield []byte

// pieceNo = 9 => byteIdx = 9 / 8 = 1 and  offset = 1
// mask = 00000010 => mask = byte(1 << 1)
// bf = 11111111.11111111
func (bf Bitfield) HavePiece(pieceNo int) bool {
	byteIdx := pieceNo / 8
	bitOffset := pieceNo % 8
	mask := byte(1 << bitOffset)
	return (bf[byteIdx] & mask) != 0
}

// pieceNo = 9 => byteIdx = 9 / 8 = 1 and  offset = 1
// mask = 00000010 => mask = byte(1 << 1)
// bf = 11111111.11111111
func (bf Bitfield) SetPiece(pieceNo int) {
	byteIdx := pieceNo / 8
	bitOffset := pieceNo % 8
	mask := byte(1 << bitOffset)
	bf[byteIdx] = bf[byteIdx] | mask
}
