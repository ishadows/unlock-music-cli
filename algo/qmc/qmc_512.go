package qmc

import (
	"encoding/binary"
	"errors"
	"io"
	"strconv"
	"strings"
)

type Mflac0Decoder struct {
	r io.ReadSeeker

	audioLen   int
	decodedKey []byte
	cipher     streamCipher
	offset     int

	rawMetaExtra1 int
	rawMetaExtra2 int
}

func (d *Mflac0Decoder) Read(p []byte) (int, error) {
	if d.cipher != nil {
		return d.readRC4(p)
	} else {
		panic("not impl")
		//return d.readPlain(p)
	}
}
func (d *Mflac0Decoder) readRC4(p []byte) (int, error) {
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

func NewMflac0Decoder(r io.ReadSeeker) (*Mflac0Decoder, error) {
	d := &Mflac0Decoder{r: r}
	err := d.searchKey()
	if err != nil {
		return nil, err
	}

	if len(d.decodedKey) == 0 {
		return nil, errors.New("invalid decoded key")
	} else if len(d.decodedKey) > 300 {
		d.cipher, err = NewRC4Cipher(d.decodedKey)
		if err != nil {
			return nil, err
		}
	} else {
		d.cipher, err = NewMapCipher(d.decodedKey)
		if err != nil {
			return nil, err
		}
	}

	_, err = d.r.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Mflac0Decoder) searchKey() error {
	if _, err := d.r.Seek(-4, io.SeekEnd); err != nil {
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
		if size < 0x300 {
			return d.readRawKey(int64(size))
		} else {
			// todo: try to use fixed key
			panic("not impl")
		}
	}
	return nil
}

func (d *Mflac0Decoder) readRawKey(rawKeyLen int64) error {
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

func (d *Mflac0Decoder) readRawMetaQTag() error {
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
