package main

import (
	"fmt"
)

func main() {
	file, err := parseTorrentFile("/home/george/licenta/swiftpeer/testdata/ubuntu-23.10.1-desktop-amd64.iso.torrent")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(file.Announce)
	fmt.Println(file.Info.Name)
	fmt.Println(file.Info.Length)
	fmt.Println(file.Info.PieceLength)

}
