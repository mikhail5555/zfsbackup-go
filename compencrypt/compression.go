package compencrypt

import (
	"compress/gzip"
	"io"
)

type CompressionWriter gzip.Writer

func NewCompressionWriter(destination io.WriteCloser) (io.WriteCloser, error) {
	return gzip.NewWriter(destination), nil
}

type DecompressionReader gzip.Reader

func NewDecompressionReader(source io.ReadCloser) (io.ReadCloser, error) {
	return gzip.NewReader(source)
}
