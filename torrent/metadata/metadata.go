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
		Name        string `bencode:"name"`
		PieceLength int    `bencode:"piece length"`
		Pieces      string `bencode:"pieces"`
		Length      int    `bencode:"length"`
		Files       []struct {
			Length int      `bencode:"length"`
			Path   []string `bencode:"path"`
		}
	}
}

var path = os.Args[1]

func NewMetadata() *Metadata {
	m := new(Metadata)
	r, err := read()
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

func (m *Metadata) InfoDict() map[string]interface{} {

	//single file case
	if m.Info.Length > 0 {
		return map[string]interface{}{
			"name":         m.Info.Name,
			"piece length": m.Info.Length,
			"pieces":       m.Info.Pieces,
			"length":       m.Info.Length,
		}
	}
	//multiple file case
	if len(m.Info.Files) > 0 {
		files := make([]interface{}, 0, len(m.Info.Files))

		for _, file := range m.Info.Files {
			interfaces := make([]interface{}, len(file.Path))
			for i, v := range file.Path {
				interfaces[i] = v
			}
			files = append(files, map[string]interface{}{
				"length": file.Length,
				"path":   interfaces,
			})
		}
		return map[string]interface{}{
			"name":         m.Info.Name,
			"piece length": m.Info.PieceLength,
			"pieces":       m.Info.Pieces,
			"files":        files,
		}
	}
	return nil
}

func read() (io.Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't open the file: %v", err)
	}

	br := bufio.NewReader(file)
	return br, nil
}
