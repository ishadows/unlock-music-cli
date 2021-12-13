package qmc

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"testing"
)

func loadTestDataRC4Mflac0() ([]byte, []byte, error) {
	encBody, err := os.ReadFile("./testdata/rc4_raw.bin")
	if err != nil {
		return nil, nil, err
	}
	encSuffix, err := os.ReadFile("./testdata/rc4_suffix_mflac0.bin")
	if err != nil {
		return nil, nil, err
	}

	target, err := os.ReadFile("./testdata/rc4_target.bin")
	if err != nil {
		return nil, nil, err
	}
	return bytes.Join([][]byte{encBody, encSuffix}, nil), target, nil

}
func TestMflac0Decoder_Read(t *testing.T) {
	raw, target, err := loadTestDataRC4Mflac0()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("mflac0-file", func(t *testing.T) {
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
			t.Errorf("Process() got = %v, want %v", buf[:32], target[:32])
		}
	})
}
