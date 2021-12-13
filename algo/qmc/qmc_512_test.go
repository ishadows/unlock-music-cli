package qmc

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
)

func loadTestDataQmcDecoder(filename string) ([]byte, []byte, error) {
	encBody, err := os.ReadFile(fmt.Sprintf("./testdata/%s_raw.bin", filename))
	if err != nil {
		return nil, nil, err
	}
	encSuffix, err := os.ReadFile(fmt.Sprintf("./testdata/%s_suffix.bin", filename))
	if err != nil {
		return nil, nil, err
	}

	target, err := os.ReadFile(fmt.Sprintf("./testdata/%s_target.bin", filename))
	if err != nil {
		return nil, nil, err
	}
	return bytes.Join([][]byte{encBody, encSuffix}, nil), target, nil

}
func TestMflac0Decoder_Read(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{"mflac0_rc4(512)", "mflac0_rc4", false},
		{"mflac_map(256)", "mflac_map", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, target, err := loadTestDataQmcDecoder(tt.filename)
			if err != nil {
				t.Fatal(err)
			}

			d, err := NewMflac0Decoder(bytes.NewReader(raw))
			if err != nil {
				t.Error(err)
			}
			buf := make([]byte, len(target))
			if _, err := io.ReadFull(d, buf); err != nil {
				t.Errorf("read bytes from decoder error = %v", err)
				return
			}
			if !reflect.DeepEqual(buf, target) {
				t.Errorf("Decrypt() got = %v, want %v", buf[:32], target[:32])
			}
		})
	}

}
