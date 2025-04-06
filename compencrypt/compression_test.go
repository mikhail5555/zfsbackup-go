package compencrypt_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/someone1/zfsbackup-go/compencrypt"
)

func TestCompressionAndDecompression(t *testing.T) {
	originalData := []byte("This is a test message that needs to be compressed and then decompressed.")
	compressedBuffer := new(bytes.Buffer)

	writer, err := compencrypt.NewCompressionWriter(compencrypt.NopWriteCloser(compressedBuffer))
	assert.NoError(t, err)

	n, err := writer.Write(originalData)
	assert.NoError(t, err)
	assert.Equal(t, len(originalData), n)

	assert.NoError(t, writer.Close())

	reader, err := compencrypt.NewDecompressionReader(io.NopCloser(compressedBuffer))
	assert.NoError(t, err)

	decompressedData, err := io.ReadAll(reader)
	assert.NoError(t, err)

	assert.NoError(t, reader.Close())
	assert.Equal(t, originalData, decompressedData)
}

func TestLargeDataCompressionAndDecompression(t *testing.T) {
	originalData := make([]byte, 1024*1024)
	for i := range originalData {
		originalData[i] = byte(i % 64)
	}
	compressedBuffer := new(bytes.Buffer)

	writer, err := compencrypt.NewCompressionWriter(compencrypt.NopWriteCloser(compressedBuffer))
	assert.NoError(t, err)

	n, err := writer.Write(originalData)
	assert.NoError(t, err)
	assert.Equal(t, len(originalData), n)

	assert.NoError(t, writer.Close())

	compressionRatio := float64(compressedBuffer.Len()) / float64(len(originalData))
	t.Logf("Compressed %d bytes to %d bytes (%.2f%% of original size)", len(originalData), compressedBuffer.Len(), compressionRatio*100)

	reader, err := compencrypt.NewDecompressionReader(io.NopCloser(compressedBuffer))
	assert.NoError(t, err)

	decompressedData, err := io.ReadAll(reader)
	assert.NoError(t, err)

	assert.NoError(t, reader.Close())

	assert.Equal(t, originalData, decompressedData)
}
