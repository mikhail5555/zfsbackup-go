package compencrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

type EncryptionWriter cipher.StreamWriter

func NewEncryptionWriter(destination io.WriteCloser, key []byte) (io.WriteCloser, error) {
	if len(key) == 0 {
		return destination, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	if _, err := destination.Write(iv); err != nil {
		return nil, err
	}

	return cipher.StreamWriter{S: stream, W: destination}, nil
}

type DecryptionReader cipher.StreamReader

func NewDecryptionReader(source io.ReadCloser, key []byte) (io.ReadCloser, error) {
	if len(key) == 0 {
		return source, nil
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(source, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	return io.NopCloser(cipher.StreamReader{S: stream, R: source}), nil
}

func NopWriteCloser(w io.Writer) io.WriteCloser {
	return &nopWriteCloser{w}
}

type nopWriteCloser struct {
	io.Writer
}

func (nwc *nopWriteCloser) Close() error { return nil }
