package main

import (
	"bufio"
	"fmt"
	"os"
	"swiftpeer/client/bencode"
)

type Announce string

type Info struct {
	Name        string
	PieceLength int
	Pieces      string
	Length      int
	Files       []struct {
		Length int
		Path   []string
	}
}
type TorrentFile struct {
	Announce Announce
	Info     Info
}

func parseTorrentFile(filename string) (TorrentFile, error) {
	file, err := os.Open(filename)
	if err != nil {
		return TorrentFile{}, fmt.Errorf("couldn't open the file: %v", err)
	}
	defer file.Close()

	reader := *bufio.NewReader(file)
	decodedData, err := bencode.NewDecoder(&reader).Decode()

	url, ok := decodedData.(map[string]interface{})["announce"].(string)

	if !ok {
		return TorrentFile{}, fmt.Errorf("invalid 'announce' field")
	}

	info, ok := decodedData.(map[string]interface{})["info"].(map[string]interface{})

	if !ok {
		return TorrentFile{}, fmt.Errorf("invalid 'info' field")
	}

	//info, err := parseInfo(path)

	length, ok := info["length"].(int)

	if !ok {
		return TorrentFile{}, fmt.Errorf("invalid 'length' field'")
	}

	pieceLength, ok := info["piece length"].(int)

	if !ok {
		return TorrentFile{}, fmt.Errorf("invalid 'piece length' field'")
	}

	pieces, ok := info["pieces"].(string)

	if !ok {
		return TorrentFile{}, fmt.Errorf("invalid 'pieces' field'")
	}

	return TorrentFile{
		Announce: Announce{url: url},
		Info: Info{
			length:      length,
			pieceLength: pieceLength,
			pieces:      pieces,
		},
	}, err

}

func parseInfo(path string) (map[string]interface{}, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't read from the file")
	}
	decodedData, _, err := decodeBencode(string(data))

	info, ok := decodedData.(map[string]interface{})["info"].(map[string]interface{})

	if !ok {
		return nil, fmt.Errorf("invalid 'info' field")
	}

	return info, err
}
