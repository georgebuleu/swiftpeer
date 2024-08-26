package metadata

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"swiftpeer/client/bencode"
	"time"
)

type File struct {
	Length int      `bencode:"length"`
	Path   []string `bencode:"path"`
}

type Info struct {
	Length      int    `bencode:"length,omitempty"`
	Files       []File `bencode:"files,omitempty"`
	Name        string `bencode:"name"`
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
	Private     int    `bencode:"-"`
}

type Metadata struct {
	Announce     string     `bencode:"announce"`
	AnnounceList [][]string `bencode:"announce-list,omitempty"`
	CreationDate int64      `bencode:"creation date,omitempty"`
	Comment      string     `bencode:"comment,omitempty"`
	CreatedBy    string     `bencode:"created by,omitempty"`
	Encoding     string     `bencode:"encoding,omitempty"`
	Info         Info       `bencode:"info"`
	InfoHash     [20]byte   `bencode:"-"` // Not part of bencode, calculated separately
	Private      int        `bencode:"private,omitempty"`
}

func NewMetadataFromFile(path string) (*Metadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't open the file: %v", err)
	}
	defer file.Close()

	return NewMetadataFromReader(file)
}

func NewMetadataFromReader(r io.Reader) (*Metadata, error) {
	m := new(Metadata)
	decoder := bencode.NewDecoder(bufio.NewReader(r))
	if err := decoder.Decode(m); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %v", err)
	}

	if err := m.calculateInfoHash(); err != nil {
		return nil, fmt.Errorf("failed to calculate info hash: %v", err)
	}

	return m, nil
}

func (m *Metadata) calculateInfoHash() error {
	var buf bytes.Buffer
	err := bencode.NewEncoder(&buf).Encode(m.Info)
	if err != nil {
		return err
	}
	m.InfoHash = sha1.Sum(buf.Bytes())
	return nil
}

func (m *Metadata) TotalLength() int64 {
	if m.Info.Length > 0 {
		return int64(m.Info.Length)
	}
	var total int64
	for _, file := range m.Info.Files {
		total += int64(file.Length)
	}
	return total
}

func (m *Metadata) PieceHashes() ([][20]byte, error) {
	pieceCount := len(m.Info.Pieces) / 20
	if len(m.Info.Pieces)%20 != 0 {
		return nil, fmt.Errorf("invalid pieces length")
	}
	hashes := make([][20]byte, pieceCount)
	for i := 0; i < pieceCount; i++ {
		copy(hashes[i][:], m.Info.Pieces[i*20:(i+1)*20])
	}
	return hashes, nil
}

func (m *Metadata) Files() []File {
	if m.Info.Length > 0 {
		return []File{{Length: m.Info.Length, Path: []string{m.Info.Name}}}
	}
	return m.Info.Files
}

func (m *Metadata) FullPath(basePath string) []string {
	var paths []string
	for _, file := range m.Files() {
		fullPath := filepath.Join(append([]string{basePath, m.Info.Name}, file.Path...)...)
		paths = append(paths, fullPath)
	}
	return paths
}

func (m *Metadata) CreationTime() time.Time {
	return time.Unix(m.CreationDate, 0)
}

func (m *Metadata) IsPrivate() bool {
	return m.Private == 1
}
