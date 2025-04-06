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
	r, err := gzip.NewReader(source)
	if err != nil {
		return nil, err
	}
	return r, nil
}
