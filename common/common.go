package common

import (
	"crypto/rand"
	"fmt"
	"time"
)

const BlockSize = 1 << 14

var ClientVersion = "SP" // SwiftPeer

// GeneratePeerId creates a new peer ID
func GeneratePeerId() [20]byte {
	var id [20]byte
	prefix := fmt.Sprintf("-%.2s%04d-", ClientVersion, time.Now().Year()%10000)
	copy(id[:], prefix)

	_, err := rand.Read(id[8:])
	if err != nil {
		// In case of error, fill with zeroes
		for i := 8; i < 20; i++ {
			id[i] = 0
		}
	}

	return id
}

// PeerIdToString converts a peer ID to a readable string
func PeerIdToString(id [20]byte) string {
	return string(id[:])
}
