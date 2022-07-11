package pow

import (
	"reflect"
	"testing"
)

const date = 1657527541

func TestComputeHashcash(t *testing.T) {
	tests := []struct {
		name          string
		zeroCount     int
		clientInfo    string
		uid           string
		maxIterations int
		want          HashcashData
		wantErr       bool
	}{
		{
			name:          "#1: Should return correct value with counter 10855",
			zeroCount:     3,
			clientInfo:    "localhost",
			maxIterations: 1000000,
			uid:           "f672e568-dc45-409d-9d47-a6df76d5633b",
			want: HashcashData{
				Version:    DefaultVersion,
				ZerosCount: 3,
				Date:       date,
				Resource:   "localhost",
				Rand:       "ZjY3MmU1NjgtZGM0NS00MDlkLTlkNDctYTZkZjc2ZDU2MzNi",
				Counter:    10855,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHashcash(tt.zeroCount, tt.clientInfo, tt.uid, date)
			got, err := h.ComputeHashcash(tt.maxIterations)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComputeHashcash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ComputeHashcash() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringify(t *testing.T) {
	type fields struct {
		Version    int
		ZerosCount int
		Date       int64
		Resource   string
		Rand       string
		Counter    int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "#1: Should return correct stringify line",
			fields: fields{
				Version:    DefaultVersion,
				ZerosCount: 3,
				Date:       date,
				Resource:   "localhost",
				Rand:       "f672e568-dc45-409d-9d47-a6df76d5633b",
				Counter:    DefaultCounter,
			},
			want: "1:3:1657527541:localhost::f672e568-dc45-409d-9d47-a6df76d5633b:0",
		},
		{
			name: "#1: Should return correct stringify line",
			fields: fields{
				Version:    DefaultVersion,
				ZerosCount: 3,
				Date:       date,
				Resource:   "localhost",
				Rand:       "f672e568-dc45-409d-9d47-a6df76d5633b",
				Counter:    10855,
			},
			want: "1:3:1657527541:localhost::f672e568-dc45-409d-9d47-a6df76d5633b:10855",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := HashcashData{
				Version:    tt.fields.Version,
				ZerosCount: tt.fields.ZerosCount,
				Date:       tt.fields.Date,
				Resource:   tt.fields.Resource,
				Rand:       tt.fields.Rand,
				Counter:    tt.fields.Counter,
			}
			if got := h.Stringify(); got != tt.want {
				t.Errorf("Stringify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerify(t *testing.T) {
	type fields struct {
		Version    int
		ZerosCount int
		Date       int64
		Resource   string
		Rand       string
		Counter    int
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "#1: Should return true on verify hashcash",
			fields: fields{
				Version:    DefaultVersion,
				ZerosCount: 3,
				Date:       date,
				Resource:   "localhost",
				Rand:       "ZjY3MmU1NjgtZGM0NS00MDlkLTlkNDctYTZkZjc2ZDU2MzNi",
				Counter:    10855,
			},
			want: true,
		},
		{
			name: "#2: Should return false on verify hashcash",
			fields: fields{
				Version:    DefaultVersion,
				ZerosCount: 3,
				Date:       date,
				Resource:   "localhost",
				Rand:       "ZjY3MmU1NjgtZGM0NS00MDlkLTlkNDctYTZkZjc2ZDU2MzNi",
				Counter:    10000,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := HashcashData{
				Version:    tt.fields.Version,
				ZerosCount: tt.fields.ZerosCount,
				Date:       tt.fields.Date,
				Resource:   tt.fields.Resource,
				Rand:       tt.fields.Rand,
				Counter:    tt.fields.Counter,
			}
			if got := h.Verify(); got != tt.want {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsHashCorrect(t *testing.T) {
	tests := []struct {
		name      string
		hash      string
		zeroCount int
		want      bool
	}{
		{
			name:      "#1: Should return true on zeroCount=3",
			hash:      "00095374cb044a3e5826136d6f8defbfc91c448d",
			zeroCount: 3,
			want:      true,
		},
		{
			name:      "#2: Should return true on zeroCount=4",
			hash:      "00005374cb044a3e5826136d6f8defbfc91c448d",
			zeroCount: 4,
			want:      true,
		},
		{
			name:      "#2: Should return false on zeroCount=5",
			hash:      "00005374cb044a3e5826136d6f8defbfc91c448d",
			zeroCount: 5,
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsHashCorrect(tt.hash, tt.zeroCount); got != tt.want {
				t.Errorf("IsHashCorrect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewHashcash(t *testing.T) {
	tests := []struct {
		name       string
		zeroCount  int
		clientInfo string
		uid        string
		want       *HashcashData
	}{
		{
			name:       "#1: Should return correct value from constructor",
			zeroCount:  3,
			clientInfo: "localhost",
			uid:        "f672e568-dc45-409d-9d47-a6df76d5633b",
			want: &HashcashData{
				Version:    DefaultVersion,
				ZerosCount: 3,
				Date:       date,
				Resource:   "localhost",
				Rand:       "ZjY3MmU1NjgtZGM0NS00MDlkLTlkNDctYTZkZjc2ZDU2MzNi",
				Counter:    DefaultCounter,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHashcash(tt.zeroCount, tt.clientInfo, tt.uid, date); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHashcash() = %v, want %v", got, tt.want)
			}
		})
	}
}
