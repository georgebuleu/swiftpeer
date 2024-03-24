package torrent

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"swiftpeer/client/torrent/parser"
)

// ip and event are optional, but that might change
func constructURL(peerID [20]byte, port int) (string, error) {
	m, err := parser.ParseFile()
	if err != nil {
		return "", err

	}
	domain, err := url.Parse(string(m.Announce))
	if err != nil {
		return "", err
	}
	t, err := toTorrent(m)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{fmt.Sprintf("%d", port)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"left":       []string{fmt.Sprintf("%d", t.Length)},
	}
	domain.RawQuery = params.Encode()
	return domain.String(), nil
}

func AnnounceTracker(peerID [20]byte, port int) error {
	u, err := constructURL(peerID, port)
	if err != nil {
		return err
	}
	fmt.Println(u)
	resp, err := http.Get(u)
	if err != nil {
		fmt.Println("error during get request: " + err.Error())
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(b))

	return nil
}
