package main

import (
	"fmt"
	"swiftpeer/client/common"
	"swiftpeer/client/torrent"
)

const Port int = 6881

func main() {

	peerId := common.GeneratePeerId()

	torrentFilePath := "testdata/nasa.torrent"
	outDir := "/home/george/test_licenta/"
	_, err := torrent.NewTorrent(torrentFilePath, peerId, Port, outDir)
	if err != nil {
		fmt.Println(err)
	}

	//err = tf.Download(outDir)
	//if err != nil {
	//	fmt.Println(err)
	//}

}
