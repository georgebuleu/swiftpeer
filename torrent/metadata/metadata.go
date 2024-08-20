package metadata

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"swiftpeer/client/bencode"
)

type Metadata struct {
	Announce     string     `bencode:"announce"`
	AnnounceList [][]string `bencode:"announce-list"`
	Info         struct {
		Length int `bencode:"length"`
		Files  []struct {
			Length int      `bencode:"length"`
			Path   []string `bencode:"path"`
		} `bencode:"files,omitempty"`
		Name        string `bencode:"name"`
		PieceLength int    `bencode:"piece length"`
		Pieces      string `bencode:"pieces"`
	}
}

var path = os.Args[1]

func NewMetadata() *Metadata {
	m := new(Metadata)
	r, err := Read()
	if err != nil {
		log.Fatal(err)
	}
	bdecoder := bencode.NewDecoder(bufio.NewReader(r))
	err = bdecoder.Decode(m)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(m.Info.Length)
	fmt.Println(m.Info.Name)
	fmt.Println(m.Info.PieceLength)

	return m
}

func Read() (io.Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't open the file: %v", err)
	}

	br := bufio.NewReader(file)
	return br, nil
}
