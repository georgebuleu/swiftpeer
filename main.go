package main

import (
	"fmt"
	"os"
	"swiftpeer/client/common"
	"swiftpeer/client/peer"
	"swiftpeer/client/statemanager"
	"swiftpeer/client/torrent"
	"swiftpeer/client/tracker"
)

const Port int = 6881

func main() {
	peers := make(peer.AddrSet)
	outFile := "/home/george/test_licenta/deb2.iso"
	tf := torrent.NewTorrent()
	if tf == nil {
		fmt.Println("main: could not load torrent file")
		return
	}
	err := tracker.GetTorrentData(tf.Announce, tf.AnnounceList, Port, tf.InfoHash, peers)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("len peers: %v\n\n", len(peers))
	for k, _ := range peers {
		fmt.Println(k)
	}
	err = DownloadFile(tf, peers, outFile)
	//if _, ok := r.(*tracker.UdpResponse); ok {
	//	err = DownloadFile(tf, r.(*tracker.UdpResponse).Peers, outFile)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//	fmt.Println()
	//} else if _, ok := r.(*tracker.CompactResponse); ok {
	//	err = DownloadFile(tf, r.(*tracker.CompactResponse).Peers, outFile)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//	fmt.Println()
	//} else if _, ok := r.(*tracker.OriginalResponse); ok {
	//	err = DownloadFile(tf, r.(*tracker.OriginalResponse).Peers, outFile)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//}
	fmt.Println()

}

//func main() {
//	var wg sync.WaitGroup
//	//outFile := "/home/george/test_licenta"
//	tf := torrent.NewTorrent()
//	if tf == nil {
//		fmt.Println("main: could not load torrent file")
//		return
//	}
//	res, _ := tracker.ParseTrackerResponse(tf)
//	HandlePeers(res.Peers, tf.InfoHash, &wg)
//	wg.Wait()
//
//}

//func HandlePeers(peers []tracker.Addr, infoHash [20]byte, wg *sync.WaitGroup) {
//	for _, peer := range peers {
//		wg.Add(1)
//		go func(p tracker.Addr) {
//			conn, err := peerconn.NewPeerConn(p, infoHash)
//			if err != nil {
//				fmt.Println(err)
//			}
//			_, err = conn.Read()
//			if err != nil {
//				fmt.Println(err)
//				return
//			}
//			wg.Done()
//		}(peer)
//	}
//}

func DownloadFile(tf *torrent.Torrent, peers peer.AddrSet, path string) error {

	tr := &statemanager.Torrent{
		Peers:       peers,
		PeerID:      common.GetPeerIdAsBytes(common.PeerId),
		InfoHash:    tf.InfoHash,
		PieceHashes: tf.PieceHashes,
		PieceLength: tf.PieceLength,
		Length:      tf.Length,
		Name:        tf.Name,
	}
	buf, err := tr.Download()
	if err != nil {
		return err
	}

	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()
	_, err = outFile.Write(buf)
	if err != nil {
		return err
	}
	return nil
}
