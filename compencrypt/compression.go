package compencrypt

import (
	"compress/gzip"
	"io"
	"sync"
)

var _ io.WriteCloser = (*CompressionWriter)(nil)

type CompressionWriter struct {
	w *gzip.Writer
}

func NewCompressionWriter(destination io.WriteCloser) *CompressionWriter {
	return &CompressionWriter{w: gzip.NewWriter(destination)}
}

func (cw *CompressionWriter) Write(p []byte) (int, error) {
	return cw.w.Write(p)
}

func (cw *CompressionWriter) Close() error {
	return cw.w.Close()
}

var _ io.ReadCloser = (*DecompressionReader)(nil)

type DecompressionReader struct {
	r      *gzip.Reader
	source io.ReadCloser
	once   sync.Once
}

func NewDecompressionReader(source io.ReadCloser) *DecompressionReader {
	return &DecompressionReader{source: source}
}

func (dr *DecompressionReader) Read(p []byte) (int, error) {
	var onceErr error
	dr.once.Do(func() {
		dr.r, onceErr = gzip.NewReader(dr.source)
	})
	if onceErr != nil {
		return 0, onceErr
	}

	return dr.r.Read(p)
}

func (dr *DecompressionReader) Close() error {
	if dr.r != nil {
		return dr.r.Close()
	}
	return dr.source.Close()
}
