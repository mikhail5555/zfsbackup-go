package compencrypt

import (
	"io"
	"sync"

	"github.com/klauspost/compress/zstd"
)

var _ io.WriteCloser = (*CompressionWriter)(nil)

type CompressionWriter struct {
	w           *zstd.Encoder
	destination io.WriteCloser
	once        sync.Once
}

func NewCompressionWriter(destination io.WriteCloser) *CompressionWriter {
	return &CompressionWriter{
		destination: destination,
	}
}

func (cw *CompressionWriter) Write(p []byte) (int, error) {
	var onceErr error
	cw.once.Do(func() {
		encoder, err := zstd.NewWriter(cw.destination)
		if err != nil {
			onceErr = err
			return
		}
		cw.w = encoder
	})
	if onceErr != nil {
		return 0, onceErr
	}

	return cw.w.Write(p)
}

func (cw *CompressionWriter) Close() error {
	if cw.w != nil {
		return cw.w.Close()
	}
	return cw.destination.Close()
}

var _ io.ReadCloser = (*DecompressionReader)(nil)

type DecompressionReader struct {
	source io.ReadCloser
	r      *zstd.Decoder
}

func NewDecompressionReader(source io.ReadCloser) *DecompressionReader {
	decoder, _ := zstd.NewReader(source)
	return &DecompressionReader{
		source: source,
		r:      decoder,
	}
}

func (dr *DecompressionReader) Read(p []byte) (int, error) {
	return dr.r.Read(p)
}

func (dr *DecompressionReader) Close() error {
	return dr.source.Close()
}
