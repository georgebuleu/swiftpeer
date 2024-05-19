package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"swiftpeer/client/bencode"
	"swiftpeer/client/torrent/metadata"
)

const HashLen = sha1.Size

// Torrent TODO: add support for multiple files
type Torrent struct {
	Announce    string
	Name        string
	PieceLength int
	InfoHash    [HashLen]byte
	PieceHashes [][HashLen]byte
	Length      int
	Files       []struct {
		Length int
		Path   string
	}
}

func NewTorrent() *Torrent {
	m := metadata.NewMetadata()
	if m == nil {
		fmt.Printf("torrent: Failed to load metadata\n")
		return nil
	}

	infoHash, err := hashInfo(m)
	if err != nil {
		fmt.Printf("torrent: %v", err)
		return nil
	}
	pieceHashes, err := splitPieces(m)
	if err != nil {
		fmt.Printf("torrent: %v", err)
		return nil
	}

	return &Torrent{
		Announce:    m.Announce,
		Name:        m.Name,
		PieceLength: m.PieceLength,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		Length:      m.Length,
		Files:       m.Files,
	}
}

// hashes the info dict
func hashInfo(m *metadata.Metadata) ([HashLen]byte, error) {
	info := m.InfoDict()
	var buf bytes.Buffer
	err := bencode.NewEncoder(&buf).Encode(info)
	//fmt.Printf("\ninfo: %s\n", buf.String())
	if err != nil {
		return [HashLen]byte{}, err
	}

	//fmt.Printf("\nhash: %v", hex.EncodeToString(sum[:]))

	return sha1.Sum(buf.Bytes()), nil
}

func splitPieces(m *metadata.Metadata) ([][HashLen]byte, error) {
	pieces := []byte(m.Pieces)

	if len(pieces)%sha1.Size != 0 {
		return nil, fmt.Errorf("invalid pieces length")
	}
	numPieces := len(pieces) / sha1.Size
	hashes := make([][HashLen]byte, numPieces)
	for i := 0; i < len(pieces); i += HashLen {
		var hash [HashLen]byte
		copy(hash[:], pieces[i:i+HashLen])
		hashes[i/20] = hash
	}
	return hashes, nil
}
