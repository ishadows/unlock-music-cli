package common

import (
	"path/filepath"
	"strings"
)

type NewDecoderFunc func([]byte) Decoder

var decoderRegistry = make(map[string][]NewDecoderFunc)

func RegisterDecoder(ext string, dispatchFunc NewDecoderFunc) {
	decoderRegistry[ext] = append(decoderRegistry[ext], dispatchFunc)
}
func GetDecoder(filename string) []NewDecoderFunc {
	ext := strings.ToLower(strings.TrimLeft(filepath.Ext(filename), "."))
	return decoderRegistry[ext]
}
