package compencrypt

import (
	"io"
	"sync"

	"github.com/klauspost/compress/zstd"
)

var _ io.WriteCloser = (*CompressionWriter)(nil)

type CompressionWriter struct {
	w           *zstd.Encoder
	destination io.Writer
	once        sync.Once
}

func NewCompressionWriter(destination io.Writer) *CompressionWriter {
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
	return nil
}

var _ io.ReadCloser = (*DecompressionReader)(nil)

type DecompressionReader struct {
	r      *zstd.Decoder
	source io.Reader
	once   sync.Once
}

func NewDecompressionReader(source io.Reader) *DecompressionReader {
	return &DecompressionReader{
		source: source,
	}
}

func (dr *DecompressionReader) Read(p []byte) (int, error) {
	var onceErr error
	dr.once.Do(func() {
		decoder, err := zstd.NewReader(dr.source)
		if err != nil {
			onceErr = err
			return
		}
		dr.r = decoder
	})
	if onceErr != nil {
		return 0, onceErr
	}

	return dr.r.Read(p)
}

func (dr *DecompressionReader) Close() error {
	if dr.r != nil {
		dr.r.Close()
	}
	return nil
}
