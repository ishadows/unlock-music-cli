package common

type RawDecoder struct {
	file     []byte
	audioExt string
}

//goland:noinspection GoUnusedExportedFunction
func NewRawDecoder(file []byte) Decoder {
	return &RawDecoder{file: file}
}

func (d RawDecoder) Validate() error {
	return nil
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
func DecoderFuncWithExt(ext string) NewDecoderFunc {
	return func(file []byte) Decoder {
		return &RawDecoder{file: file, audioExt: ext}
	}
}
func init() {
	/*RegisterDecoder("mp3", DecoderFuncWithExt("mp3"))
	RegisterDecoder("flac", DecoderFuncWithExt("flac"))
	RegisterDecoder("wav", DecoderFuncWithExt("wav"))
	RegisterDecoder("ogg", DecoderFuncWithExt("ogg"))
	RegisterDecoder("m4a", DecoderFuncWithExt("m4a"))*/
}
