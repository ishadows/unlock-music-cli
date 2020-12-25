package common

type Decoder interface {
	Validate() error
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
