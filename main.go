package main

import (
	"fmt"
	"os"
	"path/filepath"
	"swiftpeer/client/common"
	"swiftpeer/client/peer"
	"swiftpeer/client/statemanager"
	"swiftpeer/client/torrent"
	"swiftpeer/client/tracker"
)

const Port int = 6881

func main() {
	peers := make(peer.AddrSet)
	outDir := "/home/george/test_licenta/"
	tf := torrent.NewTorrent()
	if tf == nil {
		fmt.Println("main: could not load torrent file")
		return
	}
	err := tracker.GetTorrentData(tf.Announce, tf.AnnounceList, Port, tf.InfoHash, peers)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("len peers: %v\n\n", len(peers))
	for k, _ := range peers {
		fmt.Println(k)
	}
	err = DownloadFile(tf, peers, outDir)

	fmt.Println()

}

func DownloadFile(tf *torrent.Torrent, peers peer.AddrSet, outDir string) error {

	t := &statemanager.Torrent{
		Peers:       peers,
		PeerID:      common.GetPeerIdAsBytes(common.PeerId),
		InfoHash:    tf.InfoHash,
		PieceHashes: tf.PieceHashes,
		PieceLength: tf.PieceLength,
		Length:      tf.TotalLength,
		Name:        tf.Name,
		Files:       tf.Files,
	}
	buf, err := t.Download()
	if err != nil {
		return err
	}

	var usedBytes int
	for _, file := range t.Files {
		outPath := filepath.Join(outDir, file.Path)

		fmt.Printf("writing to file %q\n", outPath)
		// ensure the directory exists
		baseDir := filepath.Dir(outPath)
		_, err := os.Stat(baseDir)
		if os.IsNotExist(err) {
			err := os.MkdirAll(baseDir, os.ModePerm)
			if err != nil {
				return fmt.Errorf("making output directory: %w", err)
			}
		}
		fileRaw := buf[usedBytes : usedBytes+file.Length]

		// write to the file
		err = os.WriteFile(outPath, fileRaw, os.ModePerm)
		usedBytes += file.Length
		if err != nil {
			return fmt.Errorf("writing to file: %w", err)
		}
	}
	return nil
}
