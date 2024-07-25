package bencode

import (
	"bytes"
	"reflect"
	"testing"
)

func TestDecodeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"5:hello", "hello"},
		{"0:", ""},
		{"4:test", "test"},
		{"13:longer string", "longer string"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			reader := bytes.NewReader([]byte(tt.input))
			decoder := NewDecoder(reader)

			var result string
			err := decoder.Decode(&result)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDecodeInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"i42e", 42},
		{"i0e", 0},
		{"i-42e", -42},
		{"i1000000e", 1000000},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			reader := bytes.NewReader([]byte(tt.input))
			decoder := NewDecoder(reader)

			var result int64
			err := decoder.Decode(&result)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestDecodeList(t *testing.T) {
	tests := []struct {
		input    string
		expected []interface{}
	}{
		{"li1ei2ei3ee", []interface{}{int64(1), int64(2), int64(3)}},
		{"l1:a1:b1:ce", []interface{}{"a", "b", "c"}},
		{"li42e4:testli1ei2eee", []interface{}{int64(42), "test", []interface{}{int64(1), int64(2)}}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			reader := bytes.NewReader([]byte(tt.input))
			decoder := NewDecoder(reader)

			var result []interface{}
			err := decoder.Decode(&result)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDecodeDictToMap(t *testing.T) {
	tests := []struct {
		input    string
		expected map[string]interface{}
	}{
		{
			"d1:ai1e1:bi2ee",
			map[string]interface{}{"a": int64(1), "b": int64(2)},
		},
		{
			"d3:numi42e4:test5:valuee",
			map[string]interface{}{"num": int64(42), "test": "value"},
		},
		{
			"d4:listli1ei2ei3ee3:str4:teste",
			map[string]interface{}{"list": []interface{}{int64(1), int64(2), int64(3)}, "str": "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			reader := bytes.NewReader([]byte(tt.input))
			decoder := NewDecoder(reader)

			var result map[string]interface{}
			err := decoder.Decode(&result)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDecodeDictToStruct(t *testing.T) {
	type NestedStruct struct {
		Value string `bencode:"value"`
	}

	type TestStruct struct {
		Name       string `bencode:"name"`
		Age        int    `bencode:"age"`
		Email      string `bencode:"email_address"`
		Ignore     string `bencode:"-"`
		Default    string
		unexported string
		Nested     NestedStruct      `bencode:"nested"`
		IntSlice   []int             `bencode:"int_slice"`
		StringMap  map[string]string `bencode:"string_map"`
	}

	tests := []struct {
		name     string
		input    string
		expected TestStruct
	}{
		{
			name:  "Basic struct decoding",
			input: "d4:name5:Alice3:agei20e13:email_address15:alice@email.com7:default5:valuee",
			expected: TestStruct{
				Name:    "Alice",
				Age:     20,
				Email:   "alice@email.com",
				Default: "value",
			},
		},
		{
			name:  "Struct with missing fields",
			input: "d4:name3:Bob3:agei25ee",
			expected: TestStruct{
				Name: "Bob",
				Age:  25,
			},
		},
		{
			name:  "Struct with extra fields",
			input: "d4:name4:Jane3:agei30e13:email_address14:jane@email.com5:extra5:field7:default6:value2e",
			expected: TestStruct{
				Name:    "Jane",
				Age:     30,
				Email:   "jane@email.com",
				Default: "value2",
			},
		},
		{
			name:  "Struct with ignored field",
			input: "d4:name6:Ignore3:agei40e6:ignore10:ignoreThise",
			expected: TestStruct{
				Name: "Ignore",
				Age:  40,
			},
		},
		{
			name:  "Struct with nested struct",
			input: "d4:name4:John3:agei35e6:nestedd5:value10:nestedDataee",
			expected: TestStruct{
				Name: "John",
				Age:  35,
				Nested: NestedStruct{
					Value: "nestedData",
				},
			},
		},
		{
			name:  "Struct with integer slice",
			input: "d4:name3:Tom3:agei28e9:int_sliceli1ei2ei3eee",
			expected: TestStruct{
				Name:     "Tom",
				Age:      28,
				IntSlice: []int{1, 2, 3},
			},
		},
		{
			name:  "Struct with string map",
			input: "d4:name5:Sarah3:agei32e10:string_mapd3:key5:valuee13:email_address15:sarah@email.come",
			expected: TestStruct{
				Name:      "Sarah",
				Age:       32,
				Email:     "sarah@email.com",
				StringMap: map[string]string{"key": "value"},
			},
		},
		{
			name:     "Empty struct",
			input:    "de",
			expected: TestStruct{},
		},
		{
			name:  "Struct with zero values",
			input: "d4:name0:3:agei0e13:email_address0:e",
			expected: TestStruct{
				Name:  "",
				Age:   0,
				Email: "",
			},
		},
		{
			name:  "Struct with large integer",
			input: "d4:name3:Max3:agei2147483647ee",
			expected: TestStruct{
				Name: "Max",
				Age:  2147483647,
			},
		},
		{
			name:  "Struct with negative integer",
			input: "d4:name3:Min3:agei-100ee",
			expected: TestStruct{
				Name: "Min",
				Age:  -100,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader([]byte(tt.input))
			decoder := NewDecoder(reader)

			var result TestStruct
			err := decoder.Decode(&result)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %+v, got %+v", tt.expected, result)
			}
		})
	}
}
