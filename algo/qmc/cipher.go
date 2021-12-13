package qmc

type streamCipher interface {
	Decrypt(buf []byte, offset int)
}
