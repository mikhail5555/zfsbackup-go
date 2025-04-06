package compencrypt_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/someone1/zfsbackup-go/compencrypt"
)

func TestEncryptionAndDecryption(t *testing.T) {
	key := []byte("0123456789ABCDEF") // 16-byte key for AES-128
	originalData := []byte("This is a secret message that needs to be encrypted and then decrypted.")

	encryptedBuffer := new(bytes.Buffer)

	writer, err := compencrypt.NewEncryptionWriter(compencrypt.NopWriteCloser(encryptedBuffer), key)
	assert.NoError(t, err)

	n, err := writer.Write(originalData)
	assert.NoError(t, err)
	assert.Equal(t, len(originalData), n)

	assert.NoError(t, writer.Close())

	decReader, err := compencrypt.NewDecryptionReader(io.NopCloser(encryptedBuffer), key)
	assert.NoError(t, err)

	decryptedData := make([]byte, len(originalData))
	n, err = io.ReadFull(decReader, decryptedData)
	assert.NoError(t, err)
	assert.Equal(t, len(originalData), n)
	assert.Equal(t, originalData, decryptedData)
}

func TestLargeDataEncryptionAndDecryption(t *testing.T) {
	key := []byte("0123456789ABCDEF") // 16-byte key for AES-128

	originalData := make([]byte, 4*1024*1024)
	for i := range originalData {
		originalData[i] = byte(i % 256)
	}

	encryptedBuffer := new(bytes.Buffer)

	writer, err := compencrypt.NewEncryptionWriter(compencrypt.NopWriteCloser(encryptedBuffer), key)
	assert.NoError(t, err)

	n, err := writer.Write(originalData)
	assert.NoError(t, err)
	assert.Equal(t, len(originalData), n)

	assert.NoError(t, writer.Close())

	decReader, err := compencrypt.NewDecryptionReader(io.NopCloser(encryptedBuffer), key)
	assert.NoError(t, err)

	decryptedData := make([]byte, len(originalData))
	n, err = io.ReadFull(decReader, decryptedData)
	assert.NoError(t, err)
	assert.Equal(t, len(originalData), n)
	assert.Equal(t, originalData, decryptedData)
}
