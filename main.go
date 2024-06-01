package main

import (
	"fmt"
	"os"
	"swiftpeer/client/common"
	"swiftpeer/client/peerconn"
	"swiftpeer/client/statemanager"
	"swiftpeer/client/torrent"
	"swiftpeer/client/tracker"
	"sync"
)

const Port int = 6889

func main() {
	outFile := "/home/george/test_licenta/down1.iso"
	tf := torrent.NewTorrent()
	if tf == nil {
		fmt.Println("main: could not load torrent file")
		return
	}
	res, err := tracker.ParseTrackerResponse(tf)
	err = DownloadFile(tf, &res, outFile)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println()
}

//func main() {
//	var wg sync.WaitGroup
//	//outFile := "/home/george/test_licenta"
//	tf := torrent.NewTorrent()
//	if tf == nil {
//		fmt.Println("main: could not load torrent file")
//		return
//	}
//	res, _ := tracker.ParseTrackerResponse(tf)
//	HandlePeers(res.Peers, tf.InfoHash, &wg)
//	wg.Wait()
//
//}

func HandlePeers(peers []tracker.Peer, infoHash [20]byte, wg *sync.WaitGroup) {
	for _, peer := range peers {
		wg.Add(1)
		go func(p tracker.Peer) {
			conn, err := peerconn.NewPeerConn(p, infoHash)
			if err != nil {
				fmt.Println(err)
			}
			_, err = conn.Read()
			if err != nil {
				fmt.Println(err)
				return
			}
			wg.Done()
		}(peer)
	}
}

func DownloadFile(tf *torrent.Torrent, trackerRes *tracker.TrackerResponse, path string) error {

	tr := &statemanager.Torrent{
		Peers:       trackerRes.Peers,
		PeerID:      common.GetPeerIdAsBytes(common.PeerId),
		InfoHash:    tf.InfoHash,
		PieceHashes: tf.PieceHashes,
		PieceLength: tf.PieceLength,
		Length:      tf.Length,
		Name:        tf.Name,
	}
	buf, err := tr.Download()
	if err != nil {
		return err
	}

	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()
	_, err = outFile.Write(buf)
	if err != nil {
		return err
	}
	return nil
}
