package tm

import (
	"bytes"
	"errors"
	"github.com/umlock-music/cli/algo/common"
)

var magicHeader = []byte{0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70}

type Decoder struct {
	file []byte
}

func (d *Decoder) GetCoverImage() []byte {
	return nil
}

func (d *Decoder) GetAudioData() []byte {
	return d.file[8:]
}

func (d *Decoder) GetAudioExt() string {
	return ""
}

func (d *Decoder) GetMeta() common.Meta {
	return nil
}

func NewDecoder(data []byte) *Decoder {
	return &Decoder{file: data}
}

func (d *Decoder) Validate() error {
	if len(d.file) < 8 {
		return errors.New("invalid file size")
	}
	if !bytes.Equal(magicHeader, d.file[:8]) {
		return errors.New("not a valid tm file")
	}
	return nil
}

func (d *Decoder) Decode() error {
	return nil
}
