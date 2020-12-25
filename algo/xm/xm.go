package xm

import (
	"bytes"
	"errors"
	"github.com/umlock-music/cli/algo/common"
	"github.com/umlock-music/cli/internal/logging"
	"go.uber.org/zap"
)

var (
	xmMagicHeader    = []byte{'i', 'f', 'm', 't'}
	xmMagicHeader2   = []byte{0xfe, 0xfe, 0xfe, 0xfe}
	xmHeaders        map[string]string
	ErrXmFileSize    = errors.New("xm invalid file size")
	ErrXmMagicHeader = errors.New("xm magic header not matched")
)

func init() {
	xmHeaders = map[string]string{
		" WAV": "wav",
		"FLAC": "flac",
		" MP3": "mp3",
		" A4M": "m4a",
	}
}

type Decoder struct {
	data      []byte
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
	return d.outputExt
}

func (d *Decoder) GetMeta() common.Meta {
	return nil
}

func NewDecoder(data []byte) *Decoder {
	return &Decoder{data: data}
}

func (d *Decoder) Validate() error {
	lenData := len(d.data)
	if lenData < 16 {
		return ErrXmFileSize
	}
	if !bytes.Equal(xmMagicHeader, d.data[:4]) ||
		!bytes.Equal(xmMagicHeader2, d.data[8:12]) {
		return ErrXmMagicHeader
	}

	var ok bool
	d.outputExt, ok = xmHeaders[string(d.data[4:8])]
	if !ok {
		return errors.New("detect unknown xm file type: " + string(d.data[4:8]))
	}

	if d.data[14] != 0 {
		logging.Log().Warn("not a simple xm file", zap.Uint8("b[14]", d.data[14]))
	}
	d.headerLen = uint32(d.data[12]) | uint32(d.data[13])<<8 | uint32(d.data[14])<<16 // LittleEndian Unit24
	if d.headerLen+16 > uint32(lenData) {
		return ErrXmFileSize
	}
	return nil
}

func (d *Decoder) Decode() error {
	d.mask = d.data[15]
	d.audio = d.data[16:]
	dataLen := uint32(len(d.audio))
	for i := d.headerLen; i < dataLen; i++ {
		d.audio[i] = ^(d.audio[i] - d.mask)
	}
	return nil
}
