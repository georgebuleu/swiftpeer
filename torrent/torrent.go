package torrent

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"swiftpeer/client/bitfield"
	"swiftpeer/client/filewriter"
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
	Name         string
	InfoHash     [20]byte
	PieceHashes  [][20]byte
	PeerID       [20]byte
	Files        []FileData
}

type FileData struct {
	Length     int
	Path       string
	Writer     *filewriter.FileWriter
	Downloaded int
	Completed  bool
	Start      int
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
		Name:        md.Info.Name,
		InfoHash:    md.InfoHash,
	}

	t.NumPieces = len(md.Info.Pieces) / sha1.Size
	t.PiecesStatus = make([]byte, (t.NumPieces+7)/8)

	// Initialize PieceHashes
	t.PieceHashes = make([][20]byte, t.NumPieces)
	for i := 0; i < t.NumPieces; i++ {
		copy(t.PieceHashes[i][:], md.Info.Pieces[i*20:(i+1)*20])
	}

	// Initialize Files
	t.Files = make([]FileData, len(md.Files()))
	for i, file := range md.Files() {
		t.Files[i] = FileData{
			Length: file.Length,
			Path:   filepath.Join(file.Path...),
		}
	}

	return t, nil
}

// Existing methods...

func (t *Torrent) ComputeBounds(index int) (int, int) {
	begin := index * t.PieceLength
	end := begin + t.PieceLength

	if end > int(t.TotalLength) {
		end = int(t.TotalLength)
	}
	return begin, end
}

func (t *Torrent) ComputeSize(index int) int {
	begin, end := t.ComputeBounds(index)
	return end - begin
}

func (t *Torrent) ComputeBoundsForFile(file FileData) (int, int) {
	return file.Start, file.Start + file.Length
}

func (t *Torrent) HandlePiece(pieceIndex int, pieceData []byte) error {
	begin, end := t.ComputeBounds(pieceIndex)

	for i := range t.Files {
		file := &t.Files[i]
		if file.Completed {
			continue
		}

		fileStart, fileEnd := t.ComputeBoundsForFile(*file)
		if begin < fileEnd && end > fileStart {
			overlapStart := max(begin, fileStart)
			overlapEnd := min(end, fileEnd)
			offset := overlapStart - fileStart
			data := pieceData[overlapStart-begin : overlapEnd-begin]
			if err := file.Writer.WriteAt(data, offset); err != nil {
				return err
			}

			file.Downloaded += len(data)
			if file.Downloaded >= file.Length {
				file.Completed = true
				if err := file.Writer.Sync(); err != nil {
					return err
				}
				if err := file.Writer.Close(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (t *Torrent) SetupFiles(basePath string) error {
	currentPosition := 0
	for i := range t.Files {
		file := &t.Files[i]
		fullPath := filepath.Join(basePath, file.Path)
		if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
			return err
		}
		writer, err := filewriter.New(fullPath, file.Length)
		if err != nil {
			return err
		}
		file.Writer = writer
		file.Start = currentPosition
		currentPosition += file.Length
	}
	return nil
}

func (t *Torrent) FinalCleanup() error {
	for i := range t.Files {
		file := &t.Files[i]
		if !file.Completed && file.Writer != nil {
			file.Writer.Sync()
			file.Writer.Close()
		}
	}
	return nil
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
