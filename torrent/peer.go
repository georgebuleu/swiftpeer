package torrent

import (
	"fmt"
	"io"
	"net"
	"sync"
)

func ClientIdToByteArray(id string) [20]byte {
	var clientId [20]byte
	copy(clientId[:], id[:])
	return clientId
}

func HandlePeers(peers []Peer, wg *sync.WaitGroup) {
	for _, peer := range peers {
		wg.Add(1)
		go func(p Peer) {
			bindToPeer(p)
			wg.Done()
		}(peer)
	}
}

func bindToPeer(peer Peer) {
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
	hash, _ := HashInfo()
	h := NewHandshake(ClientIdToByteArray(ClientId), hash)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("Failed to connect to  %v. %v\n", address, err.Error())
		return
	} else {
		fmt.Printf("Successfully connected to  %v\n", address)
	}

	wrote, err := conn.Write(h.Serialize())
	if err != nil {
		fmt.Printf("Failed to WRITE %v\n", err.Error())
		return
	}
	fmt.Printf("Wrote %v bytes. \n", wrote)

	rh, err := h.Deserialize(io.Reader(conn))

	if err != nil {
		if err != io.EOF {
			fmt.Printf("Failed to READ: %v\n", err.Error())
		} else {
			fmt.Println("Peer closed the connection.")
		}
		return
	}

	if h.InfoHash != rh.InfoHash {
		fmt.Println()
	}

	fmt.Printf("Read bytes :%v\n", rh.InfoHash)
}
