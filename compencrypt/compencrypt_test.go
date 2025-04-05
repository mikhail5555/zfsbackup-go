package compencrypt_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/someone1/zfsbackup-go/compencrypt"
)

func TestCompressEncryptAndDecryptDecompress(t *testing.T) {
	// Test data
	originalData := []byte("This is a secret message that needs to be compressed, encrypted, decrypted, and then decompressed.")
	key := []byte("0123456789ABCDEF") // 16-byte key for AES-128

	// Buffer to hold the compressed and encrypted data
	processedBuffer := new(bytes.Buffer)

	// Create a compress and encrypt writer
	writer := compencrypt.NewCompressAndEncryptWriter(compencrypt.NopWriteCloser(processedBuffer), key)

	// Write data to be compressed and encrypted
	n, err := writer.Write(originalData)
	if err != nil {
		t.Fatalf("Failed to compress and encrypt data: %v", err)
	}
	if n != len(originalData) {
		t.Fatalf("Expected to write %d bytes, but wrote %d", len(originalData), n)
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close compress and encrypt writer: %v", err)
	}

	// Create a decrypt and decompress reader with the processed data
	reader := compencrypt.NewDecryptAndDecompressReader(io.NopCloser(processedBuffer), key)

	// Read, decrypt, and decompress the data
	resultData, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to decrypt and decompress data: %v", err)
	}

	// Close the reader
	if err := reader.Close(); err != nil {
		t.Fatalf("Failed to close decrypt and decompress reader: %v", err)
	}

	// Verify the result matches the original
	if !bytes.Equal(originalData, resultData) {
		t.Fatalf("Processed data does not match original data.\nOriginal: %s\nResult: %s",
			string(originalData), string(resultData))
	}
}

func TestLargeDataCompressEncryptAndDecryptDecompress(t *testing.T) {
	// Generate larger test data (1MB of repeating pattern for good compression)
	originalData := make([]byte, 1024*1024)
	for i := range originalData {
		originalData[i] = byte(i % 64)
	}

	key := []byte("0123456789ABCDEF") // 16-byte key for AES-128

	// Buffer to hold the compressed and encrypted data
	processedBuffer := new(bytes.Buffer)

	// Create a compress and encrypt writer
	writer := compencrypt.NewCompressAndEncryptWriter(compencrypt.NopWriteCloser(processedBuffer), key)

	// Write data to be compressed and encrypted
	n, err := writer.Write(originalData)
	if err != nil {
		t.Fatalf("Failed to compress and encrypt large data: %v", err)
	}
	if n != len(originalData) {
		t.Fatalf("Expected to write %d bytes, but wrote %d", len(originalData), n)
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to close compress and encrypt writer: %v", err)
	}

	// Print compression+encryption ratio
	ratio := float64(processedBuffer.Len()) / float64(len(originalData))
	t.Logf("Compressed and encrypted %d bytes to %d bytes (%.2f%% of original size)",
		len(originalData), processedBuffer.Len(), ratio*100)

	// Create a decrypt and decompress reader with the processed data
	reader := compencrypt.NewDecryptAndDecompressReader(io.NopCloser(processedBuffer), key)

	// Read, decrypt, and decompress the data
	resultData, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to decrypt and decompress large data: %v", err)
	}

	// Close the reader
	if err := reader.Close(); err != nil {
		t.Fatalf("Failed to close decrypt and decompress reader: %v", err)
	}

	// Verify the result matches the original
	if !bytes.Equal(originalData, resultData) {
		t.Fatalf("Processed large data does not match original data")
	}
}
