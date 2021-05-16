package common

import "bytes"

type Sniffer func(header []byte) bool

var snifferRegistry = map[string]Sniffer{
	".m4a":  SnifferM4A,
	".ogg":  SnifferOGG,
	".flac": SnifferFLAC,
	".wav":  SnifferWAV,
	".wma":  SnifferWMA,
	".mp3":  SnifferMP3,
}

func SniffAll(header []byte) (string, bool) {
	for ext, sniffer := range snifferRegistry {
		if sniffer(header) {
			return ext, true
		}
	}
	return "", false
}

func SnifferM4A(header []byte) bool {
	return len(header) >= 8 && bytes.Equal([]byte("ftyp"), header[4:8])
}

func SnifferOGG(header []byte) bool {
	return bytes.HasPrefix(header, []byte("OggS"))
}

func SnifferFLAC(header []byte) bool {
	return bytes.HasPrefix(header, []byte("fLaC"))
}
func SnifferMP3(header []byte) bool {
	return bytes.HasPrefix(header, []byte("ID3"))
}
func SnifferWAV(header []byte) bool {
	return bytes.HasPrefix(header, []byte("RIFF"))
}
func SnifferWMA(header []byte) bool {
	return bytes.HasPrefix(header, []byte("\x30\x26\xb2\x75\x8e\x66\xcf\x11\xa6\xd9\x00\xaa\x00\x62\xce\x6c"))
}
