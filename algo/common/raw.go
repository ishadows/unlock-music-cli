package common

import (
	"errors"
	"strings"
)

type RawDecoder struct {
	file     []byte
	audioExt string
}

func NewRawDecoder(file []byte) Decoder {
	return &RawDecoder{file: file}
}

func (d *RawDecoder) Validate() error {
	for ext, sniffer := range snifferRegistry {
		if sniffer(d.file) {
			d.audioExt = strings.ToLower(ext)
			return nil
		}
	}
	return errors.New("audio doesn't recognized")
}

func (d RawDecoder) Decode() error {
	return nil
}

func (d RawDecoder) GetCoverImage() []byte {
	return nil
}

func (d RawDecoder) GetAudioData() []byte {
	return d.file
}

func (d RawDecoder) GetAudioExt() string {
	return d.audioExt
}

func (d RawDecoder) GetMeta() Meta {
	return nil
}

func init() {
	RegisterDecoder("mp3", true, NewRawDecoder)
	RegisterDecoder("flac", true, NewRawDecoder)
	RegisterDecoder("ogg", true, NewRawDecoder)
	RegisterDecoder("m4a", true, NewRawDecoder)
	RegisterDecoder("wav", true, NewRawDecoder)
	RegisterDecoder("wma", true, NewRawDecoder)
	RegisterDecoder("aac", true, NewRawDecoder)
}
