package main

import (
	"fmt"
	"swiftpeer/client/common"
	"swiftpeer/client/torrent"
)

const Port int = 6881

func main() {

	peerId := common.GeneratePeerId()

	torrentFilePath := "testdata/debian-12.5.0-amd64-netinst.iso.torrent"
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
