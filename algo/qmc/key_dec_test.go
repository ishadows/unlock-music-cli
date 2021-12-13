package qmc

import (
	"os"
	"reflect"
	"testing"
)

func TestSimpleMakeKey(t *testing.T) {
	expect := []byte{0x69, 0x56, 0x46, 0x38, 0x2b, 0x20, 0x15, 0x0b}
	t.Run("106,8", func(t *testing.T) {
		if got := simpleMakeKey(106, 8); !reflect.DeepEqual(got, expect) {
			t.Errorf("simpleMakeKey() = %v, want %v", got, expect)
		}
	})
}

func TestDecryptKey(t *testing.T) {
	rc4Raw, err := os.ReadFile("./testdata/rc4_key_raw.bin")
	if err != nil {
		t.Error(err)
	}
	rc4Dec, err := os.ReadFile("./testdata/rc4_key.bin")
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name    string
		rawKey  []byte
		want    []byte
		wantErr bool
	}{
		{
			"512",
			rc4Raw,
			rc4Dec,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecryptKey(tt.rawKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecryptKey() got = %v..., want %v...", string(got[:32]), string(tt.want[:32]))
			}
		})
	}
}
