package common

import (
	"path/filepath"
	"strings"
)

type NewDecoderFunc func([]byte) Decoder

type decoderItem struct {
	noop    bool
	decoder NewDecoderFunc
}

var decoderRegistry = make(map[string][]decoderItem)

func RegisterDecoder(ext string, noop bool, dispatchFunc NewDecoderFunc) {
	decoderRegistry[ext] = append(decoderRegistry[ext],
		decoderItem{noop: noop, decoder: dispatchFunc})
}
func GetDecoder(filename string, skipNoop bool) (rs []NewDecoderFunc) {
	ext := strings.ToLower(strings.TrimLeft(filepath.Ext(filename), "."))
	for _, dec := range decoderRegistry[ext] {
		if skipNoop && dec.noop {
			continue
		}
		rs = append(rs, dec.decoder)
	}
	return
}
