package qmc

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"github.com/unlock-music/cli/algo/common"
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

func NewMflac256Decoder(data []byte) common.Decoder {
	return &Decoder{file: data, maskDetector: detectMflac256Mask, audioExt: "flac"}
}

func NewMgg256Decoder(data []byte) common.Decoder {
	return &Decoder{file: data, maskDetector: detectMgg256Mask, audioExt: "ogg"}
}

func (d *Decoder) Validate() error {
	if nil != d.mask {
		return nil
	}
	if nil != d.maskDetector {
		if err := d.validateKey(); err != nil {
			return err
		}
		var err error
		d.mask, err = d.maskDetector(d.file)
		return err
	}
	return errors.New("no mask or mask detector found")
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
	if d.audioExt != "" {
		return "." + d.audioExt
	}
	return ""
}

func (d Decoder) GetMeta() common.Meta {
	return nil
}

func DecoderFuncWithExt(ext string) common.NewDecoderFunc {
	return func(file []byte) common.Decoder {
		return &Decoder{file: file, audioExt: ext, mask: getDefaultMask()}
	}
}

//goland:noinspection SpellCheckingInspection
func init() {
	common.RegisterDecoder("qmc0", DecoderFuncWithExt("mp3")) //QQ Music Mp3
	common.RegisterDecoder("qmc3", DecoderFuncWithExt("mp3")) //QQ Music Mp3

	common.RegisterDecoder("qmc2", DecoderFuncWithExt("m4a")) //QQ Music M4A
	common.RegisterDecoder("qmc4", DecoderFuncWithExt("m4a")) //QQ Music M4A
	common.RegisterDecoder("qmc6", DecoderFuncWithExt("m4a")) //QQ Music M4A
	common.RegisterDecoder("qmc8", DecoderFuncWithExt("m4a")) //QQ Music M4A

	common.RegisterDecoder("qmcflac", DecoderFuncWithExt("flac")) //QQ Music Flac
	common.RegisterDecoder("qmcogg", DecoderFuncWithExt("ogg"))   //QQ Music Ogg
	common.RegisterDecoder("tkm", DecoderFuncWithExt("m4a"))      //QQ Music Accompaniment M4a

	common.RegisterDecoder("bkcmp3", DecoderFuncWithExt("mp3"))   //Moo Music Mp3
	common.RegisterDecoder("bkcflac", DecoderFuncWithExt("flac")) //Moo Music Flac

	common.RegisterDecoder("666c6163", DecoderFuncWithExt("flac")) //QQ Music Weiyun Flac
	common.RegisterDecoder("6d7033", DecoderFuncWithExt("mp3"))    //QQ Music Weiyun Mp3
	common.RegisterDecoder("6f6767", DecoderFuncWithExt("ogg"))    //QQ Music Weiyun Ogg
	common.RegisterDecoder("6d3461", DecoderFuncWithExt("m4a"))    //QQ Music Weiyun M4a
	common.RegisterDecoder("776176", DecoderFuncWithExt("wav"))    //QQ Music Weiyun Wav

	common.RegisterDecoder("mgg", NewMgg256Decoder)     //QQ Music New Ogg
	common.RegisterDecoder("mflac", NewMflac256Decoder) //QQ Music New Flac
}
