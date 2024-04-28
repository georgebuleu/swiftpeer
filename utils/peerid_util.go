package utils

func GetPeerIdAsBytes(peerId string) [20]byte {
	var id [20]byte
	copy(id[:], peerId[:])
	return id
}
