package bencode

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type Decoder struct {
	r   *bufio.Reader
	n   int
	err error
}

func NewDecoder(r *bufio.Reader) *Decoder {
	return &Decoder{r: bufio.NewReader(r)}
}

func (d *Decoder) BytesParsed() int {
	return d.n
}

func (d *Decoder) decodeInteger() (int, error) {
	_, err := d.discardByte()
	if err != nil {
		return 0, err
	}
	s, err := d.readBytes('e')
	if err != nil {
		return 0, err
	}
	if s[0] == '-' && s[1] == '0' {
		return 0, fmt.Errorf("invalid encoding: %v%v", s[0], s[1])
	}
	if s[0] == '0' && s[1] != 'e' {
		return 0, fmt.Errorf("leading encoding is invalid: %v%v", s[0], s[1])
	}
	val, err := strconv.Atoi(string(s[:len(s)-1]))
	if err != nil {
		return 0, err
	}

	return val, nil
}

func (d *Decoder) decodeString() (string, error) {
	lengthStr, err := d.readBytes(':')
	if err != nil {
		return "", err
	}

	length, err := strconv.Atoi(string(lengthStr[:len(lengthStr)-1]))
	if err != nil {
		return "", err
	}

	data := make([]byte, length)
	_, err = d.read(data)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (d *Decoder) decodeList() ([]interface{}, error) {
	result := make([]interface{}, 0)
	_, err := d.discardByte() //consume l
	if err != nil {
		return nil, err
	}
	next, err := d.peek()
	for next != 'e' && err == nil {

		elem, err := d.Decode()
		if err != nil {
			return nil, err
		}

		result = append(result, elem)

		next, err = d.peek()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

	}

	if next == 'e' {
		_, err := d.readByte() // Consume 'e'
		if err != nil && err != io.EOF {
			return nil, err
		}
	}

	return result, nil
}

func (d *Decoder) decodeDictionary() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	_, err := d.discardByte() //consume d
	if err != nil {
		return nil, err
	}
	next, err := d.peek()

	for next != 'e' && err == nil {

		nxtStr, err := d.peek()
		if err != nil {
			return nil, err
		}
		if !d.isValidKey(nxtStr) {
			return nil, fmt.Errorf("keys can be only encoded as string %v", nxtStr)
		}
		key, err := d.decodeString()
		if err != nil {
			return nil, err
		}

		val, err := d.Decode()
		if err != nil {
			return nil, err
		}

		result[key] = val
		next, err = d.peek()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}

	if next == 'e' {
		_, err := d.readByte() // Consume 'e'
		if err != nil && err != io.EOF {
			return nil, err
		}
	}

	return result, nil
}

func (d *Decoder) Decode() (interface{}, error) {
	next, err := d.peek()
	if err != nil {
		return nil, err
	}
	switch next {
	case 'i':
		return d.decodeInteger()
	case 'l':
		return d.decodeList()
	case 'd':
		return d.decodeDictionary()
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return d.decodeString()
	default:
		return nil, fmt.Errorf("unknown token: %q", next)
	}
}

func (d *Decoder) read(data []byte) (int, error) {
	n, err := d.r.Read(data)
	d.n += n
	return n, err
}
func (d *Decoder) readByte() (byte, error) {
	b, err := d.r.ReadByte()
	d.n++
	return b, err
}

func (d *Decoder) readBytes(delim byte) ([]byte, error) {
	data, err := d.r.ReadBytes(delim)
	d.n += len(data)

	return data, err
}

func (d *Decoder) peek() (byte, error) {
	b, err := d.r.Peek(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}
func (d *Decoder) discardByte() (int, error) {
	return d.r.Discard(1)
}

func (d *Decoder) isValidKey(str byte) bool {
	switch str {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	default:
		return false
	}
}