package compencrypt

import (
	"io"
)

type CompressAndEncryptWriter io.WriteCloser

func NewCompressAndEncryptWriter(destination io.WriteCloser, key []byte) CompressAndEncryptWriter {
	return NewCompressionWriter(NewEncryptionWriter(destination, key))
}

type DecryptAndDecompressReader io.ReadCloser

func NewDecryptAndDecompressReader(source io.ReadCloser, key []byte) DecryptAndDecompressReader {
	return NewDecompressionReader(NewDecryptionReader(source, key))
}
