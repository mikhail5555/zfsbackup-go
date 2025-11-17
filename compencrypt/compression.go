package compencrypt

import (
	"io"

	"github.com/klauspost/compress/zstd"
)

type CompressionWriter zstd.Encoder

func NewCompressionWriter(destination io.WriteCloser) (io.WriteCloser, error) {
	return zstd.NewWriter(destination)
}

type DecompressionReader struct {
	*zstd.Decoder
}

func NewDecompressionReader(source io.ReadCloser) (io.ReadCloser, error) {
	reader, err := zstd.NewReader(source)
	if err != nil {
		return nil, err
	}
	return &DecompressionReader{reader}, nil
}

func (r *DecompressionReader) Close() error {
	r.Decoder.Close()
	return nil
}
