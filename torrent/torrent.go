package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"path/filepath"
	"swiftpeer/client/bencode"
	"swiftpeer/client/torrent/metadata"
)

const HashLen = sha1.Size

// Torrent TODO: add support for multiple files
type Torrent struct {
	Announce     string
	AnnounceList [][]string
	Name         string
	PieceLength  int
	InfoHash     [HashLen]byte
	PieceHashes  [][HashLen]byte
	TotalLength  int
	Files        []struct {
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
	if len(m.Info.Files) == 0 && m.Info.Length == 0 {
		fmt.Printf("torrent: no length or files\n")
	}

	t := &Torrent{
		Announce:     m.Announce,
		AnnounceList: m.AnnounceList,
		Name:         m.Info.Name,
		PieceLength:  m.Info.PieceLength,
		InfoHash:     infoHash,
		PieceHashes:  pieceHashes,
	}

	if m.Info.Length != 0 {
		t.Files = append(t.Files, struct {
			Length int
			Path   string
		}{
			Length: m.Info.Length,
			Path:   m.Info.Name,
		})
		t.TotalLength = m.Info.Length
	} else {
		for _, file := range m.Info.Files {
			paths := append([]string{m.Info.Name}, file.Path...)
			t.Files = append(t.Files, struct {
				Length int
				Path   string
			}{
				Length: file.Length,
				Path:   filepath.Join(paths...),
			})
			t.TotalLength += file.Length
		}

	}

	return t
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
	pieces := []byte(m.Info.Pieces)

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
