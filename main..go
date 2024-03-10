package main

import (
	"fmt"
	"swiftpeer/client/torrent"
)

func main() {

	hashedInfo, err := torrent.HashInfo()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(hashedInfo)
	fmt.Println(len(hashedInfo))
	//fmt.Println(info["piece length"])

}
