package main

import (
	"fmt"
	"swiftpeer/client/conn"
	"swiftpeer/client/tracker"
	"sync"
)

const Port int = 6889

func main() {
	var wg sync.WaitGroup
	res, err := tracker.ParseTrackerResponse()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println()
	//torrent.HandlePeersSeq(res.Peers)
	conn.HandlePeers(res.Peers, &wg)
	wg.Wait()
}
