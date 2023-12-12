package bencode

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

func TestDecoder_BytesParsed(t *testing.T) {
	r := *bufio.NewReader(strings.NewReader("i42e"))
	decoder := NewDecoder(&r)

	_, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Error decoding: %v", err)
	}

	bytesParsed := decoder.BytesParsed()
	expectedBytesParsed := len("i42e") - 1 //len() - 1 because it discards 'e'
	if bytesParsed != expectedBytesParsed {
		t.Errorf("BytesParsed() = %v, want %v", bytesParsed, expectedBytesParsed)
	}
}

func TestDecoder_Decode(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    interface{}
		wantErr bool
	}{

		{"Integer", "i42e", 42, false},
		{"String", "5:hello", "hello", false},
		{"List", "li42e3:fooe", []interface{}{42, "foo"}, false},
		{"Dictionary", "d3:foo3:bare", map[string]interface{}{"foo": "bar"}, false},
		//TODO fix the invalid input test, parameters of the are not the same for all encoding types
		//{"InvalidInput", "invalid", nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tc.input))
			decoder := NewDecoder(r)

			got, err := decoder.Decode()

			if (err != nil) != tc.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tc.wantErr)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Decode() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDecoder_decodeDictionary(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    map[string]interface{}
		wantErr bool
	}{
		{"EmptyDictionary", "de", map[string]interface{}{}, false},
		{"SingleElement", "d3:key5:valuee", map[string]interface{}{"key": "value"}, false},
		{"MultipleElements", "d3:foo3:bar3:baz3:quxe", map[string]interface{}{"foo": "bar", "baz": "qux"}, false},
		{"NestedDictionary", "d4:dictd4:key14:val14:key24:val2ee", map[string]interface{}{"dict": map[string]interface{}{"key1": "val1", "key2": "val2"}}, false},
		{"InvalidInput", "invalid", nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tc.input))
			decoder := NewDecoder(r)

			got, err := decoder.decodeDictionary()

			if (err != nil) != tc.wantErr {
				t.Errorf("decodeDictionary() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("decodeDictionary() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDecoder_DecodeInteger(t *testing.T) {
	testCases := []struct {
		input     string
		expected  int
		wantError bool
	}{
		{"i123e", 123, false},
		{"i0e", 0, false},
		{"i-42e", -42, false},
		{"i-0e", 0, true},
		{"i02e", 0, true},
	}

	for _, tc := range testCases {
		r := bufio.NewReader(strings.NewReader(tc.input))
		decoder := NewDecoder(r)

		result, err := decoder.Decode()
		if (err != nil) != tc.wantError {
			t.Errorf("Error decoding input %s: %v", tc.input, err)
		}

		if got, ok := result.(int); ok {
			if got != tc.expected {
				t.Errorf("For input %s, expected %d, but got %d", tc.input, tc.expected, got)
			}
		} else {
			t.Errorf("For input %s, expected an integer result, but got %v", tc.input, result)
		}
	}
}

func TestDecoder_decodeList(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    []interface{}
		wantErr bool
	}{
		{"EmptyList", "le", []interface{}{}, false},
		{"SingleElement", "li42ee", []interface{}{42}, false},
		{"MultipleElements", "li42e3:fooe", []interface{}{42, "foo"}, false},
		{"NestedList", "li42eli1ei2eeli3eeli4eee", []interface{}{42, []interface{}{1, 2}, []interface{}{3}, []interface{}{4}}, false},
		{"FullyNestedList", "li42eli1ei2eli3eli4eeeee", []interface{}{42, []interface{}{1, 2, []interface{}{3, []interface{}{4}}}}, false},
		{"InvalidInput", "invalid", nil, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tc.input))
			decoder := NewDecoder(r)

			got, err := decoder.decodeList()

			if (err != nil) != tc.wantErr {
				t.Errorf("decodeList() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("decodeList() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDecoder_decodeString(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"EmptyString", "0:", "", false},
		{"ShortString", "4:abcd", "abcd", false},
		{"LongString", "11:hello world", "hello world", false},
		{"InvalidInput", "invalid", "", true},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tc.input))
			decoder := NewDecoder(r)

			got, err := decoder.decodeString()

			if (err != nil) != tc.wantErr {
				t.Errorf("decodeString() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if got != tc.want {
				t.Errorf("decodeString() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDecoder_read(t *testing.T) {
	testCases := []struct {
		name       string
		input      string
		readData   []byte
		want       int
		wantErr    bool
		wantErrMsg string
	}{
		{"ReadEmptyData", "data", []byte{}, 0, false, ""},
		{"ReadSomeData", "abcd", make([]byte, 2), 2, false, ""},
		{"ReadAllData", "xyz", make([]byte, 3), 3, false, ""},
		{"ReadMoreThanAvailableData", "123", make([]byte, 4), 3, false, "EOF"},
		{"InvalidInput", "", make([]byte, 2), 0, true, "EOF"},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tc.input))
			decoder := NewDecoder(r)

			got, err := decoder.read(tc.readData)

			if (err != nil) != tc.wantErr {
				t.Errorf("read() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if tc.wantErr && err != nil && !strings.Contains(err.Error(), tc.wantErrMsg) {
				t.Errorf("read() error message = %v, wantErr %v", err, tc.wantErrMsg)
				return
			}

			if got != tc.want {
				t.Errorf("read() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDecoder_readByte(t *testing.T) {
	testCases := []struct {
		name       string
		input      string
		want       byte
		wantErr    bool
		wantErrMsg string
	}{
		{"ReadByteFromEmptyInput", "", 0, true, "EOF"},
		{"ReadFirstByte", "abc", 'a', false, ""},
		{"ReadNextByte", "xyz", 'x', false, ""},
		{"ReadByteFromEnd", "p", 'p', false, ""},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tc.input))
			decoder := NewDecoder(r)

			got, err := decoder.readByte()

			if (err != nil) != tc.wantErr {
				t.Errorf("readByte() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if tc.wantErr && err != nil && !strings.Contains(err.Error(), tc.wantErrMsg) {
				t.Errorf("readByte() error message = %v, wantErr %v", err, tc.wantErrMsg)
				return
			}

			if got != tc.want {
				t.Errorf("readByte() got = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDecoder_readBytes(t *testing.T) {
	testCases := []struct {
		name       string
		input      string
		delimiter  byte
		want       []byte
		wantErr    bool
		wantErrMsg string
	}{
		{"ReadBytesFromEmptyInput", "", 'e', []byte(""), true, "EOF"},
		{"ReadBytesUntilDelimiter", "abcd:xyz", ':', []byte("abcd:"), false, ""},
		{"ReadBytesUntilEnd", "1234567", '6', []byte("123456"), false, ""},
		{"ReadBytesInvalidDelimiter", "hello", 0, []byte("hello"), true, "EOF"},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tc.input))
			decoder := NewDecoder(r)

			got, err := decoder.readBytes(tc.delimiter)

			if (err != nil) != tc.wantErr {
				t.Errorf("readBytes() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			if tc.wantErr && err != nil && !strings.Contains(err.Error(), tc.wantErrMsg) {
				t.Errorf("readBytes() error message = %v, wantErr %v", err, tc.wantErrMsg)
				return
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("readBytes() got = %v, want %v", got, tc.want)
			}
		})
	}
}
