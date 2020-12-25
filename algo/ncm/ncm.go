package ncm

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/umlock-music/cli/algo/common"
	"github.com/umlock-music/cli/internal/logging"
	"github.com/umlock-music/cli/internal/utils"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strings"
)

var (
	magicHeader = []byte{
		0x43, 0x54, 0x45, 0x4E, 0x46, 0x44, 0x41, 0x4D}
	keyCore = []byte{
		0x68, 0x7a, 0x48, 0x52, 0x41, 0x6d, 0x73, 0x6f,
		0x35, 0x6b, 0x49, 0x6e, 0x62, 0x61, 0x78, 0x57}
	keyMeta = []byte{
		0x23, 0x31, 0x34, 0x6C, 0x6A, 0x6B, 0x5F, 0x21,
		0x5C, 0x5D, 0x26, 0x30, 0x55, 0x3C, 0x27, 0x28}
)

func NewDecoder(data []byte) *Decoder {
	return &Decoder{
		data:       data,
		fileLength: uint32(len(data)),
	}
}

type Decoder struct {
	data       []byte
	fileLength uint32

	key []byte
	box []byte

	metaRaw  []byte
	metaType string
	Meta     RawMeta

	Cover []byte
	Audio []byte

	offsetKey   uint32
	offsetMeta  uint32
	offsetCover uint32
	offsetAudio uint32
}

func (f *Decoder) Validate() bool {
	if !bytes.Equal(magicHeader, f.data[:len(magicHeader)]) {
		return false
	}

	/*if status.IsDebug {
		logging.Log().Info("the unknown field of the header is: \n" + spew.Sdump(f.data[8:10]))
	}*/
	f.offsetKey = 8 + 2
	return true
}

//todo: 读取前进行检查长度，防止越界
func (f *Decoder) readKeyData() error {
	if f.offsetKey == 0 || f.offsetKey+4 > f.fileLength {
		return errors.New("invalid cover data offset")
	}
	bKeyLen := f.data[f.offsetKey : f.offsetKey+4]
	iKeyLen := binary.LittleEndian.Uint32(bKeyLen)
	f.offsetMeta = f.offsetKey + 4 + iKeyLen

	bKeyRaw := make([]byte, iKeyLen)
	for i := uint32(0); i < iKeyLen; i++ {
		bKeyRaw[i] = f.data[i+4+f.offsetKey] ^ 0x64
	}

	f.key = utils.PKCS7UnPadding(utils.DecryptAes128Ecb(bKeyRaw, keyCore))[17:]
	return nil
}

func (f *Decoder) readMetaData() error {
	if f.offsetMeta == 0 || f.offsetMeta+4 > f.fileLength {
		return errors.New("invalid meta data offset")
	}
	bMetaLen := f.data[f.offsetMeta : f.offsetMeta+4]
	iMetaLen := binary.LittleEndian.Uint32(bMetaLen)
	f.offsetCover = f.offsetMeta + 4 + iMetaLen
	if iMetaLen == 0 {
		return errors.New("no any meta data found")
	}

	// Why sub 22: Remove "163 key(Don't modify):"
	bKeyRaw := make([]byte, iMetaLen-22)
	for i := uint32(0); i < iMetaLen-22; i++ {
		bKeyRaw[i] = f.data[f.offsetMeta+4+22+i] ^ 0x63
	}

	cipherText, err := base64.StdEncoding.DecodeString(string(bKeyRaw))
	if err != nil {
		return errors.New("decode ncm meta failed: " + err.Error())
	}
	metaRaw := utils.PKCS7UnPadding(utils.DecryptAes128Ecb(cipherText, keyMeta))
	sepIdx := bytes.IndexRune(metaRaw, ':')
	if sepIdx == -1 {
		return errors.New("invalid ncm meta data")
	}

	f.metaType = string(metaRaw[:sepIdx])
	f.metaRaw = metaRaw[sepIdx+1:]
	return nil
}

func (f *Decoder) buildKeyBox() {
	box := make([]byte, 256)
	for i := 0; i < 256; i++ {
		box[i] = byte(i)
	}

	keyLen := len(f.key)
	var j byte
	for i := 0; i < 256; i++ {
		j = box[i] + j + f.key[i%keyLen]
		box[i], box[j] = box[j], box[i]
	}

	f.box = make([]byte, 256)
	var _i byte
	for i := 0; i < 256; i++ {
		_i = byte(i + 1)
		si := box[_i]
		sj := box[_i+si]
		f.box[i] = box[si+sj]
	}
}

func (f *Decoder) parseMeta() error {
	switch f.metaType {
	case "music":
		f.Meta = new(RawMetaMusic)
		return json.Unmarshal(f.metaRaw, f.Meta)
	case "dj":
		f.Meta = new(RawMetaDJ)
	default:
		return errors.New("unknown ncm meta type: " + f.metaType)
	}
	return nil
}

func (f *Decoder) readCoverData() error {
	if f.offsetCover == 0 || f.offsetCover+13 > f.fileLength {
		return errors.New("invalid cover data offset")
	}

	coverLenStart := f.offsetCover + 5 + 4
	bCoverLen := f.data[coverLenStart : coverLenStart+4]

	/*if status.IsDebug {
		logging.Log().Info("the unknown field of the cover is: \n" +
			spew.Sdump(f.data[f.offsetCover:f.offsetCover+5]))
		coverLen2 := f.data[f.offsetCover+5 : f.offsetCover+5+4] // it seems that always the same
		if !bytes.Equal(coverLen2, bCoverLen) {
			logging.Log().Warn("special file found! 2 cover length filed no the same!")
		}
	}*/

	iCoverLen := binary.LittleEndian.Uint32(bCoverLen)
	f.offsetAudio = coverLenStart + 4 + iCoverLen
	if iCoverLen == 0 {
		return errors.New("no any cover data found")
	}
	f.Cover = f.data[coverLenStart+4 : 4+coverLenStart+iCoverLen]
	return nil
}

func (f *Decoder) readAudioData() error {
	if f.offsetAudio == 0 || f.offsetAudio > f.fileLength {
		return errors.New("invalid audio offset")
	}
	audioRaw := f.data[f.offsetAudio:]
	audioLen := len(audioRaw)
	f.Audio = make([]byte, audioLen)
	for i := uint32(0); i < uint32(audioLen); i++ {
		f.Audio[i] = f.box[i&0xff] ^ audioRaw[i]
	}
	return nil
}

func (f *Decoder) Decode() error {
	if err := f.readKeyData(); err != nil {
		return err
	}
	f.buildKeyBox()

	err := f.readMetaData()
	if err == nil {
		err = f.parseMeta()
	}
	if err != nil {
		logging.Log().Warn("parse ncm meta data failed", zap.Error(err))
	}

	err = f.readCoverData()
	if err != nil {
		logging.Log().Warn("parse ncm cover data failed", zap.Error(err))
	}

	return f.readAudioData()
}

func (f Decoder) GetAudioExt() string {
	if f.Meta != nil {
		return f.Meta.GetFormat()
	}
	return ""
}

func (f Decoder) GetAudioData() []byte {
	return f.Audio
}

func (f Decoder) GetCoverImage() []byte {
	if f.Cover != nil {
		return f.Cover
	}
	{
		imgURL := f.Meta.GetAlbumImageURL()
		if f.Meta != nil && !strings.HasPrefix(imgURL, "http") {
			return nil
		}
		resp, err := http.Get(imgURL)
		if err != nil {
			logging.Log().Warn("download image failed", zap.Error(err), zap.String("url", imgURL))
			return nil
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			logging.Log().Warn("download image failed", zap.String("http", resp.Status),
				zap.String("url", imgURL))
			return nil
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logging.Log().Warn("download image failed", zap.Error(err), zap.String("url", imgURL))
			return nil
		}
		return data
	}
}

func (f Decoder) GetMeta() common.Meta {
	return f.Meta
}
