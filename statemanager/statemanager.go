package statemanager

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"runtime"
	"swiftpeer/client/message"
	"swiftpeer/client/peer"
	"swiftpeer/client/peerconn"
	"time"
)

const maxRequest = 5
const maxBlockSize = 2 << 13

// Torrent used to store the necessary information to download  the peers
type Torrent struct {
	Name        string
	Length      int
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	PeerID      [20]byte
	Peers       []peer.Peer
}

// used to track the progress of a piece
type pieceState struct {
	peerConn   *peerconn.PeerConn
	index      int
	downloaded int
	requested  int
	left       int
	data       []byte
	nPeers     int // counts how many peers are active
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

	pc.Conn.SetDeadline(time.Now().Add(30 * time.Second))
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

func (t *Torrent) startTask(peer peer.Peer, pieceQueue chan *pieceTask, completed chan *pieceCompleted) {
	pc, err := peerconn.NewPeerConn(peer, t.InfoHash)

	if err != nil {
		fmt.Printf("Failed to complete the handshake with %v. Disconnecting\n", peer.IP)
		fmt.Printf("Reason: %v\n", err)
		return
	}

	fmt.Printf("Completed the handshake with %v.\n", peer.IP)

	err = pc.SendUnchoke()
	if err != nil {
		fmt.Printf("Failed to send unchoke to %v: %v\n", peer.IP, err)
		return
	}

	err = pc.SendInterested()
	if err != nil {
		fmt.Printf("Failed to send interested to %v: %v\n", peer.IP, err)
		return
	}

	for pieceTask := range pieceQueue {
		if !pc.Pieces.HavePiece(pieceTask.index) {
			pieceQueue <- pieceTask
			continue
		}

		buff, err := prepareDownload(pc, pieceTask)
		if err != nil {
			fmt.Printf("\nError downloading piece %d from %v:\n", pieceTask.index, peer.IP)
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

func (t *Torrent) Download() ([]byte, error) {
	fmt.Printf("Starting download for%v\n", t.Name)
	piecesQueue := make(chan *pieceTask, len(t.PieceHashes))
	completed := make(chan *pieceCompleted)
	for idx, hash := range t.PieceHashes {
		length := t.computeSize(idx)
		piecesQueue <- &pieceTask{idx, hash, length}
	}

	for _, peer := range t.Peers {
		go t.startTask(peer, piecesQueue, completed)
	}
	log.Printf("pieces in compeleted %v out of %v\n", len(completed), len(t.PieceHashes))
	buf := make([]byte, t.Length)
	finishedPieces := 0

	for finishedPieces < len(t.PieceHashes) {
		piece := <-completed
		begin, end := t.computeBounds(piece.index)

		copy(buf[begin:end], piece.buf)
		finishedPieces++

		log.Printf("pieces in compeleted %v out of %v\n", finishedPieces, len(t.PieceHashes))
		log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", float64(finishedPieces)/float64(len(t.PieceHashes))*100, piece.index, runtime.NumGoroutine()-1)
	}
	close(piecesQueue)

	return buf, nil
}
