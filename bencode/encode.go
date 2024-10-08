package bencode

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

// Encode ignores private fields of struct
func (e *Encoder) Encode(v interface{}) error {
	buf := &bytes.Buffer{}

	if err := e.bencode(buf, reflect.ValueOf(v)); err != nil {
		return err
	}
	_, err := buf.WriteTo(e.w)
	return err
}

func (e *Encoder) bencode(buf *bytes.Buffer, v reflect.Value) error {
	switch v.Kind() {
	case reflect.String:
		e.encodeString(buf, v.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		e.encodeInt(buf, v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		e.encodeUint(buf, v.Uint())
	case reflect.Array, reflect.Slice:
		return e.encodeList(buf, v)
	case reflect.Map:
		return e.encodeDict(buf, v)
	case reflect.Struct:
		return e.encodeStruct(buf, v)
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return fmt.Errorf("cannot encode nil value")
		}
		return e.bencode(buf, v.Elem())
	default:
		return fmt.Errorf("incompatible type: %T", v.Type())
	}
	return nil
}

func (e *Encoder) encodeString(buf *bytes.Buffer, s string) {
	buf.WriteString(strconv.Itoa(len(s)))
	buf.WriteByte(':')
	buf.WriteString(s)
}

func (e *Encoder) encodeInt(buf *bytes.Buffer, i int64) {
	buf.WriteByte('i')
	buf.WriteString(strconv.FormatInt(i, 10))
	buf.WriteByte('e')
}

func (e *Encoder) encodeUint(buf *bytes.Buffer, u uint64) {
	buf.WriteByte('i')
	buf.WriteString(strconv.FormatUint(u, 10))
	buf.WriteByte('e')
}

func (e *Encoder) encodeList(buf *bytes.Buffer, v reflect.Value) error {
	buf.WriteByte('l')
	for i := 0; i < v.Len(); i++ {
		if err := e.bencode(buf, v.Index(i)); err != nil {
			return err
		}
	}
	buf.WriteByte('e')
	return nil
}

func (e *Encoder) encodeDict(buf *bytes.Buffer, v reflect.Value) error {

	buf.WriteByte('d')
	keys := v.MapKeys()

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})

	for _, key := range keys {
		e.encodeString(buf, key.String())
		if err := e.bencode(buf, v.MapIndex(key)); err != nil {
			return err
		}
	}
	buf.WriteByte('e')
	return nil
}

func (e *Encoder) encodeStruct(buf *bytes.Buffer, v reflect.Value) error {
	buf.WriteByte('d')
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		fieldValue := v.Field(i)
		key := field.Name
		tag := field.Tag.Get("bencode")
		if tag != "" {
			if tag == "-" {
				continue
			}
			parts := strings.Split(tag, ",")
			key = parts[0]
			if len(parts) > 1 && parts[1] == "omitempty" {
				if isEmptyValue(fieldValue) {
					continue
				}
			}
		}
		e.encodeString(buf, key)
		if err := e.bencode(buf, fieldValue); err != nil {
			return err
		}
	}
	buf.WriteByte('e')
	return nil
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
