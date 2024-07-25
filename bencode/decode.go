package bencode

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type Decoder struct {
	r *bufio.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: bufio.NewReader(r)}
}

func (d *Decoder) Decode(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("bencode: Decode requires a non-nil pointer")
	}
	return d.bdecode(val.Elem())
}

func (d *Decoder) bdecode(v reflect.Value) error {
	next, err := d.r.Peek(1)
	if err != nil {
		return err
	}

	// Handle interface{} type
	if v.Kind() == reflect.Interface && v.IsNil() {
		v.Set(reflect.New(reflect.TypeOf((*interface{})(nil)).Elem()).Elem())
	}

	switch next[0] {
	case 'i':
		if v.Kind() == reflect.Interface {
			var i int64
			err := d.decodeInt(reflect.ValueOf(&i).Elem())
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(i))
		} else {
			return d.decodeInt(v)
		}
	case 'l':
		if v.Kind() == reflect.Interface {
			var slice []interface{}
			sliceValue := reflect.ValueOf(&slice).Elem()
			err := d.decodeList(sliceValue)
			if err != nil {
				return err
			}
			v.Set(sliceValue)
		} else {
			return d.decodeList(v)
		}
	case 'd':
		if v.Kind() == reflect.Interface {
			m := make(map[string]interface{})
			mapValue := reflect.ValueOf(m)
			err := d.decodeDict(mapValue)
			if err != nil {
				return err
			}
			v.Set(mapValue)
		} else {
			return d.decodeDict(v)
		}
	default:
		if v.Kind() == reflect.Interface {
			var s string
			err := d.decodeString(reflect.ValueOf(&s).Elem())
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(s))
		} else {
			return d.decodeString(v)
		}
	}
	return nil
}

func (d *Decoder) decodeInt(v reflect.Value) error {
	d.r.Discard(1) // Consume 'i'
	numStr, err := d.readUntil('e')
	if err != nil {
		return err
	}
	num, err := strconv.ParseInt(string(numStr), 10, 64)
	if err != nil {
		return err
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.OverflowInt(num) {
			return fmt.Errorf("bencode: integer overflow")
		}
		v.SetInt(num)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if num < 0 {
			return fmt.Errorf("bencode: cannot bdecode negative integer into unsigned type")
		}
		if v.OverflowUint(uint64(num)) {
			return fmt.Errorf("bencode: integer overflow")
		}
		v.SetUint(uint64(num))
	default:
		return fmt.Errorf("bencode: cannot bdecode integer into %v", v.Type())
	}
	return nil
}

func (d *Decoder) decodeString(v reflect.Value) error {
	lenStr, err := d.readUntil(':')
	if err != nil {
		return err
	}
	length, err := strconv.Atoi(string(lenStr))
	if err != nil {
		return err
	}
	data := make([]byte, length)
	_, err = io.ReadFull(d.r, data)
	if err != nil {
		return err
	}
	switch v.Kind() {
	case reflect.String:
		v.SetString(string(data))
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			v.SetBytes(data)
		} else {
			return fmt.Errorf("bencode: cannot bdecode string into %v", v.Type())
		}
	default:
		return fmt.Errorf("bencode: cannot bdecode string into %v", v.Type())
	}
	return nil
}

func (d *Decoder) decodeList(v reflect.Value) error {
	d.r.Discard(1) // Consume 'l'

	if v.Kind() != reflect.Slice {
		v.Set(reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf((*interface{})(nil)).Elem()), 0, 0))
	}

	for {
		next, err := d.r.Peek(1)
		if err != nil {
			return err
		}
		if next[0] == 'e' {
			d.r.Discard(1)
			return nil
		}

		var elem reflect.Value
		if v.Type().Elem().Kind() == reflect.Interface {
			// If the slice element type is interface{}, create a new interface{} value
			elem = reflect.New(reflect.TypeOf((*interface{})(nil)).Elem()).Elem()
		} else {
			// Otherwise, create a new value of the slice element type
			elem = reflect.New(v.Type().Elem()).Elem()
		}

		if err := d.bdecode(elem); err != nil {
			return err
		}

		v.Set(reflect.Append(v, elem))
	}
}

func (d *Decoder) decodeDict(v reflect.Value) error {
	d.r.Discard(1) // Consume 'd'
	switch v.Kind() {
	case reflect.Map:
		return d.decodeDictToMap(v)
	case reflect.Struct:
		return d.decodeDictToStruct(v)
	default:
		return fmt.Errorf("bencode: cannot bdecode dict into %v", v.Type())
	}
}

func (d *Decoder) decodeDictToMap(v reflect.Value) error {
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}
	for {
		next, err := d.r.Peek(1)
		if err != nil {
			return err
		}
		if next[0] == 'e' {
			d.r.Discard(1)
			return nil
		}
		key := reflect.New(v.Type().Key()).Elem()
		if err := d.decodeString(key); err != nil {
			return err
		}
		elem := reflect.New(v.Type().Elem()).Elem()
		if err := d.bdecode(elem); err != nil {
			return err
		}
		v.SetMapIndex(key, elem)
	}
}

func (d *Decoder) decodeDictToStruct(v reflect.Value) error {
	typ := v.Type()
	fieldMap := make(map[string]reflect.Value)

	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := v.Field(i)

		// Check if the field is exported
		if field.PkgPath != "" {
			continue
		}

		tag := field.Tag.Get("bencode")
		if tag == "" {
			// If no tag, use the field name
			fieldMap[strings.ToLower(field.Name)] = fieldValue
		} else if tag == "-" {
			// Skip this field if tag is "-"
			continue
		} else {
			// Use the tag value
			fieldMap[tag] = fieldValue
		}
	}

	for {
		next, err := d.r.Peek(1)
		if err != nil {
			return err
		}
		if next[0] == 'e' {
			d.r.Discard(1)
			return nil
		}

		var keyStr string
		if err := d.decodeString(reflect.ValueOf(&keyStr).Elem()); err != nil {
			return err
		}

		fieldValue, ok := fieldMap[keyStr]
		if !ok {
			if err := d.skipValue(); err != nil {
				return err
			}
			continue
		}

		if err := d.bdecode(fieldValue); err != nil {
			return err
		}
	}
}

func (d *Decoder) readUntil(delim byte) ([]byte, error) {
	data, err := d.r.ReadBytes(delim)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 && data[len(data)-1] == delim {
		return data[:len(data)-1], nil
	}
	return data, nil
}

func (d *Decoder) skipValue() error {
	next, err := d.r.Peek(1)
	if err != nil {
		return err
	}
	switch next[0] {
	case 'i':
		_, err = d.readUntil('e')
	case 'l':
		d.r.Discard(1)
		for {
			next, err := d.r.Peek(1)
			if err != nil {
				return err
			}
			if next[0] == 'e' {
				d.r.Discard(1)
				return nil
			}
			if err := d.skipValue(); err != nil {
				return err
			}
		}
	case 'd':
		d.r.Discard(1)
		for {
			next, err := d.r.Peek(1)
			if err != nil {
				return err
			}
			if next[0] == 'e' {
				d.r.Discard(1)
				return nil
			}
			if err := d.skipValue(); err != nil { // Skip key
				return err
			}
			if err := d.skipValue(); err != nil { // Skip value
				return err
			}
		}
	default:
		lengthStr, err := d.readUntil(':')
		if err != nil {
			return err
		}
		length, err := strconv.Atoi(string(lengthStr))
		if err != nil {
			return err
		}
		_, err = io.CopyN(io.Discard, d.r, int64(length))
	}
	return err
}
