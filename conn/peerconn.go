package conn

import (
	"fmt"
	"io"
	"net"
	"swiftpeer/client/handshake"
	"swiftpeer/client/torrent"
	"swiftpeer/client/tracker"
	"sync"
	"time"
)

type PeerConn struct {
	conn     net.Conn
	peer     tracker.Peer
	peerId   [20]byte
	infoHash [20]byte
	isChoked bool
}

func InitHandshake(conn net.Conn, infoHash [20]byte, peerId [20]byte) {
	hs := handshake.NewHandshake(peerId, infoHash)
	_, err := conn.Write(hs.Serialize())
	if err != nil {
		fmt.Printf("Failed to send handshake with %v : %v", conn.RemoteAddr(), err)
		return
	}
	answer, err := hs.Deserialize(io.Reader(conn))
	if err != nil {
		if err != io.EOF {
			fmt.Printf("Failed to READ: %v\n", err.Error())
		} else {
			fmt.Println("Peer closed the connection.")
		}
		return
	}
	if hs.InfoHash != answer.InfoHash {
		fmt.Println("Different hash")
		return
	}

	fmt.Printf("Read bytes :%v\n", answer.InfoHash)
}

func ClientIdToByteArray(id string) [20]byte {
	var clientId [20]byte
	copy(clientId[:], id[:])
	return clientId
}

func HandlePeers(peers []tracker.Peer, wg *sync.WaitGroup) {
	for _, peer := range peers {
		wg.Add(1)
		go func(p tracker.Peer) {
			bindToPeer(p)
			wg.Done()
		}(peer)
	}
}

func bindToPeer(peer tracker.Peer) {
	ip := net.ParseIP(peer.IP)
	if ip == nil {
		fmt.Printf("invalid IP address: %s", peer.IP)
		return
	}
	var address string
	if ip.To4() != nil {
		address = fmt.Sprintf("%v:%v", peer.IP, peer.Port)
	} else {
		address = fmt.Sprintf("[%v]:%v", peer.IP, peer.Port)
	}
	hash, _ := torrent.HashInfo()
	conn, err := net.DialTimeout("tcp", address, time.Duration(time.Second*10))

	if err != nil {
		fmt.Printf("Failed to connect to  %v. %v\n", address, err.Error())
		return
	} else {
		fmt.Printf("Successfully connected to  %v\n", address)
	}

	InitHandshake(conn, hash, ClientIdToByteArray(handshake.ClientId))

}
