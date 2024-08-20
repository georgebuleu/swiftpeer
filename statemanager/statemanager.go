package statemanager

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"swiftpeer/client/filewriter"
	"swiftpeer/client/message"
	"swiftpeer/client/peer"
	"swiftpeer/client/peerconn"
	"time"
)

const maxRequest = 5
const maxBlockSize = 2 << 13

type FileData struct {
	Length     int
	Path       string
	Writer     *filewriter.FileWriter
	Downloaded int
	Completed  bool
	Start      int
}

// Torrent used to store the necessary information to download  the peers
type Torrent struct {
	Name        string
	Length      int
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	PeerID      [20]byte
	Peers       peer.AddrSet
	Files       []FileData
}

// used to track the progress of a piece
type pieceState struct {
	peerConn   *peerconn.PeerConn
	index      int
	downloaded int
	requested  int
	left       int
	data       []byte
}

type pieceTask struct {
	index  int
	hash   [20]byte
	length int
}

type pieceCompleted struct {
	index int
	buf   []byte
}

func (s *pieceState) handleMessage() error {

	m, err := s.peerConn.Read()
	if err != nil {
		return err
	}

	//checks for keep alive message which has length 0, no id and no payload
	if m == nil {
		return nil
	}

	switch m.Id {
	case message.ChokeMsg:
		s.peerConn.IsChoked = true
	case message.UnchokeMsg:
		s.peerConn.IsChoked = false
	case message.PieceMsg:
		received, err := m.ProcessPieceMsg(s.index, s.data)
		if err != nil {
			fmt.Println("error during piece message")
			return err
		}
		s.downloaded += received
		s.left--
	case message.HaveMsg:
		index, err := m.ProcessHaveMsg()
		if err != nil {
			fmt.Println("error during unchoke message")
			return err
		}
		s.peerConn.Pieces.SetPiece(index)
	default:
		return nil
	}
	return nil
}

func prepareDownload(pc *peerconn.PeerConn, task *pieceTask) ([]byte, error) {
	state := pieceState{
		peerConn: pc,
		index:    task.index,
		data:     make([]byte, task.length),
	}

	pc.Conn.SetDeadline(time.Now().Add(25 * time.Second))
	defer pc.Conn.SetDeadline(time.Time{})

	for state.downloaded < task.length {
		if !state.peerConn.IsChoked {
			for state.left < maxRequest && state.requested < task.length {
				blockSize := maxBlockSize

				if task.length-state.requested < blockSize {
					blockSize = task.length - state.requested
				}

				err := pc.SendRequestMsg(task.index, state.requested, blockSize)
				if err != nil {
					log.Printf("Failed to send request message for piece %d: %v\n", task.index, err)
					return nil, err
				}
				state.left++
				state.requested += blockSize
			}
		}

		err := state.handleMessage()
		if err != nil {
			fmt.Println("\nerror while handling message in preparing download")
			return nil, err
		}
	}

	return state.data, nil
}

func (t *Torrent) startTask(peer string, pieceQueue chan *pieceTask, completed chan *pieceCompleted) {
	pc, err := peerconn.NewPeerConn(peer, t.InfoHash)

	if err != nil {
		fmt.Printf("Failed to complete the handshake with %v. Disconnecting\n", peer)
		return
	}

	defer pc.Conn.Close()

	fmt.Printf("Completed the handshake with %v.\n", peer)

	err = pc.SendUnchoke()
	if err != nil {
		fmt.Printf("Failed to send unchoke to %v: %v\n", peer, err)
	}

	err = pc.SendInterested()
	if err != nil {
		fmt.Printf("Failed to send interested to %v: %v\n", peer, err)
	}

	for pieceTask := range pieceQueue {
		if !pc.Pieces.HasPiece(pieceTask.index) {
			pieceQueue <- pieceTask
			continue
		}

		buff, err := prepareDownload(pc, pieceTask)
		if err != nil {
			fmt.Printf("\nError downloading piece %d from %v:\n", pieceTask.index, peer)
			fmt.Println(err)
			pieceQueue <- pieceTask
			return
		}

		valid := checkIntegrity(pieceTask, buff)
		if !valid {
			pieceQueue <- pieceTask
			continue
		} else {
			pc.SendHave(pieceTask.index)
			completed <- &pieceCompleted{pieceTask.index, buff}
		}
	}
}

func (t *Torrent) computeBounds(index int) (int, int) {
	begin := index * t.PieceLength
	end := begin + t.PieceLength

	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func (t *Torrent) computeSize(index int) int {
	begin, end := t.computeBounds(index)
	return end - begin
}

func checkIntegrity(task *pieceTask, data []byte) bool {
	h := sha1.Sum(data)
	if !bytes.Equal(h[:], task.hash[:]) {
		fmt.Printf("index %v did not pass integrity check\n", task.index)
		return false
	}
	return true
}

func (t *Torrent) Download(path string) error {

	if err := t.setupFiles(path); err != nil {
		return err
	}
	defer t.finalCleanup()

	fmt.Printf("Starting download for%v\n", t.Name)
	piecesQueue := make(chan *pieceTask, len(t.PieceHashes))
	completed := make(chan *pieceCompleted)
	for idx, hash := range t.PieceHashes {
		length := t.computeSize(idx)
		piecesQueue <- &pieceTask{idx, hash, length}
	}

	for p, _ := range t.Peers {
		go t.startTask(p, piecesQueue, completed)

	}
	log.Printf("pieces in compeleted %v out of %v\n", len(completed), len(t.PieceHashes))

	finishedPieces := 0

	timeout := time.After(20 * time.Second)

	for finishedPieces < len(t.PieceHashes) {
		select {
		case piece := <-completed:
			// Directly write to the appropriate file using memory-mapped region
			if err := t.handlePiece(piece.index, piece.buf); err != nil {
				fmt.Printf("Failed to handle piece %d: %v\n", piece.index, err)
				return err
			}
			finishedPieces++
			log.Printf("Downloaded piece #%d out of #%d\n", finishedPieces, len(t.PieceHashes))
			timeout = time.After(30 * time.Second) // Reset timeout after each successful piece handling
			log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", float64(finishedPieces)/float64(len(t.PieceHashes))*100, piece.index, runtime.NumGoroutine()-1)

		case <-timeout:
			fmt.Printf("Timeout: No pieces completed within the last 30 seconds")
			return fmt.Errorf("download timeout")
		}
	}
	close(piecesQueue)

	return nil
}

func (t *Torrent) setupFiles(basePath string) error {
	currentPosition := 0
	for i, file := range t.Files {
		fullPath := filepath.Join(basePath, file.Path)
		if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
			return err
		}
		writer, err := filewriter.New(fullPath, file.Length)
		if err != nil {
			return err
		}
		t.Files[i].Writer = writer
		t.Files[i].Start = currentPosition
		currentPosition += file.Length
	}
	return nil
}

func (t *Torrent) computeBoundsForFile(file FileData) (int, int) {
	return file.Start, file.Start + file.Length
}

func (t *Torrent) handlePiece(pieceIndex int, pieceData []byte) error {
	begin, end := t.computeBounds(pieceIndex)

	for _, file := range t.Files {
		if file.Completed {
			continue
		}

		fileStart, fileEnd := t.computeBoundsForFile(file)
		//Check if the piece is between the bounds of the file
		if begin < fileEnd && end > fileStart {
			overlapStart := max(begin, fileStart)
			overlapEnd := min(end, fileEnd)
			offset := overlapStart - fileStart
			data := pieceData[overlapStart-begin : overlapEnd-begin]
			if err := file.Writer.WriteAt(data, offset); err != nil {
				return err
			}

			// Update downloaded count and check for completion
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

func (t *Torrent) finalCleanup() error {
	for _, file := range t.Files {
		if !file.Completed {
			if file.Writer != nil {
				file.Writer.Sync()
				file.Writer.Close()
			}
		}
	}
	return nil
}
