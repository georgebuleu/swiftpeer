package bencode

import (
	"bufio"
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
	s, err := d.readBytes('e')
	if err != nil {
		return 0, err
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

	data, err := d.readBytes(byte(length))
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (d *Decoder) decodeList() ([]interface{}, error) {
	var result []interface{}

	for {
		next, err := d.peek()
		if err != nil {
			return nil, err
		}
		if next == 'e' {
			_, err := d.readByte()
			return result, err
		}

		elem, err := d.Decode()
		if err != nil {
			return nil, err
		}
		result = append(result, elem)
	}

}

func (d *Decoder) decodeDictionary() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for {
		next, err := d.peek()
		if err != nil {
			return nil, err
		}
		if next == 'e' {
			_, err := d.readByte()
			return result, err
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
	}
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
	default:
		return d.decodeString()
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
