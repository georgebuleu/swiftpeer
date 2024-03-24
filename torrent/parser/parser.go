package parser

import (
	"bufio"
	"fmt"
	"os"
	"swiftpeer/client/torrent/bencode"
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
type Metadata struct {
	Announce Announce
	Info     Info
}

var path = os.Args[1]

/*
used to parse the torrent file and return it as a Metadata type
*/
func ParseFile() (Metadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return Metadata{}, fmt.Errorf("couldn't open the file: %v", err)
	}
	defer file.Close()
	decodedData, err := bencode.NewDecoder(bufio.NewReader(file)).Decode()
	if err != nil {
		return Metadata{}, fmt.Errorf("couldn't decode the file: %v", err)
	}

	var announce Announce
	var info Info

	switch data := decodedData.(type) {
	case map[string]interface{}:
		if announceData, ok := data["announce"]; ok {
			if url, ok := announceData.(string); ok {
				announce = Announce(url)
			} else {
				return Metadata{}, fmt.Errorf("invalid 'announce' field type")
			}
		} else {
			return Metadata{}, fmt.Errorf("missing 'announce' field")
		}

		if infoData, ok := data["info"].(map[string]interface{}); ok {

			if _, ok := infoData["length"]; ok {
				if _, ok := infoData["files"]; ok {
					return Metadata{}, fmt.Errorf("error: only key length or a key files, but not both or neither")
				}
			}

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
				return Metadata{}, fmt.Errorf("missing 'length' or 'files' field")
			}
		} else {
			return Metadata{}, fmt.Errorf("invalid 'info' field")
		}

	default:
		return Metadata{}, fmt.Errorf("invalid format")
	}

	return Metadata{
		Announce: announce,
		Info:     info,
	}, err
}

/*
used to parse the info field from the torrent file and return it as a map
because the encoder and decoder from the bencode package do not support user-defined types
and this used by the HashInfo function to encode the info field and hash it
*/
func ParseInfo() (info map[string]interface{}, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't open the file: %v", err)
	}
	defer file.Close()

	decodedData, err := bencode.NewDecoder(bufio.NewReader(file)).Decode()
	if err != nil {
		return nil, fmt.Errorf("couldn't decode the file: %v", err)
	}

	switch data := decodedData.(type) {
	case map[string]interface{}:

		if infoData, ok := data["info"].(map[string]interface{}); ok {

			if _, ok := infoData["length"]; ok {
				if _, ok := infoData["files"]; ok {
					return nil, fmt.Errorf("only key length or a key files, but not both or neither")
				}
			}

			if _, ok := infoData["length"]; ok {

				info = map[string]interface{}{
					"length":       infoData["length"].(int),
					"name":         infoData["name"].(string),
					"piece length": infoData["piece length"].(int),
					"pieces":       infoData["pieces"].(string),
				}
			} else if filesData, ok := infoData["files"]; ok {
				var files []map[string]interface{}

				for _, fileData := range filesData.([]interface{}) {
					fileData := fileData.(map[string]interface{})
					files = append(files, map[string]interface{}{
						"length": fileData["length"].(int),
						"path":   fileData["path"].(string),
					})
				}

				info = map[string]interface{}{
					"files":        files,
					"name":         infoData["name"].(string),
					"piece length": infoData["piece length"].(int),
					"pieces":       infoData["pieces"].(string),
				}
			} else {
				return nil, fmt.Errorf("missing 'length' or 'files' field")
			}
		} else {
			return nil, fmt.Errorf("invalid 'info' field")
		}
	default:
		return nil, fmt.Errorf("invalid format")
	}
	return info, nil
}
