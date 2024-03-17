package bencode

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func TestEncoder_bencode(t *testing.T) {
	type fields struct {
		w io.Writer
	}
	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Bencode String",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				data: "test",
			},
			want:    []byte("4:test"),
			wantErr: false,
		},
		{
			name: "Bencode Integer",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				data: 42,
			},
			want:    []byte("i42e"),
			wantErr: false,
		},
		{
			name: "Bencode List",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				data: []interface{}{"one", 2, "three"},
			},
			want:    []byte("l3:onei2e5:threee"),
			wantErr: false,
		},
		{
			name: "Bencode Dictionary",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				data: map[string]interface{}{"key1": "value1", "key2": 42},
			},
			want:    []byte("d4:key16:value14:key2i42ee"),
			wantErr: false,
		},

		// Add more test cases as needed...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Encoder{
				w: tt.fields.w,
			}
			got, err := e.bencode(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("bencode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("bencode() got = %v, want %v\nstring(got) = %v", got, tt.want, string(got))
			}
		})
	}
}

func TestEncoder_encodeDict(t *testing.T) {
	type fields struct {
		w io.Writer
	}
	type args struct {
		data map[string]interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Encode Empty Dictionary",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				data: map[string]interface{}{},
			},
			want:    []byte("de"),
			wantErr: false,
		},
		{
			name: "Encode Dictionary with String and Integer",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				data: map[string]interface{}{"key1": "value1", "key2": 42},
			},
			want:    []byte("d4:key16:value14:key2i42ee"),
			wantErr: false,
		},
		// Add more test cases as needed...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Encoder{
				w: tt.fields.w,
			}
			got, err := e.encodeDict(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("encodeDict() error = %v, wantErr %v\n got string = %v", err, tt.wantErr, string(got))

				return
			}
			fmt.Printf("\ngot =\n %v, \nwant =\n %v\n", got, tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("encodeDict() got = %v, want %v\n got string = %v", got, tt.want, string(got))
			}
		})
	}
}

func TestEncoder_encodeInteger(t *testing.T) {
	type fields struct {
		w io.Writer
	}
	type args struct {
		data int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		{
			name: "Encode Positive Integer",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				data: 42,
			},
			want: []byte("i42e"),
		},
		{
			name: "Encode Negative Integer",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				data: -123,
			},
			want: []byte("i-123e"),
		},
		{
			name: "Encode Zero",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				data: 0,
			},
			want: []byte("i0e"),
		},
		// Add more test cases as needed...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Encoder{
				w: tt.fields.w,
			}
			if got := e.encodeInteger(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("encodeInteger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncoder_encodeList(t *testing.T) {
	type fields struct {
		w io.Writer
	}
	type args struct {
		data []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "Encode Empty List",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				data: []interface{}{},
			},
			want:    []byte("le"),
			wantErr: false,
		},
		{
			name: "Encode List with String, Integer, and List",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				data: []interface{}{"one", 2, []interface{}{"nested", "list"}},
			},
			want:    []byte("l3:onei2el6:nested4:listee"),
			wantErr: false,
		},
		// Add more test cases as needed...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Encoder{
				w: tt.fields.w,
			}
			got, err := e.encodeList(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("encodeList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("encodeList() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncoder_encodeString(t *testing.T) {
	type fields struct {
		w io.Writer
	}
	type args struct {
		data string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []byte
	}{
		{
			name: "Encode Empty String",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				data: "",
			},
			want: []byte("0:"),
		},
		{
			name: "Encode Non-Empty String",
			fields: fields{
				w: &bytes.Buffer{},
			},
			args: args{
				data: "hello",
			},
			want: []byte("5:hello"),
		},
		// Add more test cases as needed...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Encoder{
				w: tt.fields.w,
			}
			got := e.encodeString(tt.args.data)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("encodeString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncoderDecoder(t *testing.T) {
	buf := new(bytes.Buffer)

	encoder := NewEncoder(buf)
	err := encoder.Encode("hello")
	if err != nil {
		t.Fatalf("Error encoding: %v", err)
	}

	decoder := NewDecoder(bufio.NewReader(buf))
	got, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Error decoding: %v", err)
	}

	if !reflect.DeepEqual(got, "hello") {
		t.Errorf("got = %v, want %v", got, "hello")
	}
}
