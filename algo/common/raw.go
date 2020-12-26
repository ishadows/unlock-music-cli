package common

type RawDecoder struct {
	file []byte
}

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
	return ""
}

func (d RawDecoder) GetMeta() Meta {
	return nil
}

func init() {
	RegisterDecoder("mp3", NewRawDecoder)
	RegisterDecoder("flac", NewRawDecoder)
	RegisterDecoder("wav", NewRawDecoder)
	RegisterDecoder("ogg", NewRawDecoder)
	RegisterDecoder("m4a", NewRawDecoder)
}
