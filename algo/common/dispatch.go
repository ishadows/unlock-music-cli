package common

type NewDecoderFunc func([]byte) Decoder

var decoderRegistry = make(map[string][]NewDecoderFunc)

func RegisterDecoder(ext string, dispatchFunc NewDecoderFunc) {
	decoderRegistry[ext] = append(decoderRegistry[ext], dispatchFunc)
}
func GetDecoder(ext string) []NewDecoderFunc {
	return decoderRegistry[ext]
}
