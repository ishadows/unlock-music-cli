package kgm

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/unlock-music/cli/algo/common"
	"github.com/unlock-music/cli/internal/logging"
)

var (
	vprHeader = []byte{
		0x05, 0x28, 0xBC, 0x96, 0xE9, 0xE4, 0x5A, 0x43,
		0x91, 0xAA, 0xBD, 0xD0, 0x7A, 0xF5, 0x36, 0x31}
	kgmHeader = []byte{
		0x7C, 0xD5, 0x32, 0xEB, 0x86, 0x02, 0x7F, 0x4B,
		0xA8, 0xAF, 0xA6, 0x8E, 0x0F, 0xFF, 0x99, 0x14}
	ErrKgmMagicHeader = errors.New("kgm/vpr magic header not matched")
)

type Decoder struct {
	file  []byte
	key   []byte
	isVpr bool
	audio []byte
}

func NewDecoder(file []byte) common.Decoder {
	return &Decoder{
		file: file,
	}
}

func (d Decoder) GetCoverImage() []byte {
	return nil
}

func (d Decoder) GetAudioData() []byte {
	return d.audio
}

func (d Decoder) GetAudioExt() string {
	return "" // use sniffer
}

func (d Decoder) GetMeta() common.Meta {
	return nil
}

func (d *Decoder) Validate() error {
	if bytes.Equal(kgmHeader, d.file[:len(kgmHeader)]) {
		d.isVpr = false
	} else if bytes.Equal(vprHeader, d.file[:len(vprHeader)]) {
		d.isVpr = true
	} else {
		return ErrKgmMagicHeader
	}

	d.key = d.file[0x1c:0x2c]
	d.key = append(d.key, 0x00)
	_ = d.file[0x2c:0x3c] //todo: key2
	return nil
}

func (d *Decoder) Decode() error {
	headerLen := binary.LittleEndian.Uint32(d.file[0x10:0x14])
	dataEncrypted := d.file[headerLen:]
	lenData := len(dataEncrypted)
	initMask()
	if fullMaskLen < lenData {
		logging.Log().Warn("The file is too large and the processed audio is incomplete, " +
			"please report to us about this file at https://github.com/unlock-music/cli/issues")
		lenData = fullMaskLen
	}
	d.audio = make([]byte, lenData)

	for i := 0; i < lenData; i++ {
		med8 := dataEncrypted[i] ^ d.key[i%17] ^ maskV2PreDef[i%(16*17)] ^ maskV2[i>>4]
		d.audio[i] = med8 ^ (med8&0xf)<<4
	}
	if d.isVpr {
		for i := 0; i < lenData; i++ {
			d.audio[i] ^= maskDiffVpr[i%17]
		}
	}
	return nil
}
func init() {
	// Kugou
	common.RegisterDecoder("kgm", false, NewDecoder)
	common.RegisterDecoder("kgma", false, NewDecoder)
	// Viper
	common.RegisterDecoder("vpr", false, NewDecoder)
}
