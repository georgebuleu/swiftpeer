package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"swiftpeer/client/torrent/bencode"
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

type TrackerResponse struct {
	FailureReason  string
	WarningMessage string
	Interval       int
	MinInterval    int
	TrackerID      string
	Complete       int
	Incomplete     int
	Peers          []Peer
}

type Peer struct {
	IP   string
	Port int
}

var path = os.Args[1]

func readFile() (io.Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't open the file: %v", err)
	}

	br := bufio.NewReader(file)
	return br, nil
}

//func ParseResponse(reader io.Reader) (TrackerResponse, error) {
//	decodedData, err := bencode.NewDecoder(bufio.NewReader(reader)).Decode()
//	if err != nil {
//		return TrackerResponse{}, fmt.Errorf("couldn't decode the file: %v", err)
//	}
//
//	var trackerResponse TrackerResponse
//
//	switch data := decodedData.(type) {
//	case map[string]interface{}:
//		if failureReason, ok := data["failure reason"].(string); ok {
//			trackerResponse.FailureReason = failureReason
//		}
//		if warningMessage, ok := data["warning message"].(string); ok {
//			trackerResponse.WarningMessage = warningMessage
//		}
//		if interval, ok := data["interval"].(int); ok {
//			trackerResponse.Interval = interval
//		}
//		if minInterval, ok := data["min interval"].(int); ok {
//			trackerResponse.MinInterval = minInterval
//		}
//		if trackerID, ok := data["tracker id"].(string); ok {
//			trackerResponse.TrackerID = trackerID
//		}
//		if complete, ok := data["complete"].(int); ok {
//			trackerResponse.Complete = complete
//		}
//		if incomplete, ok := data["incomplete"].(int); ok {
//			trackerResponse.Incomplete = incomplete
//		}
//		if peers, ok := data["peers"]; ok {
//			trackerResponse.Peers = parsePeers(peers)
//		}
//	default:
//		return TrackerResponse{}, fmt.Errorf("invalid format")
//	}
//
//	return trackerResponse, nil
//}

//func parsePeers(peers interface{}) []Peer {
//
//}

func ParseMetadata() (Metadata, error) {

	r, err := readFile()
	if err != nil {
		return Metadata{}, fmt.Errorf("couldn't open the file: %v", err)

	}

	decodedData, err := bencode.NewDecoder(bufio.NewReader(r)).Decode()

	if err != nil {

		return Metadata{}, fmt.Errorf("couldn't decode the file: %v", err)

	}

	var metadata Metadata

	switch data := decodedData.(type) {

	case map[string]interface{}:

		if announce, ok := data["announce"]; ok {

			if url, ok := announce.(string); ok {

				metadata.Announce = url

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

				metadata.Name = infoData["name"].(string)

				metadata.Length = infoData["length"].(int)

				metadata.PieceLength = infoData["piece length"].(int)

				metadata.Pieces = infoData["pieces"].(string)

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

				metadata.Name = infoData["name"].(string)

				metadata.Length = infoData["length"].(int)

				metadata.Pieces = infoData["pieces"].(string)

				metadata.Files = files

			} else {

				return Metadata{}, fmt.Errorf("missing 'length' or 'files' field")

			}

		} else {

			return Metadata{}, fmt.Errorf("invalid 'info' field")

		}

	default:

		return Metadata{}, fmt.Errorf("invalid format")

	}

	return metadata, nil

}

func GetInfoDictionary(m Metadata) map[string]interface{} {

	if m.Length > 0 {
		return map[string]interface{}{
			"name":         m.Name,
			"piece length": m.PieceLength,
			"pieces":       m.Pieces,
			"length":       m.Length,
		}
	}
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
