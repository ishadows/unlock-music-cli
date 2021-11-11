package xm

import (
	"bytes"
	"errors"
	"github.com/unlock-music/cli/algo/common"
	"github.com/unlock-music/cli/internal/logging"
	"go.uber.org/zap"
)

var (
	magicHeader  = []byte{'i', 'f', 'm', 't'}
	magicHeader2 = []byte{0xfe, 0xfe, 0xfe, 0xfe}
	typeMapping  = map[string]string{
		" WAV": "wav",
		"FLAC": "flac",
		" MP3": "mp3",
		" A4M": "m4a",
	}
	ErrFileSize    = errors.New("xm invalid file size")
	ErrMagicHeader = errors.New("xm magic header not matched")
)

type Decoder struct {
	file      []byte
	headerLen uint32
	outputExt string
	mask      byte
	audio     []byte
}

func (d *Decoder) GetCoverImage() []byte {
	return nil
}

func (d *Decoder) GetAudioData() []byte {
	return d.audio
}

func (d *Decoder) GetAudioExt() string {
	if d.outputExt != "" {
		return "." + d.outputExt

	}
	return ""
}

func (d *Decoder) GetMeta() common.Meta {
	return nil
}

func NewDecoder(data []byte) common.Decoder {
	return &Decoder{file: data}
}

func (d *Decoder) Validate() error {
	lenData := len(d.file)
	if lenData < 16 {
		return ErrFileSize
	}
	if !bytes.Equal(magicHeader, d.file[:4]) ||
		!bytes.Equal(magicHeader2, d.file[8:12]) {
		return ErrMagicHeader
	}

	var ok bool
	d.outputExt, ok = typeMapping[string(d.file[4:8])]
	if !ok {
		return errors.New("detect unknown xm file type: " + string(d.file[4:8]))
	}

	if d.file[14] != 0 {
		logging.Log().Warn("not a simple xm file", zap.Uint8("b[14]", d.file[14]))
	}
	d.headerLen = uint32(d.file[12]) | uint32(d.file[13])<<8 | uint32(d.file[14])<<16 // LittleEndian Unit24
	if d.headerLen+16 > uint32(lenData) {
		return ErrFileSize
	}
	return nil
}

func (d *Decoder) Decode() error {
	d.mask = d.file[15]
	d.audio = d.file[16:]
	dataLen := uint32(len(d.audio))
	for i := d.headerLen; i < dataLen; i++ {
		d.audio[i] = ^(d.audio[i] - d.mask)
	}
	return nil
}

func DecoderFuncWithExt(ext string) common.NewDecoderFunc {
	return func(file []byte) common.Decoder {
		return &Decoder{file: file, outputExt: ext}
	}
}

func init() {
	// Xiami Wav/M4a/Mp3/Flac
	common.RegisterDecoder("xm", false, NewDecoder)
	// Xiami Typed Format
	common.RegisterDecoder("wav", false, DecoderFuncWithExt("wav"))
	common.RegisterDecoder("mp3", false, DecoderFuncWithExt("mp3"))
	common.RegisterDecoder("flac", false, DecoderFuncWithExt("flac"))
	common.RegisterDecoder("m4a", false, DecoderFuncWithExt("m4a"))
}
