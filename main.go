package main

import (
	"fmt"
	"swiftpeer/client/common"
	"swiftpeer/client/peer"
	"swiftpeer/client/peerconn"
	"swiftpeer/client/torrent"
	"swiftpeer/client/tracker"
	"sync"
	"time"
)

const port int = 6881

func main() {

	tf, err := torrent.NewTorrent("testdata/debian-12.5.0-amd64-netinst.iso.torrent")
	if tf == nil {
		fmt.Println(err)
		return
	}

	peerId := common.GeneratePeerId()

	peers := make(peer.AddrSet)
	err = tracker.GetTorrentData(tf.Metadata.Announce, tf.Metadata.AnnounceList, port, tf.Metadata.InfoHash, peerId, peers)
	if err != nil {
		fmt.Println(err)
	}
	for p, _ := range peers {
		fmt.Println(p)
	}
	tf.Peers = peers
	testPeerConnections(tf)

}

func testPeerConnectionsSe(t *torrent.Torrent) {
	peerID := common.GeneratePeerId()
	successfulConnections := 0

	totalPeers := len(t.Peers)
	fmt.Printf("Attempting to connect to %d peers sequentially...\n", totalPeers)

	for addr := range t.Peers {
		fmt.Printf("Attempting to connect to peer %s\n", addr)

		startTime := time.Now()
		pc, err := peerconn.NewPeerConn(addr, t.Metadata.InfoHash, peerID)
		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("Failed to connect to peer %s: %v (Time taken: %v)\n", addr, err, duration)
		} else {
			successfulConnections++
			fmt.Printf("Successfully connected to peer %s (Time taken: %v)\n", addr, duration)
			pc.Close()
		}

		// Optional: add a small delay between connection attempts
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("\nConnection test completed.\n")
	fmt.Printf("Successful connections: %d out of %d\n", successfulConnections, totalPeers)
}

func testPeerConnections(t *torrent.Torrent) {
	peerID := common.GeneratePeerId()
	var wg sync.WaitGroup
	successfulConnections := 0
	var mu sync.Mutex

	totalPeers := len(t.Peers)
	fmt.Printf("Attempting to connect to %d peers...\n", totalPeers)

	for addr := range t.Peers {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()

			pc, err := peerconn.NewPeerConn(addr, t.Metadata.InfoHash, peerID)
			if err != nil {
				fmt.Printf("Failed to connect to peer %s: %v\n", addr, err)
				return
			}
			defer pc.Close()

			mu.Lock()
			successfulConnections++
			mu.Unlock()

			fmt.Printf("Successfully connected to peer %s\n", addr)
		}(addr)
	}

	wg.Wait()

	fmt.Printf("\nConnection test completed.\n")
	fmt.Printf("Successful connections: %d out of %d\n", successfulConnections, totalPeers)
}
