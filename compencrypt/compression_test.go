package compencrypt_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/someone1/zfsbackup-go/compencrypt"
)

func TestCompressionAndDecompression(t *testing.T) {
	// Test data
	originalData := []byte("This is a test message that needs to be compressed and then decompressed.")

	// Buffer to hold the compressed data
	compressedBuffer := new(bytes.Buffer)

	// Create a compression writer
	compWriter := compencrypt.NewCompressionWriter(compressedBuffer)

	// Write data to be compressed
	n, err := compWriter.Write(originalData)
	if err != nil {
		t.Fatalf("Failed to compress data: %v", err)
	}
	if n != len(originalData) {
		t.Fatalf("Expected to write %d bytes, but wrote %d", len(originalData), n)
	}

	// Close the writer
	if err := compWriter.Close(); err != nil {
		t.Fatalf("Failed to close compression writer: %v", err)
	}

	// Create a decompression reader with the compressed data
	decompReader := compencrypt.NewDecompressionReader(compressedBuffer)

	// Read and decompress the data
	decompressedData, err := io.ReadAll(decompReader)
	if err != nil {
		t.Fatalf("Failed to decompress data: %v", err)
	}

	// Close the reader
	if err := decompReader.Close(); err != nil {
		t.Fatalf("Failed to close decompression reader: %v", err)
	}

	// Verify the decompressed data matches the original
	if !bytes.Equal(originalData, decompressedData) {
		t.Fatalf("Decompressed data does not match original data.\nOriginal: %s\nDecompressed: %s",
			string(originalData), string(decompressedData))
	}
}

func TestLargeDataCompressionAndDecompression(t *testing.T) {
	// Generate larger test data (1MB of repeating pattern for good compression)
	originalData := make([]byte, 1024*1024)
	for i := range originalData {
		originalData[i] = byte(i % 64)
	}

	// Buffer to hold the compressed data
	compressedBuffer := new(bytes.Buffer)

	// Create a compression writer
	compWriter := compencrypt.NewCompressionWriter(compressedBuffer)

	// Write data to be compressed
	n, err := compWriter.Write(originalData)
	if err != nil {
		t.Fatalf("Failed to compress large data: %v", err)
	}
	if n != len(originalData) {
		t.Fatalf("Expected to write %d bytes, but wrote %d", len(originalData), n)
	}

	// Close the writer
	if err := compWriter.Close(); err != nil {
		t.Fatalf("Failed to close compression writer: %v", err)
	}

	// Print compression ratio
	compressionRatio := float64(compressedBuffer.Len()) / float64(len(originalData))
	t.Logf("Compressed %d bytes to %d bytes (%.2f%% of original size)",
		len(originalData), compressedBuffer.Len(), compressionRatio*100)

	// Create a decompression reader with the compressed data
	decompReader := compencrypt.NewDecompressionReader(compressedBuffer)

	// Read and decompress the data
	decompressedData, err := io.ReadAll(decompReader)
	if err != nil {
		t.Fatalf("Failed to decompress large data: %v", err)
	}

	// Close the reader
	if err := decompReader.Close(); err != nil {
		t.Fatalf("Failed to close decompression reader: %v", err)
	}

	// Verify the decompressed data matches the original
	if !bytes.Equal(originalData, decompressedData) {
		t.Fatalf("Decompressed large data does not match original data")
	}
}
