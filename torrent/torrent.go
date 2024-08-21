package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"path/filepath"
	"swiftpeer/client/bitfield"
	"swiftpeer/client/peer"
	"swiftpeer/client/torrent/metadata"
)

// Torrent represents an active torrent download
type Torrent struct {
	Metadata     *metadata.Metadata
	PieceLength  int
	TotalLength  int64
	NumPieces    int
	Peers        peer.AddrSet
	PiecesStatus bitfield.Bitfield // Bitfield of downloaded pieces
}

func NewTorrent(path string) (*Torrent, error) {
	md, err := metadata.NewMetadataFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata: %v", err)
	}

	t := &Torrent{
		Metadata:    md,
		PieceLength: md.Info.PieceLength,
		TotalLength: md.TotalLength(),
		Peers:       make(peer.AddrSet),
	}

	t.NumPieces = len(md.Info.Pieces) / sha1.Size
	t.PiecesStatus = make([]byte, (t.NumPieces+7)/8)

	return t, nil
}

func (t *Torrent) AddPeer(p peer.Peer) {
	addr, err := p.FormatAddress()
	if err == nil {
		t.Peers[addr] = struct{}{}
	}
}

func (t *Torrent) MarkPieceComplete(index int) {
	byteIndex := index / 8
	bitOffset := index % 8
	t.PiecesStatus[byteIndex] |= 1 << uint(7-bitOffset)
}

func (t *Torrent) IsPieceComplete(index int) bool {
	byteIndex := index / 8
	bitOffset := index % 8
	return t.PiecesStatus[byteIndex]&(1<<uint(7-bitOffset)) != 0
}

func (t *Torrent) DownloadedPieces() int {

	return t.PiecesStatus.Count()

}

func (t *Torrent) Progress() float64 {
	return float64(t.DownloadedPieces()) / float64(t.NumPieces)
}

func (t *Torrent) IsComplete() bool {
	return t.DownloadedPieces() == t.NumPieces
}

// PieceSize computes the piece based of position
func (t *Torrent) PieceSize(index int) int {
	if index == t.NumPieces-1 {
		return int(t.TotalLength % int64(t.PieceLength))
	}
	return t.PieceLength
}

func (t *Torrent) VerifyPiece(index int, pieceData []byte) bool {
	var metadataHash [20]byte
	copy(metadataHash[:], t.Metadata.Info.Pieces[index*sha1.Size:(index+1)*sha1.Size])
	return bytes.Equal(metadataHash[:], pieceData)
}

func (t *Torrent) FilePathForPiece(index int) (string, int64) {
	offset := int64(index) * int64(t.PieceLength)
	// Single file torrent
	if t.Metadata.Info.Length > 0 {
		return t.Metadata.Info.Name, offset
	}
	for _, file := range t.Metadata.Files() {
		if offset < int64(file.Length) {
			return filepath.Join(file.Path...), offset
		}
		offset -= int64(file.Length)
	}
	return "", 0 // Should never happen if the torrent is valid
}
