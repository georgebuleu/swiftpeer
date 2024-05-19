package main

import (
	"fmt"
	"swiftpeer/client/peerconn"
	"swiftpeer/client/torrent"
	"swiftpeer/client/tracker"
	"sync"
)

const Port int = 6889

// for test purposes
func HandlePeers(peers []tracker.Peer, infoHash [20]byte, wg *sync.WaitGroup) {
	for _, peer := range peers {
		wg.Add(1)
		go func(p tracker.Peer) {
			_, err := peerconn.NewPeerConn(p, infoHash)
			if err != nil {
				fmt.Println(err)
			}
			wg.Done()
		}(peer)
	}
}

func main() {
	var wg sync.WaitGroup
	t := torrent.NewTorrent()
	if t == nil {
		fmt.Println("main: could not load torrent file")
		return
	}
	res, err := tracker.ParseTrackerResponse()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println()
	HandlePeers(res.Peers, t.InfoHash, &wg)
	wg.Wait()
}
