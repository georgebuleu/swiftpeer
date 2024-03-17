package bencode

import (
	"fmt"
	"io"
	"sort"
	"strconv"
)

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

func (e *Encoder) Encode(data interface{}) error {
	encodedData, err := e.bencode(data)
	if err != nil {
		return err
	}
	_, err = e.w.Write(encodedData)
	return err
}

func (e *Encoder) bencode(data interface{}) ([]byte, error) {
	switch data.(type) {
	case string:
		return e.encodeString(data.(string)), nil
	case int:
		return e.encodeInteger(data.(int)), nil
	case []interface{}:
		return e.encodeList(data.([]interface{}))
	case map[string]interface{}:
		return e.encodeDict(data.(map[string]interface{}))
	default:
		return nil, fmt.Errorf("incompatible type")
	}
}

func (e *Encoder) encodeString(data string) []byte {
	return []byte(fmt.Sprintf("%d:%s", len(data), data))
}

func (e *Encoder) encodeInteger(data int) []byte {
	return []byte("i" + strconv.Itoa(data) + "e")
}

func (e *Encoder) encodeList(data []interface{}) ([]byte, error) {
	encodedBytes := make([]byte, 0)
	encodedBytes = append(encodedBytes, "l"...)
	for _, val := range data {
		tmp, err := e.bencode(val)
		if err != nil {
			return nil, err
		}
		encodedBytes = append(encodedBytes, tmp...)
	}
	encodedBytes = append(encodedBytes, "e"...)
	return encodedBytes, nil
}

func (e *Encoder) encodeDict(data map[string]interface{}) ([]byte, error) {
	encodedBytes := make([]byte, 0)
	encodedBytes = append(encodedBytes, "d"...)
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		{
			encodedBytes = append(encodedBytes, e.encodeString(key)...)

			encodedVal, err := e.bencode(data[key])
			if err != nil {
				return nil, err
			}
			encodedBytes = append(encodedBytes, encodedVal...)
		}
	}
	encodedBytes = append(encodedBytes, "e"...)
	return encodedBytes, nil
}
