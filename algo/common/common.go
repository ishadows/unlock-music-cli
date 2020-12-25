package common

type Decoder interface {
	Validate() bool
	Decode() error
	GetCoverImage() []byte
	GetAudioData() []byte
	GetAudioExt() string
	GetMeta() Meta
}

type Meta interface {
	GetArtists() []string
	GetTitle() string
	GetAlbum() string
}
