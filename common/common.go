package common

const BlockSize = 1 << 14

// PeerId refers to my client this, this client.
const PeerId = "-SP1011-IJLasf24lqrI"

func GetPeerIdAsBytes(peerId string) [20]byte {
	var id [20]byte
	copy(id[:], peerId[:])
	return id
}
