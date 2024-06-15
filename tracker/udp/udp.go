package udp

import (
	"encoding/binary"
	"fmt"
	"net"
	"swiftpeer/client/tracker"
	"time"
)

const (
	connectAction  = 0
	announceAction = 1
)

// magic constant
const (
	protocolId      = 0x41727101980
	connPacketSize  = 16
	announceReqSize = 98
)

func connectToTracker(trackerAddr string) (uint64, error) {
	conn, err := net.Dial("udp", trackerAddr)

	if err != nil {
		return 0, err
	}
	transactionId := uint32(time.Now().UnixNano())
	req := make([]byte, connPacketSize)
	binary.BigEndian.PutUint64(req[:4], protocolId)
	binary.BigEndian.PutUint32(req[8:], connectAction)
	binary.BigEndian.PutUint32(req[12:], transactionId)

	_, err = conn.Write(req)

	if err != nil {
		return 0, err
	}

	if len(req) != connPacketSize {
		return 0, fmt.Errorf("missing data while creating connection")
	}

	resp := make([]byte, connPacketSize)
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Read(resp)
	if err != nil {
		return 0, err
	}

	if len(resp) != connPacketSize {
		return 0, fmt.Errorf("invalid packte size, expected %v, got %v", connPacketSize, len(resp))
	}

	if binary.BigEndian.Uint32(resp[:4]) != connectAction {
		return 0, fmt.Errorf("invalid action")
	}

	if binary.BigEndian.Uint32(resp[4:8]) != transactionId {
		return 0, fmt.Errorf("transaction Id mismatch")
	}
	return binary.BigEndian.Uint64(resp[8:]), nil

}

func announceToTracker(conn net.Conn, trackerAddr string, infoHash [20]byte, clientId [20]byte, port int) ([]tracker.Peer, error) {
	connectionId, err := connectToTracker(trackerAddr)
	if err != nil {
		return nil, err
	}

	transactionId := uint32(time.Now().UnixNano())
	req := make([]byte, announceReqSize)

	binary.BigEndian.PutUint64(req[:8], connectionId)
	binary.BigEndian.PutUint32(req[8:12], announceAction)
	binary.BigEndian.PutUint32(req[12:16], transactionId)
	copy(req[16:], infoHash[:])
	copy(req[36:], clientId[:])
	binary.BigEndian.PutUint64(req[56:], 0)            //downloaded
	binary.BigEndian.PutUint64(req[64:], 0)            //left
	binary.BigEndian.PutUint64(req[72:], 0)            //uploaded
	binary.BigEndian.PutUint32(req[80:], 0)            //event
	binary.BigEndian.PutUint32(req[84:], 0)            //ip address
	binary.BigEndian.PutUint32(req[88:], 0)            //key
	binary.BigEndian.PutUint32(req[92:], -1)           //num_want
	binary.BigEndian.PutUint16(req[96:], uint16(port)) //port

	_, err = conn.Write(req)

	resp := make([]byte, 512) //TODO: proper size
	_, err = conn.Read(resp)
	if len(resp) < 20 {
		return nil, err
	}
	if binary.BigEndian.Uint32(resp[0:4]) != announceAction {
		return nil, fmt.Errorf("invalid action")
	}

	if binary.BigEndian.Uint32(resp[4:8]) != transactionId {
		return nil, fmt.Errorf("transactionId mismatch")
	}

	_ = binary.BigEndian.Uint32(resp[8:12])         //interval
	_ = binary.BigEndian.Uint32(resp[12:16])        //leechers
	seeders := binary.BigEndian.Uint32(resp[16:20]) //seeders

	peers := make([]tracker.Peer, seeders)

	for i := 20; i < len(resp); i += 6 {
		peers = append(peers, tracker.Peer{
			IP:   string(resp[i : i+4]),
			Port: int(binary.BigEndian.Uint32(resp[i+4 : i+6])),
		})
	}
	return peers, err
}
