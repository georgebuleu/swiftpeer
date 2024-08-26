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
	"time"
)

const Port int = 6881

func main() {
	torrentFilePath := "testdata/debian-12.5.0-amd64-netinst.iso.torrent"
	peers := make(peer.AddrSet)
	outDir := "/home/george/test_licenta/"
	tf, err := torrent.NewTorrent(torrentFilePath)
	if err != nil {
		fmt.Println(err)
	}
	peerId := common.GeneratePeerId()

	err = tracker.GetTorrentData(tf.Metadata.Announce, tf.Metadata.AnnounceList, Port, tf.InfoHash, peerId, peers)
	if err != nil {
		fmt.Println(err)
	}
	for p, _ := range peers {
		fmt.Println(p)
	}

	fmt.Println("Peers2")
	peers2 := make(peer.AddrSet)
	time.Sleep(35 * time.Second)
	err = tracker.GetTorrentData(tf.Metadata.Announce, tf.Metadata.AnnounceList, Port, tf.InfoHash, peerId, peers2)
	if err != nil {
		fmt.Println(err)
	}

	for p, _ := range peers2 {
		peers[p] = struct{}{}
	}

	err = DownloadFile(tf, peers, peerId, outDir)

	fmt.Println()

}

func DownloadFile(tf *torrent.Torrent, peers peer.AddrSet, peerId [20]byte, outDir string) error {

	t := &statemanager.Torrent{
		Peers:       peers,
		PeerID:      peerId,
		InfoHash:    tf.InfoHash,
		PieceHashes: tf.PieceHashes,
		PieceLength: tf.PieceLength,
		Length:      int(tf.TotalLength),
		Name:        tf.Name,
		Files:       make([]statemanager.FileData, len(tf.Files)),
	}
	for i, _ := range tf.Files {
		t.Files[i].Path = tf.Files[i].Path
		t.Files[i].Length = tf.Files[i].Length
	}

	for _, file := range t.Files {
		outPath := filepath.Join(outDir, file.Path)
		baseDir := filepath.Dir(outPath)
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
				return fmt.Errorf("making output directory: %w", err)
			}
		}
	}

	if err := t.Download(outDir); err != nil {
		return err
	}
	return nil
}
