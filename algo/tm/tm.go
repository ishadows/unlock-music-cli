package tm

import (
	"bytes"
	"errors"
	"github.com/unlock-music/cli/algo/common"
)

var replaceHeader = []byte{0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70}
var magicHeader = []byte{0x51, 0x51, 0x4D, 0x55} //0x15, 0x1D, 0x1A, 0x21

type Decoder struct {
	file        []byte
	audio       []byte
	headerMatch bool
	audioExt    string
}

func (d *Decoder) GetCoverImage() []byte {
	return nil
}

func (d *Decoder) GetAudioData() []byte {
	return d.audio
}

func (d *Decoder) GetAudioExt() string {
	if d.audioExt != "" {
		return "." + d.audioExt
	}
	return ""
}

func (d *Decoder) GetMeta() common.Meta {
	return nil
}

func (d *Decoder) Validate() error {
	if len(d.file) < 8 {
		return errors.New("invalid file size")
	}
	if !bytes.Equal(magicHeader, d.file[:4]) {
		return errors.New("not a valid tm file")
	}
	d.headerMatch = true
	return nil
}

func (d *Decoder) Decode() error {
	d.audio = d.file
	if d.headerMatch {
		for i := 0; i < 8; i++ {
			d.audio[i] = replaceHeader[i]
		}
		d.audioExt = "m4a"
	}
	return nil
}

//goland:noinspection GoUnusedExportedFunction
func NewDecoder(data []byte) common.Decoder {
	return &Decoder{file: data}
}

func DecoderFuncWithExt(ext string) common.NewDecoderFunc {
	return func(file []byte) common.Decoder {
		return &Decoder{file: file, audioExt: ext}
	}
}

func init() {
	// QQ Music IOS M4a
	common.RegisterDecoder("tm2", false, DecoderFuncWithExt("m4a"))
	common.RegisterDecoder("tm6", false, DecoderFuncWithExt("m4a"))
	// QQ Music IOS Mp3
	common.RegisterDecoder("tm0", false, common.NewRawDecoder)
	common.RegisterDecoder("tm3", false, common.NewRawDecoder)

}
