package tracker

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/url"
	"swiftpeer/client/common"
	"swiftpeer/client/peer"
	"time"
)

const (
	connectAction  = 0
	announceAction = 1
)

const (
	// magic constant
	protocolId           = 0x41727101980
	connPacketSize       = 16
	announceReqSize      = 98
	maxRetries           = 1
	minRetries           = 0
	connIdExpirationTime = 60 * time.Second
)

type UdpResponse struct {
	Interval int
	Leechers int
	Peers    []peer.Peer
}

type UdpClient struct {
	Conn     *net.UDPConn
	InfoHash [20]byte
	Port     int
	PeerId   [20]byte
	connId   uint64
	connTime time.Time
}

func getPeersFromUDPTracker(u *url.URL, infoHash [20]byte, port int) (*UdpResponse, error) {
	client, err := NewClient(u, infoHash, port)
	if err != nil {
		return nil, err
	}

	return client.GetData()
}

func NewClient(url *url.URL, infoHash [20]byte, port int) (*UdpClient, error) {
	udpAddr, err := net.ResolveUDPAddr(url.Scheme, url.Host)
	if err != nil {
		return nil, fmt.Errorf("couldn't resolve udp addr: %v", err)
	}
	c, err := net.DialUDP("udp", nil, udpAddr)

	if err != nil {
		return nil, fmt.Errorf("while dialing udp tracker: %v", err)
	}
	err = c.SetReadBuffer(2048)
	if err != nil {
		return nil, fmt.Errorf("upd read buffer: %v", err)
	}
	return &UdpClient{
		Conn:     c,
		InfoHash: infoHash,
		Port:     port,
		PeerId:   common.GetPeerIdAsBytes(common.PeerId),
	}, nil
}

func (client *UdpClient) GetData() (*UdpResponse, error) {
	var err error
	for n := 0; n <= maxRetries; n++ {
		client.Conn.SetDeadline(time.Now().Add(15 * (1 << n) * time.Second))
		err = client.connect()
		if err == nil {
			client.Conn.SetDeadline(time.Time{})
			break
		}
		if n == maxRetries {
			return nil, fmt.Errorf("max retries reached for connect: %v", err)
		}
	}

	r := new(UdpResponse)
	for n := 0; n <= maxRetries; n++ {

		if time.Since(client.connTime) > connIdExpirationTime {
			client.Conn.SetDeadline(time.Now().Add(time.Second * 5))
			err = client.connect()
			if err != nil {
				return nil, fmt.Errorf("failed to reconnect: %v", err)
			}
		}
		client.Conn.SetDeadline(time.Now().Add(15 * (1 << n) * time.Second))
		err = client.requestData(r)
		if err == nil {
			client.Conn.SetDeadline(time.Time{})
			break
		}
		if n == maxRetries {
			return nil, fmt.Errorf("max retries reached for announce: %v", err)
		}
	}

	return r, nil
}

func (client *UdpClient) connect() error {

	transactionId := uint32(time.Now().UnixNano())
	req := make([]byte, connPacketSize)
	binary.BigEndian.PutUint64(req[:8], protocolId)
	binary.BigEndian.PutUint32(req[8:12], connectAction)
	binary.BigEndian.PutUint32(req[12:16], transactionId)

	_, err := client.Conn.Write(req)

	if err != nil {
		return err
	}

	if len(req) != connPacketSize {
		return fmt.Errorf("missing data while creating connection")
	}

	resp := make([]byte, connPacketSize)
	_, err = client.Conn.Read(resp)
	if err != nil {
		return err
	}

	if len(resp) != connPacketSize {
		return fmt.Errorf("invalid packte size, expected %v, got %v", connPacketSize, len(resp))
	}

	if binary.BigEndian.Uint32(resp[:4]) != connectAction {
		return fmt.Errorf("invalid action")
	}

	if binary.BigEndian.Uint32(resp[4:8]) != transactionId {
		return fmt.Errorf("transaction Id mismatch")
	}
	client.connTime = time.Now()
	client.connId = binary.BigEndian.Uint64(resp[8:])
	return nil
}

func (client *UdpClient) requestData(data *UdpResponse) error {

	transactionId := uint32(time.Now().UnixNano())
	req := make([]byte, announceReqSize)

	binary.BigEndian.PutUint64(req[:8], client.connId)
	binary.BigEndian.PutUint32(req[8:12], announceAction)
	binary.BigEndian.PutUint32(req[12:16], transactionId)
	copy(req[16:], client.InfoHash[:])
	copy(req[36:], client.PeerId[:])
	binary.BigEndian.PutUint64(req[56:], 0)                   //downloaded
	binary.BigEndian.PutUint64(req[64:], 0)                   //left
	binary.BigEndian.PutUint64(req[72:], 0)                   //uploaded
	binary.BigEndian.PutUint32(req[80:], 0)                   //event
	binary.BigEndian.PutUint32(req[84:], 0)                   //ip address
	binary.BigEndian.PutUint32(req[88:], 0)                   //key
	binary.BigEndian.PutUint32(req[92:], 0xFFFFFFFF)          //num_want
	binary.BigEndian.PutUint16(req[96:], uint16(client.Port)) //port

	_, err := client.Conn.Write(req)

	resp := make([]byte, 2048) //TODO: find a better default size
	n, err := client.Conn.Read(resp)
	resp = resp[:n]
	if len(resp) < 20 {
		return err
	}
	if binary.BigEndian.Uint32(resp[0:4]) != announceAction {
		return fmt.Errorf("invalid action")
	}

	if binary.BigEndian.Uint32(resp[4:8]) != transactionId {
		return fmt.Errorf("transactionId mismatch")
	}

	data.Interval = int(binary.BigEndian.Uint32(resp[8:12]))  //interval
	data.Leechers = int(binary.BigEndian.Uint32(resp[12:16])) //leechers
	_ = binary.BigEndian.Uint32(resp[16:20])                  //seeders

	var peers []peer.Peer

	for i := 20; i+6 < len(resp); i += 6 {
		peers = append(peers, peer.Peer{
			IP:   fmt.Sprintf("%d.%d.%d.%d", resp[i], resp[i+1], resp[i+2], resp[i+3]),
			Port: int(binary.BigEndian.Uint16(resp[i+4 : i+6])),
		})
	}
	data.Peers = peers
	return nil
}
