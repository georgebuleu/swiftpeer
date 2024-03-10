package torrent

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"
	"swiftpeer/client/torrent/bencode"
)

var PATH = os.Getenv("TORRENT_FILE")

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
type File struct {
	Announce Announce
	Info     Info
}

func splitPieces() ([][20]byte, error) {
	file, err := ParseFile(PATH)
	pieces := []byte(file.Info.Pieces)
	if err != nil {
		return nil, err
	}
	if len(pieces)%sha1.Size != 0 {
		return nil, fmt.Errorf("invalid pieces length")
	}
	numPieces := len(pieces) / sha1.Size
	hashes := make([][20]byte, numPieces)
	for i := 0; i < len(pieces); i += 20 {
		var hash [20]byte
		copy(hash[:], pieces[i:i+20])
		hashes[i/20] = hash
	}
	return hashes, nil
}

func HashInfo() ([20]byte, error) {
	info, err := ParseInfo(PATH)

	if err != nil {
		return [20]byte{}, err
	}
	var buf bytes.Buffer
	err = bencode.NewEncoder(&buf).Encode(info)
	if err != nil {
		return [20]byte{}, err
	}
	return sha1.Sum(buf.Bytes()), nil
}

/*
used to parse the torrent file and return it as a File type
*/
func ParseFile(path string) (File, error) {
	file, err := os.Open(path)
	if err != nil {
		return File{}, fmt.Errorf("couldn't open the file: %v", err)
	}
	defer file.Close()
	decodedData, err := bencode.NewDecoder(bufio.NewReader(file)).Decode()
	if err != nil {
		return File{}, fmt.Errorf("couldn't decode the file: %v", err)
	}

	var announce Announce
	var info Info

	switch data := decodedData.(type) {
	case map[string]interface{}:
		if announceData, ok := data["announce"]; ok {
			if url, ok := announceData.(string); ok {
				announce = Announce(url)
			} else {
				return File{}, fmt.Errorf("invalid 'announce' field type")
			}
		} else {
			return File{}, fmt.Errorf("missing 'announce' field")
		}

		if infoData, ok := data["info"].(map[string]interface{}); ok {

			if _, ok := infoData["length"]; ok {
				if _, ok := infoData["files"]; ok {
					return File{}, fmt.Errorf("error: only key length or a key files, but not both or neither")
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
				return File{}, fmt.Errorf("missing 'length' or 'files' field")
			}
		} else {
			return File{}, fmt.Errorf("invalid 'info' field")
		}

	default:
		return File{}, fmt.Errorf("invalid format")
	}

	return File{
		Announce: announce,
		Info:     info,
	}, err
}

/*
used to parse the info field from the torrent file and return it as a map
because the encoder and decoder from the bencode package do not support user-defined types
and this used by the HashInfo function to encode the info field and hash it
*/
func ParseInfo(path string) (info map[string]interface{}, err error) {
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
					"name":         infoData["name"].(string),
					"length":       infoData["length"].(int),
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
					"name":         infoData["name"].(string),
					"piece length": infoData["piece length"].(int),
					"pieces":       infoData["pieces"].(string),
					"files":        files,
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
