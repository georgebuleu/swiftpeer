package main

import (
	"fmt"
	"os"
	"swiftpeer/client/torrent"
)

func main() {
	file, err := torrent.ParseFile(os.Getenv("TORRENT_FILE_PATH"))

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(file.Announce)
	fmt.Println(file.Info.Name)
	fmt.Println(file.Info.Length)
	fmt.Println(file.Info.PieceLength)
}
