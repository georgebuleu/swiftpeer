package main

import (
	"fmt"
	"swiftpeer/client/torrent"
)

const Port int = 6889

func main() {
	peerID := [20]byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j'}
	err := torrent.AnnounceTracker(peerID, Port)
	if err != nil {
		fmt.Println(err)
	}
}
