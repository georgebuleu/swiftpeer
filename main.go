package main

import (
	"fmt"
	"swiftpeer/client/common"
	"swiftpeer/client/torrent"
)

const Port int = 6881

func main() {

	peerId := common.GeneratePeerId()

	torrentFilePath := "testdata/spider-mantheanimatedseries_archive.torrent"
	outDir := "/home/george/test_licenta/"
	t, err := torrent.NewTorrent(torrentFilePath, peerId, Port, outDir)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = t.Download(outDir)
	if err != nil {
		fmt.Println(err)
	}

}
