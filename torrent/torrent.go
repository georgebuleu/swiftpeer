package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"swiftpeer/client/torrent/bencode"
)

const HashLen = sha1.Size

// Torrent TODO: add support for multiple files
type Torrent struct {
	Announce    Announce
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
	file, err := ParseFile()
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
	info, err := ParseInfo()

	if err != nil {
		return [HashLen]byte{}, err
	}
	var buf bytes.Buffer
	err = bencode.NewEncoder(&buf).Encode(info)
	if err != nil {
		return [HashLen]byte{}, err
	}
	return sha1.Sum(buf.Bytes()), nil
}

func toTorrent(f File) (Torrent, error) {
	infoHash, err := hashInfo()
	if err != nil {
		return Torrent{}, err
	}

	pieceHashes, err := splitPieces()
	if err != nil {
		return Torrent{}, err
	}

	return Torrent{
		Announce:    f.Announce,
		Name:        f.Info.Name,
		PieceLength: f.Info.PieceLength,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		Length:      f.Info.Length,
		Files:       f.Info.Files,
	}, nil
}
