package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

const payloadSize = 5*1024*1024 - 1 // +-5Mb
const dataOffset = 1000

var messageBytes = []byte("hello world, this is a test to double check that the data stays the same")

func main() {
	mode := os.Getenv("MODE")
	switch mode {
	case "SEND":
		data := make([]byte, payloadSize)
		if _, err := rand.Read(data); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate random data: %v", err)
			os.Exit(1)
		}

		var embedCount int
		for i := 0; i < len(data); i += len(messageBytes) + dataOffset {
			copy(data[i:i+len(messageBytes)], messageBytes)
			embedCount++
		}

		if _, err := os.Stdout.Write(data); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write data to stdout: %v", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "Sent data: %d, Embedded %d times, First 100 bytes: %s\n", len(data), embedCount, data[:100])

		os.Exit(0)
	case "RECEIVE":
		buf := make([]byte, payloadSize)
		n, err := io.ReadFull(os.Stdin, buf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read data: %v", err)
			os.Exit(1)
		}

		go func() {
			n, _ := io.Copy(io.Discard, os.Stdin)
			fmt.Fprintf(os.Stderr, "Discarded %d bytes\n", n)
			if n > 0 {
				fmt.Fprintf(os.Stderr, "Data mismatch, expected %d bytes, got %d bytes\n", payloadSize, n)
				os.Exit(1)
			}
		}()

		var checkCount int

		for i := 0; i < len(buf); i += len(messageBytes) + dataOffset {
			if !bytes.Equal(buf[i:i+len(messageBytes)], messageBytes) {
				fmt.Fprintf(os.Stderr, "Data mismatch, expected %v, got %v\n", messageBytes, buf[i:i+len(messageBytes)])
				os.Exit(1)
			}
			checkCount++
		}

		fmt.Fprintf(os.Stderr, "Received data: %d, Checked %d times, First 100 bytes: %s\n", n, checkCount, buf[:100])
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unrecognized mode: %s\n", mode)
		os.Exit(1)
	}
}
