package parser

import (
	"reflect"
	"testing"
)

func TestParseFile(t *testing.T) {
	tests := []struct {
		name    string
		want    Metadata
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFile()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseInfo(t *testing.T) {
	tests := []struct {
		name     string
		wantInfo map[string]interface{}
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			wantInfo: map[string]interface{}{
				"announce": "https://torrent.ubuntu.com/announce",
				"info": map[string]interface{}{
					"length":       282413312,
					"name":         "ubuntu-20.04.3-desktop-amd64.iso",
					"piece length": 262144,
					"pieces":       "A string of 20-byte SHA1 hash values",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInfo, err := ParseInfo()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotInfo, tt.wantInfo) {
				t.Errorf("ParseInfo() gotInfo = %v, want %v", gotInfo, tt.wantInfo)
			}
		})
	}
}
