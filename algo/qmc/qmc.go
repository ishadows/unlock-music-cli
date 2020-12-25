package qmc

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"github.com/umlock-music/cli/algo/common"
	"github.com/umlock-music/cli/internal/logging"
	"go.uber.org/zap"
)

var (
	ErrQmcFileLength      = errors.New("invalid qmc file length")
	ErrQmcKeyDecodeFailed = errors.New("base64 decode qmc key failed")
	ErrQmcKeyLength       = errors.New("unexpected decoded qmc key length")
)

type Decoder struct {
	file         []byte
	maskDetector func(encodedData []byte) (*Key256Mask, error)
	mask         *Key256Mask
	audioExt     string
	key          []byte
	audio        []byte
}

func NewDefaultDecoder(data []byte) *Decoder {
	return &Decoder{file: data, mask: getDefaultMask()}
}

func NewMflac256Decoder(data []byte) *Decoder {
	return &Decoder{file: data, maskDetector: detectMflac256Mask, audioExt: "flac"}
}

func NewMgg256Decoder(data []byte) *Decoder {
	return &Decoder{file: data, maskDetector: detectMgg256Mask, audioExt: "ogg"}
}

func (d *Decoder) Validate() bool {
	if nil != d.mask {
		return true
	}
	if nil != d.maskDetector {
		if err := d.validateKey(); err != nil {
			logging.Log().Error("detect file failed", zap.Error(err))
			return false
		}
		d.mask, _ = d.maskDetector(d.file)
	}
	return d.mask != nil
}

func (d *Decoder) validateKey() error {
	lenData := len(d.file)
	if lenData < 4 {
		return ErrQmcFileLength
	}

	keyLen := binary.LittleEndian.Uint32(d.file[lenData-4:])
	if lenData < int(keyLen+4) {
		return ErrQmcFileLength
	}
	var err error
	d.key, err = base64.StdEncoding.DecodeString(
		string(d.file[lenData-4-int(keyLen) : lenData-4]))
	if err != nil {
		return ErrQmcKeyDecodeFailed
	}

	if len(d.key) != 272 {
		return ErrQmcKeyLength
	}
	d.file = d.file[:lenData-4-int(keyLen)]
	return nil

}

func (d *Decoder) Decode() error {
	d.audio = d.mask.Decrypt(d.file)
	return nil
}

func (d Decoder) GetCoverImage() []byte {
	return nil
}

func (d Decoder) GetAudioData() []byte {
	return d.audio
}

func (d Decoder) GetAudioExt() string {
	return d.audioExt
}

func (d Decoder) GetMeta() common.Meta {
	return nil
}
