package qmc

import (
	"errors"
)

// A rc4Cipher is an instance of RC4 using a particular key.
type rc4Cipher struct {
	box    []byte
	key    []byte
	hash   uint32
	boxTmp []byte
}

// NewRC4Cipher creates and returns a new rc4Cipher. The key argument should be the
// RC4 key, at least 1 byte and at most 256 bytes.
func NewRC4Cipher(key []byte) (*rc4Cipher, error) {
	n := len(key)
	if n == 0 {
		return nil, errors.New("crypto/rc4: invalid key size")
	}

	var c = rc4Cipher{key: key}
	c.box = make([]byte, n)
	c.boxTmp = make([]byte, n)

	for i := 0; i < n; i++ {
		c.box[i] = byte(i)
	}

	var j = 0
	for i := 0; i < n; i++ {
		j = (j + int(c.box[i]) + int(key[i%n])) % n
		c.box[i], c.box[j] = c.box[j], c.box[i]
	}
	c.getHashBase()
	return &c, nil
}

func (c *rc4Cipher) getHashBase() {
	c.hash = 1
	for i := 0; i < len(c.key); i++ {
		v := uint32(c.key[i])
		if v == 0 {
			continue
		}
		nextHash := c.hash * v
		if nextHash == 0 || nextHash <= c.hash {
			break
		}
		c.hash = nextHash
	}
}

const rc4SegmentSize = 5120

func (c *rc4Cipher) Process(src []byte, offset int) {
	toProcess := len(src)
	processed := 0
	markProcess := func(p int) (finished bool) {
		offset += p
		toProcess -= p
		processed += p
		return toProcess == 0
	}

	if offset < 128 {
		blockSize := toProcess
		if blockSize > 128-offset {
			blockSize = 128 - offset
		}
		c.encFirstSegment(src[:blockSize], offset)
		if markProcess(blockSize) {
			return
		}
	}

	if offset%rc4SegmentSize != 0 {
		blockSize := toProcess
		if blockSize > rc4SegmentSize-offset%rc4SegmentSize {
			blockSize = rc4SegmentSize - offset%rc4SegmentSize
		}
		k := src[processed : processed+blockSize]
		c.encASegment(k, offset)
		if markProcess(blockSize) {
			return
		}
	}
	for toProcess > rc4SegmentSize {
		c.encASegment(src[processed:processed+rc4SegmentSize], offset)
		markProcess(rc4SegmentSize)
	}

	if toProcess > 0 {
		c.encASegment(src[processed:], offset)
	}
}
func (c *rc4Cipher) encFirstSegment(buf []byte, offset int) {
	n := len(c.box)
	for i := 0; i < len(buf); i++ {
		idx1 := offset + i
		segmentID := int(c.key[idx1%n])
		idx2 := int(float64(c.hash) / float64((idx1+1)*segmentID) * 100.0)
		buf[i] ^= c.key[idx2%n]
	}
}

func (c *rc4Cipher) encASegment(buf []byte, offset int) {
	n := len(c.box)
	copy(c.boxTmp, c.box)

	segmentID := (offset / rc4SegmentSize) & 0x1FF

	if n <= segmentID {
		return
	}

	idx2 := int64(float64(c.hash) /
		float64((offset/rc4SegmentSize+1)*int(c.key[segmentID])) *
		100.0)
	skipLen := int((idx2 & 0x1FF) + int64(offset%rc4SegmentSize))

	j, k := 0, 0

	for i := -skipLen; i < len(buf); i++ {
		j = (j + 1) % n
		k = (int(c.boxTmp[j]) + k) % n
		c.boxTmp[j], c.boxTmp[k] = c.boxTmp[k], c.boxTmp[j]
		if i >= 0 {
			buf[i] ^= c.boxTmp[int(c.boxTmp[j])+int(c.boxTmp[k])%n]
		}
	}
}
