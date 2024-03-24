package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"swiftpeer/client/torrent/bencode"
	"swiftpeer/client/torrent/parser"
)

const HashLen = sha1.Size

// Torrent TODO: add support for multiple files
type Torrent struct {
	Announce    parser.Announce
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

func splitPieces() ([][HashLen]byte, error) {
	file, err := parser.ParseFile()
	pieces := []byte(file.Info.Pieces)
	if err != nil {
		return nil, err
	}
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

func hashInfo() ([HashLen]byte, error) {
	info, err := parser.ParseInfo()
	//fmt.Println(info)

	if err != nil {
		return [HashLen]byte{}, err
	}
	var buf bytes.Buffer
	err = bencode.NewEncoder(&buf).Encode(info)
	//fmt.Printf("\ninfo: %s\n", buf.String())
	if err != nil {
		return [HashLen]byte{}, err
	}

	//fmt.Printf("\nhash: %v", hex.EncodeToString(sum[:]))

	return sha1.Sum(buf.Bytes()), nil
}

func toTorrent(m parser.Metadata) (Torrent, error) {
	infoHash, err := hashInfo()
	if err != nil {
		return Torrent{}, err
	}
	pieceHashes, err := splitPieces()
	if err != nil {
		return Torrent{}, err
	}

	return Torrent{
		Announce:    m.Announce,
		Name:        m.Info.Name,
		PieceLength: m.Info.PieceLength,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		Length:      m.Info.Length,
		Files:       m.Info.Files,
	}, nil
}
