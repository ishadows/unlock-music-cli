package kwm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/unlock-music/cli/algo/common"
	"strconv"
	"strings"
	"unicode"
)

var (
	magicHeader = []byte{
		0x79, 0x65, 0x65, 0x6C, 0x69, 0x6F, 0x6E, 0x2D,
		0x6B, 0x75, 0x77, 0x6F, 0x2D, 0x74, 0x6D, 0x65}
	ErrKwFileSize    = errors.New("kwm invalid file size")
	ErrKwMagicHeader = errors.New("kwm magic header not matched")
)

const keyPreDefined = "MoOtOiTvINGwd2E6n0E1i7L5t2IoOoNk"

type Decoder struct {
	file []byte

	key       []byte
	outputExt string
	bitrate   int
	mask      []byte

	audio []byte
}

func (d *Decoder) GetCoverImage() []byte {
	return nil
}

func (d *Decoder) GetAudioData() []byte {
	return d.audio
}

func (d *Decoder) GetAudioExt() string {
	return "." + d.outputExt
}

func (d *Decoder) GetMeta() common.Meta {
	return nil
}

func NewDecoder(data []byte) common.Decoder {
	//todo: Notice the input data will be changed for now
	return &Decoder{file: data}
}

func (d *Decoder) Validate() error {
	lenData := len(d.file)
	if lenData < 1024 {
		return ErrKwFileSize
	}
	if !bytes.Equal(magicHeader, d.file[:16]) {
		return ErrKwMagicHeader
	}

	return nil
}

func generateMask(key []byte) []byte {
	keyInt := binary.LittleEndian.Uint64(key)
	keyStr := strconv.FormatUint(keyInt, 10)
	keyStrTrim := padOrTruncate(keyStr, 32)
	mask := make([]byte, 32)
	for i := 0; i < 32; i++ {
		mask[i] = keyPreDefined[i] ^ keyStrTrim[i]
	}
	return mask
}

func (d *Decoder) parseBitrateAndType() {
	bitType := string(bytes.TrimRight(d.file[0x30:0x38], string(byte(0))))
	charPos := 0
	for charPos = range bitType {
		if !unicode.IsNumber(rune(bitType[charPos])) {
			break
		}
	}
	var err error
	d.bitrate, err = strconv.Atoi(bitType[:charPos])
	if err != nil {
		d.bitrate = 0
	}
	d.outputExt = strings.ToLower(bitType[charPos:])

}

func (d *Decoder) Decode() error {
	d.parseBitrateAndType()

	d.mask = generateMask(d.file[0x18:0x20])

	d.audio = d.file[1024:]
	dataLen := len(d.audio)
	for i := 0; i < dataLen; i++ {
		d.audio[i] ^= d.mask[i&0x1F] //equals: [i % 32]
	}
	return nil
}

func padOrTruncate(raw string, length int) string {
	lenRaw := len(raw)
	out := raw
	if lenRaw == 0 {
		out = string(make([]byte, length))
	} else if lenRaw > length {
		out = raw[:length]
	} else if lenRaw < length {
		_tmp := make([]byte, 32)
		for i := 0; i < 32; i++ {
			_tmp[i] = raw[i%lenRaw]
		}
		out = string(_tmp)
	}
	return out
}

func init() {
	// Kuwo Mp3/Flac
	common.RegisterDecoder("kwm", false, NewDecoder)
	common.RegisterDecoder("kwm", false, common.NewRawDecoder)
}
