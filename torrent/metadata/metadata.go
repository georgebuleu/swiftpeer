package metadata

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"swiftpeer/client/bencode"
)

type Metadata struct {
	Announce    string
	Name        string
	PieceLength int
	Pieces      string
	Length      int
	Files       []struct {
		Length int
		Path   string
	}
}

var path = os.Args[1]

func NewMetadata() *Metadata {
	m := new(Metadata)
	err := m.load()
	if err != nil {
		fmt.Printf(err.Error() + "\n")
		return nil
	}
	return m
}

func (m *Metadata) load() error {

	r, err := read()
	if err != nil {
		return fmt.Errorf("couldn't open the file: %v", err)
	}

	decodedData, err := bencode.NewDecoder(bufio.NewReader(r)).Decode()
	if err != nil {
		return fmt.Errorf("couldn't decode the file: %v", err)
	}

	switch data := decodedData.(type) {

	case map[string]interface{}:
		if announce, ok := data["announce"]; ok {

			if url, ok := announce.(string); ok {

				m.Announce = url

			} else {

				return fmt.Errorf("invalid 'announce' field type")

			}

		} else {

			return fmt.Errorf("missing 'announce' field")

		}

		if infoData, ok := data["info"].(map[string]interface{}); ok {

			if _, ok := infoData["length"]; ok {

				if _, ok := infoData["files"]; ok {

					return fmt.Errorf("error: only key length or a key files, but not both or neither")

				}
			}

			if _, ok := infoData["length"]; ok {

				m.Name = infoData["name"].(string)

				m.Length = infoData["length"].(int)

				m.PieceLength = infoData["piece length"].(int)

				m.Pieces = infoData["pieces"].(string)

			} else if filesData, ok := infoData["files"]; ok {

				var files []struct {
					Length int

					Path string
				}

				for _, fileData := range filesData.([]interface{}) {

					fileData := fileData.(map[string]interface{})

					files = append(files, struct {
						Length int

						Path string
					}{

						Length: fileData["length"].(int),

						Path: fileData["path"].(string),
					})

				}

				m.Name = infoData["name"].(string)

				m.Length = infoData["length"].(int)

				m.Pieces = infoData["pieces"].(string)

				m.Files = files

			} else {

				return fmt.Errorf("missing 'length' or 'files' field")

			}
		} else {

			return fmt.Errorf("invalid 'info' field")

		}
	default:
		return fmt.Errorf("invalid format")

	}
	return nil
}

func (m *Metadata) InfoDict() map[string]interface{} {

	//single file case
	if m.Length > 0 {
		return map[string]interface{}{
			"name":         m.Name,
			"piece length": m.PieceLength,
			"pieces":       m.Pieces,
			"length":       m.Length,
		}
	}
	//multiple file case
	if len(m.Files) > 0 {
		var files []map[string]interface{}
		for _, file := range m.Files {
			files = append(files, map[string]interface{}{
				"length": file.Length,
				"path":   file.Path,
			})
		}
		return map[string]interface{}{
			"name":         m.Name,
			"piece length": m.PieceLength,
			"pieces":       m.Pieces,
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
