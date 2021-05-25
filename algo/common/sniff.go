package common

import "bytes"

type Sniffer func(header []byte) bool

var snifferRegistry = map[string]Sniffer{
	".mp3":  SnifferMP3,
	".flac": SnifferFLAC,
	".ogg":  SnifferOGG,
	".m4a":  SnifferM4A,
	".wav":  SnifferWAV,
	".wma":  SnifferWMA,
	".aac":  SnifferAAC,
	".dff":  SnifferDFF,
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
func SnifferAAC(header []byte) bool {
	return bytes.HasPrefix(header, []byte{0xFF, 0xF1})
}

// SnifferDFF sniff a DSDIFF format
// reference to: https://www.sonicstudio.com/pdf/dsd/DSDIFF_1.5_Spec.pdf
func SnifferDFF(header []byte) bool {
	return bytes.HasPrefix(header, []byte("FRM8"))
}
