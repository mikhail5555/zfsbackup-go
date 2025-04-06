package main

import (
	"crypto/rand"
	"fmt"
	"os"
)

func main() {
	const size = 5*1024*1024 - 1
	data := make([]byte, size)
	for i := range data[:len(data)/2] {
		data[i] = byte(i % 256)
	}

	// Fill with random data
	if _, err := rand.Read(data[len(data)/2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate random data: %v", err)
		os.Exit(1)
	}

	// Write the random data to stdout
	if _, err := os.Stdout.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write data to stdout: %v", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Received data: %d\n", size)

	os.Exit(0)
}
