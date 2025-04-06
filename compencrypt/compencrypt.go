package compencrypt

import (
	"io"
)

type CompressAndEncryptWriter io.WriteCloser

func NewCompressAndEncryptWriter(destination io.WriteCloser, key []byte) (CompressAndEncryptWriter, error) {
	encryptionWriter, err := NewEncryptionWriter(destination, key)
	if err != nil {
		return nil, err
	}
	return NewCompressionWriter(encryptionWriter)
}

type DecryptAndDecompressReader io.ReadCloser

func NewDecryptAndDecompressReader(source io.ReadCloser, key []byte) (DecryptAndDecompressReader, error) {
	decryptionReader, err := NewDecryptionReader(source, key)
	if err != nil {
		return nil, err
	}
	return NewDecompressionReader(decryptionReader)
}
