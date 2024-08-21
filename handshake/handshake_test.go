package handshake

import (
	"bytes"
	"reflect"
	"testing"
)

func TestHandshake_Deserialize(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    *Handshake
		wantErr bool
	}{
		{
			name: "Valid Handshake",
			input: []byte{
				byte(len(pstr)),                                                                               // pstrlen
				'B', 'i', 't', 'T', 'o', 'r', 'r', 'e', 'n', 't', ' ', 'p', 'r', 'o', 't', 'o', 'c', 'o', 'l', // pstr
				0, 0, 0, 0, 0, 0, 0, 0, // reserved
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, // InfoHash
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, // PeerId
			},
			want: &Handshake{
				Pstr:     pstr,
				InfoHash: [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19},
				PeerId:   [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19},
			},
			wantErr: false,
		},
		{
			name:    "Invalid Handshake - Empty Input",
			input:   []byte{},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid Handshake - Incorrect pstr Length",
			input: []byte{
				byte(20),
				'B', 'i', 't', 'T', 'o', 'r', 'r', 'e', 'n', 't', ' ', 'p', 'r', 'o', 't', 'o', 'c', 'o', 'l', 'X',
				0, 0, 0, 0, 0, 0, 0, 0, // reserved
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, // InfoHash
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, // PeerId
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Invalid Handshake - Incomplete Body",
			input: []byte{
				byte(len(pstr)),
				'B', 'i', 't', 'T', 'o', 'r', 'r', 'e', 'n', 't', ' ', 'p', 'r', 'o', 't', 'o', 'c', 'o', 'l',
				0, 0, 0, 0, 0, 0, 0, 0, // reserved
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, // incomplete InfoHash and PeerId
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Valid Handshake - Different InfoHash and PeerId",
			input: []byte{
				byte(len(pstr)),
				'B', 'i', 't', 'T', 'o', 'r', 'r', 'e', 'n', 't', ' ', 'p', 'r', 'o', 't', 'o', 'c', 'o', 'l',
				0, 0, 0, 0, 0, 0, 0, 0, // reserved
				10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, // InfoHash
				30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, // PeerId
			},
			want: &Handshake{
				Pstr:     pstr,
				InfoHash: [20]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29},
				PeerId:   [20]byte{30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49},
			},
			wantErr: false,
		},
		{
			name: "Invalid Handshake - Zero Length pstr",
			input: []byte{
				0,
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handshake{}
			got, err := h.Deserialize(bytes.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Deserialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Deserialize() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandshake_Serialize(t *testing.T) {
	tests := []struct {
		name   string
		fields *Handshake
		want   []byte
	}{
		{
			name: "Valid Handshake",
			fields: &Handshake{
				Pstr:     pstr,
				InfoHash: [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19},
				PeerId:   [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19},
			},
			want: []byte{
				byte(len(pstr)),                                                                               // pstrlen
				'B', 'i', 't', 'T', 'o', 'r', 'r', 'e', 'n', 't', ' ', 'p', 'r', 'o', 't', 'o', 'c', 'o', 'l', // pstr
				0, 0, 0, 0, 0, 0, 0, 0, // reserved
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, // InfoHash
				0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, // PeerId
			},
		},
		{
			name: "Different InfoHash and PeerId",
			fields: &Handshake{
				Pstr:     pstr,
				InfoHash: [20]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29},
				PeerId:   [20]byte{30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49},
			},
			want: []byte{
				byte(len(pstr)),                                                                               // pstrlen
				'B', 'i', 't', 'T', 'o', 'r', 'r', 'e', 'n', 't', ' ', 'p', 'r', 'o', 't', 'o', 'c', 'o', 'l', // pstr
				0, 0, 0, 0, 0, 0, 0, 0, // reserved
				10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, // InfoHash
				30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, // PeerId
			},
		},
		{
			name: "Empty InfoHash and PeerId",
			fields: &Handshake{
				Pstr:     pstr,
				InfoHash: [20]byte{},
				PeerId:   [20]byte{},
			},
			want: []byte{
				byte(len(pstr)),                                                                               // pstrlen
				'B', 'i', 't', 'T', 'o', 'r', 'r', 'e', 'n', 't', ' ', 'p', 'r', 'o', 't', 'o', 'c', 'o', 'l', // pstr
				0, 0, 0, 0, 0, 0, 0, 0, // reserved
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // InfoHash
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // PeerId
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fields.Serialize()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Serialize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewHandshake(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			peerId   [20]byte
			infoHash [20]byte
		}
		want *Handshake
	}{
		{
			name: "Valid Handshake",
			args: struct {
				peerId   [20]byte
				infoHash [20]byte
			}{
				peerId:   toByteArray("peer_id_1234567890"),
				infoHash: toByteArray("info_hash_12345678"),
			},
			want: &Handshake{
				Pstr:     pstr,
				PeerId:   toByteArray("peer_id_1234567890"),
				InfoHash: toByteArray("info_hash_12345678"),
			},
		},
		// Additional cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewHandshake(tt.args.peerId, tt.args.infoHash)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHandshake() = %v, want %v", got, tt.want)
			}
		})
	}
}

// helper function that converts a string to [20]byte
func toByteArray(s string) [20]byte {
	var arr [20]byte
	copy(arr[:], s)
	return arr
}
