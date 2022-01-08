package qmc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/unlock-music/cli/algo/common"
)

type Decoder struct {
	r       io.ReadSeeker
	fileExt string

	audioLen   int
	decodedKey []byte
	cipher     streamCipher
	offset     int

	rawMetaExtra1 int
	rawMetaExtra2 int
}

// Read implements io.Reader, offer the decrypted audio data.
// Validate should call before Read to check if the file is valid.
func (d *Decoder) Read(p []byte) (int, error) {
	n := len(p)
	if d.audioLen-d.offset <= 0 {
		return 0, io.EOF
	} else if d.audioLen-d.offset < n {
		n = d.audioLen - d.offset
	}
	m, err := d.r.Read(p[:n])
	if m == 0 {
		return 0, err
	}

	d.cipher.Decrypt(p[:m], d.offset)
	d.offset += m
	return m, err
}

func NewDecoder(r io.ReadSeeker) (*Decoder, error) {
	d := &Decoder{r: r}
	err := d.searchKey()
	if err != nil {
		return nil, err
	}

	if len(d.decodedKey) > 300 {
		d.cipher, err = NewRC4Cipher(d.decodedKey)
		if err != nil {
			return nil, err
		}
	} else if len(d.decodedKey) != 0 {
		d.cipher, err = NewMapCipher(d.decodedKey)
		if err != nil {
			return nil, err
		}
	} else {
		d.cipher = NewStaticCipher()
	}

	_, err = d.r.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Decoder) Validate() error {
	buf := make([]byte, 16)
	if _, err := io.ReadFull(d.r, buf); err != nil {
		return err
	}
	_, err := d.r.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	d.cipher.Decrypt(buf, 0)
	fileExt, ok := common.SniffAll(buf)
	if !ok {
		return errors.New("detect file type failed")
	}
	d.fileExt = fileExt
	return nil
}

func (d Decoder) GetFileExt() string {
	return d.fileExt
}

func (d *Decoder) searchKey() error {
	fileSizeM4, err := d.r.Seek(-4, io.SeekEnd)
	if err != nil {
		return err
	}
	buf, err := io.ReadAll(io.LimitReader(d.r, 4))
	if err != nil {
		return err
	}
	if string(buf) == "QTag" {
		if err := d.readRawMetaQTag(); err != nil {
			return err
		}
	} else {
		size := binary.LittleEndian.Uint32(buf)
		if size < 0x300 && size != 0 {
			return d.readRawKey(int64(size))
		} else {
			// try to use default static cipher
			d.audioLen = int(fileSizeM4 + 4)
			return nil
		}
	}
	return nil
}

func (d *Decoder) readRawKey(rawKeyLen int64) error {
	audioLen, err := d.r.Seek(-(4 + rawKeyLen), io.SeekEnd)
	if err != nil {
		return err
	}
	d.audioLen = int(audioLen)

	rawKeyData, err := io.ReadAll(io.LimitReader(d.r, rawKeyLen))
	if err != nil {
		return err
	}

	d.decodedKey, err = DecryptKey(rawKeyData)
	if err != nil {
		return err
	}

	return nil
}

func (d *Decoder) readRawMetaQTag() error {
	// get raw meta data len
	if _, err := d.r.Seek(-8, io.SeekEnd); err != nil {
		return err
	}
	buf, err := io.ReadAll(io.LimitReader(d.r, 4))
	if err != nil {
		return err
	}
	rawMetaLen := int64(binary.BigEndian.Uint32(buf))

	// read raw meta data
	audioLen, err := d.r.Seek(-(8 + rawMetaLen), io.SeekEnd)
	if err != nil {
		return err
	}
	d.audioLen = int(audioLen)
	rawMetaData, err := io.ReadAll(io.LimitReader(d.r, rawMetaLen))
	if err != nil {
		return err
	}

	items := strings.Split(string(rawMetaData), ",")
	if len(items) != 3 {
		return errors.New("invalid raw meta data")
	}

	d.decodedKey, err = DecryptKey([]byte(items[0]))
	if err != nil {
		return err
	}

	d.rawMetaExtra1, err = strconv.Atoi(items[1])
	if err != nil {
		return err
	}
	d.rawMetaExtra2, err = strconv.Atoi(items[2])
	if err != nil {
		return err
	}

	return nil
}

//goland:noinspection SpellCheckingInspection
func init() {
	supportedExts := []string{
		"qmc0", "qmc3", //QQ Music MP3
		"qmc2", "qmc4", "qmc6", "qmc8", //QQ Music M4A
		"qmcflac", //QQ Music FLAC
		"qmcogg",  //QQ Music OGG

		"tkm", //QQ Music Accompaniment M4A

		"bkcmp3", "bkcm4a", "bkcflac", "bkcwav", "bkcape", "bkcogg", "bkcwma", //Moo Music

		"666c6163", //QQ Music Weiyun Flac
		"6d7033",   //QQ Music Weiyun Mp3
		"6f6767",   //QQ Music Weiyun Ogg
		"6d3461",   //QQ Music Weiyun M4a
		"776176",   //QQ Music Weiyun Wav

		"mgg", "mgg1", "mggl", //QQ Music New Ogg
		"mflac", "mflac0", //QQ Music New Flac
	}
	for _, ext := range supportedExts {
		common.RegisterDecoder(ext, false, newCompactDecoder)
	}
}

type compactDecoder struct {
	decoder   *Decoder
	createErr error
	buf       *bytes.Buffer
}

func newCompactDecoder(p []byte) common.Decoder {
	r := bytes.NewReader(p)
	d, err := NewDecoder(r)
	c := compactDecoder{
		decoder:   d,
		createErr: err,
	}
	return &c
}

func (c *compactDecoder) Validate() error {
	if c.createErr != nil {
		return c.createErr
	}
	return c.decoder.Validate()
}

func (c *compactDecoder) Decode() error {
	if c.createErr != nil {
		return c.createErr
	}
	c.buf = bytes.NewBuffer(nil)
	_, err := io.Copy(c.buf, c.decoder)
	return err
}

func (c *compactDecoder) GetCoverImage() []byte {
	return nil
}

func (c *compactDecoder) GetAudioData() []byte {
	return c.buf.Bytes()
}

func (c *compactDecoder) GetAudioExt() string {
	if c.createErr != nil {
		return ""
	}
	return c.decoder.GetFileExt()
}

func (c *compactDecoder) GetMeta() common.Meta {
	return nil
}
