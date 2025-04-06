package compencrypt_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/someone1/zfsbackup-go/compencrypt"
)

func TestCompressEncryptAndDecryptDecompress(t *testing.T) {
	key := []byte("0123456789ABCDEF") // 16-byte key for AES-128
	originalData := []byte("This is a secret message that needs to be compressed, encrypted, decrypted, and then decompressed.")

	processedBuffer := new(bytes.Buffer)

	writer, err := compencrypt.NewCompressAndEncryptWriter(compencrypt.NopWriteCloser(processedBuffer), key)
	assert.NoError(t, err)

	n, err := writer.Write(originalData)
	assert.NoError(t, err)
	assert.Equal(t, len(originalData), n)

	assert.NoError(t, writer.Close())

	reader, err := compencrypt.NewDecryptAndDecompressReader(io.NopCloser(processedBuffer), key)
	assert.NoError(t, err)

	resultData, err := io.ReadAll(reader)
	assert.NoError(t, err)

	assert.NoError(t, reader.Close())

	assert.Equal(t, originalData, resultData)
}

func TestLargeDataCompressEncryptAndDecryptDecompress(t *testing.T) {
	key := []byte("0123456789ABCDEF") // 16-byte key for AES-128

	originalData := make([]byte, 4*1024*1024)
	for i := range originalData {
		originalData[i] = byte(i % 256)
	}

	processedBuffer := new(bytes.Buffer)

	writer, err := compencrypt.NewCompressAndEncryptWriter(compencrypt.NopWriteCloser(processedBuffer), key)
	assert.NoError(t, err)

	n, err := writer.Write(originalData)
	assert.NoError(t, err)
	assert.Equal(t, len(originalData), n)

	assert.NoError(t, writer.Close())

	ratio := float64(processedBuffer.Len()) / float64(len(originalData))
	t.Logf("Compressed and encrypted %d bytes to %d bytes (%.2f%% of original size)", len(originalData), processedBuffer.Len(), ratio*100)

	reader, err := compencrypt.NewDecryptAndDecompressReader(io.NopCloser(processedBuffer), key)
	assert.NoError(t, err)

	resultData, err := io.ReadAll(reader)
	assert.NoError(t, err)

	assert.NoError(t, reader.Close())

	assert.Equal(t, originalData, resultData)
}
