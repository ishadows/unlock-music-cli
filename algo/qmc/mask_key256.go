package qmc

import (
	"bytes"
	"errors"
	"github.com/unlock-music/cli/internal/logging"
	"go.uber.org/zap"
)

var (
	ErrFailToMatchMask = errors.New("can not match at least one key")
	ErrTestDataLength  = errors.New("invalid length of test file")
	ErrMaskLength128   = errors.New("incorrect mask length 128")
	ErrMaskLength44    = errors.New("incorrect mask length 44")
	ErrMaskDecode      = errors.New("decode mask-128 to mask-58 failed")
	ErrDetectFlacMask  = errors.New("can not detect mflac mask")
	ErrDetectMggMask   = errors.New("can not detect mgg mask")
)

type Key256Mask struct {
	matrix []byte // Mask 128
}

func NewKey256FromMask128(mask128 []byte) (*Key256Mask, error) {
	if len(mask128) != 128 {
		return nil, ErrMaskLength128
	}
	q := &Key256Mask{matrix: mask128}
	return q, nil
}

func NewKey256FromMask44(mask44 []byte) (*Key256Mask, error) {
	mask128, err := convertKey256Mask44to128(mask44)
	if err != nil {
		return nil, err
	}
	q := &Key256Mask{matrix: mask128}
	return q, nil
}

func (q *Key256Mask) getMatrix44() (mask44 []byte, err error) {
	if len(q.matrix) != 128 {
		return nil, ErrMaskLength128
	}
	matrix44 := make([]byte, 44)
	idx44 := 0
	for _, it256 := range key256MappingAll {
		if it256 != nil {
			it256Len := len(it256)
			for i := 1; i < it256Len; i++ {
				if q.matrix[it256[0]] != q.matrix[it256[i]] {
					return nil, ErrMaskDecode
				}
			}
			q.matrix[idx44] = q.matrix[it256[0]]
			idx44++
		}
	}
	return matrix44, nil
}

func (q *Key256Mask) Decrypt(data []byte) []byte {
	dst := make([]byte, len(data))
	index := -1
	maskIdx := -1
	for cur := 0; cur < len(data); cur++ {
		index++
		maskIdx++
		if index == 0x8000 || (index > 0x8000 && (index+1)%0x8000 == 0) {
			index++
			maskIdx++
		}
		if maskIdx >= 128 {
			maskIdx -= 128
		}
		dst[cur] = data[cur] ^ q.matrix[maskIdx]
	}
	return dst
}

func convertKey256Mask44to128(mask44 []byte) ([]byte, error) {
	if len(mask44) != 44 {
		return nil, ErrMaskLength44
	}
	mask128 := make([]byte, 128)
	idx44 := 0
	for _, it256 := range key256MappingAll {
		if it256 != nil {
			for _, idx128 := range it256 {
				mask128[idx128] = mask44[idx44]
			}
			idx44++
		}
	}
	return mask128, nil
}

func getDefaultMask() *Key256Mask {
	y, _ := NewKey256FromMask44(defaultKey256Mask44)
	return y
}

func detectMflac256Mask(input []byte) (*Key256Mask, error) {
	var q *Key256Mask
	var rtErr = ErrDetectFlacMask

	lenData := len(input)
	lenTest := 0x8000
	if lenData < 0x8000 {
		lenTest = lenData
	}

	for blockIdx := 0; blockIdx < lenTest; blockIdx += 128 {
		var err error
		q, err = NewKey256FromMask128(input[blockIdx : blockIdx+128])
		if err != nil {
			continue
		}
		if bytes.Equal(headerFlac, q.Decrypt(input[:len(headerFlac)])) {
			rtErr = nil
			break
		}
	}
	return q, rtErr
}

func detectMgg256Mask(input []byte) (*Key256Mask, error) {
	if len(input) < 0x100 {
		return nil, ErrTestDataLength
	}

	matrixConf := make([]map[uint8]uint, 44) //meaning: [idx58][value]confidence
	for i := uint(0); i < 44; i++ {
		matrixConf[i] = make(map[uint8]uint)
	}

	page2size := input[0x54] ^ input[0xC] ^ oggPublicHeader1[0xC]
	spHeader, spConf := generateOggFullHeader(int(page2size))
	lenTest := len(spHeader)

	for idx128 := 0; idx128 < lenTest; idx128++ {
		confidence := spConf[idx128]
		if confidence > 0 {
			mask := input[idx128] ^ spHeader[idx128]

			idx44 := key256Mapping128to44[idx128&0x7f] // equals: [idx128 % 128]
			if _, ok2 := matrixConf[idx44][mask]; ok2 {
				matrixConf[idx44][mask] += confidence
			} else {
				matrixConf[idx44][mask] = confidence
			}
		}
	}

	matrix := make([]uint8, 44)
	var err error
	for i := uint(0); i < 44; i++ {
		matrix[i], err = decideMgg256MaskItemConf(matrixConf[i])
		if err != nil {
			return nil, err
		}
	}
	q, err := NewKey256FromMask44(matrix)
	if err != nil {
		return nil, err
	}
	if bytes.Equal(headerOgg, q.Decrypt(input[:len(headerOgg)])) {
		return q, nil
	}
	return nil, ErrDetectMggMask
}

func generateOggFullHeader(pageSize int) ([]byte, []uint) {
	spec := make([]byte, pageSize+1)

	spec[0], spec[1], spec[pageSize] = uint8(pageSize), 0xFF, 0xFF
	for i := 2; i < pageSize; i++ {
		spec[i] = 0xFF
	}
	specConf := make([]uint, pageSize+1)
	specConf[0], specConf[1], specConf[pageSize] = 6, 0, 0
	for i := 2; i < pageSize; i++ {
		specConf[i] = 4
	}
	allConf := append(oggPublicConfidence1, specConf...)
	allConf = append(allConf, oggPublicConfidence2...)

	allHeader := bytes.Join(
		[][]byte{oggPublicHeader1, spec, oggPublicHeader2},
		[]byte{},
	)
	return allHeader, allConf
}

func decideMgg256MaskItemConf(confidence map[uint8]uint) (uint8, error) {
	lenConf := len(confidence)
	if lenConf == 0 {
		return 0xff, ErrFailToMatchMask
	} else if lenConf > 1 {
		logging.Log().Warn("there are 2 potential value for the mask", zap.Any("confidence", confidence))
	}
	result := uint8(0)
	conf := uint(0)
	for idx, item := range confidence {
		if item > conf {
			result = idx
			conf = item
		}
	}
	return result, nil
}
