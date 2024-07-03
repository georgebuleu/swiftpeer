package bencode

import (
	"bytes"
	"testing"
)

func TestEncoder_Encode(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		want    string
		wantErr bool
	}{
		{
			name: "Encode String",
			data: "test",
			want: "4:test",
		},
		{
			name: "Encode Integer",
			data: 42,
			want: "i42e",
		},
		{
			name: "Encode List",
			data: []interface{}{"one", 2, "three"},
			want: "l3:onei2e5:threee",
		},
		{
			name: "Encode Dictionary",
			data: map[string]interface{}{"key1": "value1", "key2": 42},
			want: "d4:key16:value14:key2i42ee",
		},
		{
			name: "Encode Empty Dictionary",
			data: map[string]interface{}{},
			want: "de",
		},
		{
			name: "Encode Empty List",
			data: []interface{}{},
			want: "le",
		},
		{
			name: "Encode Nested Structures",
			data: map[string]interface{}{
				"list": []interface{}{"hello", 5, []interface{}{1, "world"}},
				"num":  42,
			},
			want: "d4:listl5:helloi5eli1e5:worldee3:numi42ee",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			encoder := NewEncoder(buf)
			err := encoder.Encode(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encoder.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("Encoder.Encode() got = %v, want %v", got, tt.want)
			}
		})
	}
}
