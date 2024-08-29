package main

import (
	"flag"
	"fmt"
	"os"
	"swiftpeer/client/common"
	"swiftpeer/client/torrent"
)

const Port int = 6881

func main() {

	torrentFilePath := flag.String("t", "", "Path to the torrent file")
	outDir := flag.String("o", "", "Output directory for downloaded files")
	flag.Parse()

	if *torrentFilePath == "" || *outDir == "" {
		fmt.Println("Usage: program -t <torrent-file-path> -o <output-directory>")
		os.Exit(1)
	}

	peerId := common.GeneratePeerId()

	t, err := torrent.NewTorrent(*torrentFilePath, peerId, Port, *outDir)
	if err != nil {
		fmt.Println("Error creating torrent:", err)
		return
	}

	err = t.Download(*outDir)
	if err != nil {
		fmt.Println("Error downloading torrent:", err)
	}

}
