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
		Path   string
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
	if err != nil {
		return TorrentFile{}, fmt.Errorf("couldn't decode the file: %v", err)
	}

	var announce Announce
	var info Info

	switch data := decodedData.(type) {
	case map[string]interface{}:
		if announceData, ok := data["announce"]; ok {
			if url, ok := announceData.(string); ok {
				announce = Announce(url)
			} else {
				return TorrentFile{}, fmt.Errorf("invalid 'announce' field type")
			}
		} else {
			return TorrentFile{}, fmt.Errorf("missing 'announce' field")
		}

		if infoData, ok := data["info"].(map[string]interface{}); ok {

			if _, ok := infoData["length"]; ok {

				info = Info{
					Name:        infoData["name"].(string),
					Length:      infoData["length"].(int),
					PieceLength: infoData["piece length"].(int),
					Pieces:      infoData["pieces"].(string),
				}
			} else if filesData, ok := infoData["files"]; ok {
				var files []struct {
					Length int
					Path   string
				}

				for _, fileData := range filesData.([]interface{}) {
					fileData := fileData.(map[string]interface{})
					files = append(files, struct {
						Length int
						Path   string
					}{
						Length: fileData["length"].(int),
						Path:   fileData["path"].(string),
					})
				}

				info = Info{
					Name:        infoData["name"].(string),
					PieceLength: infoData["piece length"].(int),
					Pieces:      infoData["pieces"].(string),
					Files:       files,
				}
			} else {
				return TorrentFile{}, fmt.Errorf("missing 'length' or 'files' field")
			}
		} else {
			return TorrentFile{}, fmt.Errorf("invalid 'info' field")
		}

	default:
		return TorrentFile{}, fmt.Errorf("invalid format")
	}

	return TorrentFile{
		Announce: announce,
		Info:     info,
	}, err
}
