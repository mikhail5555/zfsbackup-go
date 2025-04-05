package compencrypt_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/someone1/zfsbackup-go/compencrypt"
)

func TestEncryptionAndDecryption(t *testing.T) {
	// Test data
	originalData := []byte("This is a secret message that needs to be encrypted and then decrypted.")
	key := []byte("0123456789ABCDEF") // 16-byte key for AES-128

	// Buffer to hold the encrypted data
	encryptedBuffer := new(bytes.Buffer)

	// Create an encryption writer
	encWriter := compencrypt.NewEncryptionWriter(compencrypt.NopWriteCloser(encryptedBuffer), key)

	// Write data to be encrypted
	n, err := encWriter.Write(originalData)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}
	if n != len(originalData) {
		t.Fatalf("Expected to write %d bytes, but wrote %d", len(originalData), n)
	}

	// Close the writer
	if err := encWriter.Close(); err != nil {
		t.Fatalf("Failed to close encryption writer: %v", err)
	}

	// Create a decryption reader with the encrypted data
	decReader := compencrypt.NewDecryptionReader(io.NopCloser(encryptedBuffer), key)

	// Read and decrypt the data
	decryptedData := make([]byte, len(originalData))
	n, err = io.ReadFull(decReader, decryptedData)
	if err != nil {
		t.Fatalf("Failed to decrypt data: %v", err)
	}
	if n != len(originalData) {
		t.Fatalf("Expected to read %d bytes, but read %d", len(originalData), n)
	}

	// Verify the decrypted data matches the original
	if !bytes.Equal(originalData, decryptedData) {
		t.Fatalf("Decrypted data does not match original data.\nOriginal: %s\nDecrypted: %s",
			string(originalData), string(decryptedData))
	}
}

func TestLargeDataEncryptionAndDecryption(t *testing.T) {
	// Generate larger test data (100KB)
	originalData := make([]byte, 100*1024)
	for i := range originalData {
		originalData[i] = byte(i % 256)
	}

	key := []byte("0123456789ABCDEF") // 16-byte key for AES-128

	// Buffer to hold the encrypted data
	encryptedBuffer := new(bytes.Buffer)

	// Create an encryption writer
	encWriter := compencrypt.NewEncryptionWriter(compencrypt.NopWriteCloser(encryptedBuffer), key)

	// Write data to be encrypted
	n, err := encWriter.Write(originalData)
	if err != nil {
		t.Fatalf("Failed to encrypt large data: %v", err)
	}
	if n != len(originalData) {
		t.Fatalf("Expected to write %d bytes, but wrote %d", len(originalData), n)
	}

	// Close the writer
	if err := encWriter.Close(); err != nil {
		t.Fatalf("Failed to close encryption writer: %v", err)
	}

	// Create a decryption reader with the encrypted data
	decReader := compencrypt.NewDecryptionReader(io.NopCloser(encryptedBuffer), key)

	// Read and decrypt the data
	decryptedData := make([]byte, len(originalData))
	n, err = io.ReadFull(decReader, decryptedData)
	if err != nil {
		t.Fatalf("Failed to decrypt large data: %v", err)
	}
	if n != len(originalData) {
		t.Fatalf("Expected to read %d bytes, but read %d", len(originalData), n)
	}

	// Verify the decrypted data matches the original
	if !bytes.Equal(originalData, decryptedData) {
		t.Fatalf("Decrypted large data does not match original data")
	}
}
