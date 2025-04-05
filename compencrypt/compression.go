package compencrypt

import (
	"compress/gzip"
	"io"
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
	r *gzip.Reader
}

func NewDecompressionReader(source io.ReadCloser) *DecompressionReader {
	r, _ := gzip.NewReader(source)
	return &DecompressionReader{r: r}
}

func (dr *DecompressionReader) Read(p []byte) (int, error) {
	return dr.r.Read(p)
}

func (dr *DecompressionReader) Close() error {
	return dr.r.Close()
}
