package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

func main() {
	mode := os.Getenv("MODE")
	switch mode {
	case "SEND":
		const size = 5*1024*1024 - 1
		data := make([]byte, size)
		for i := range data[:len(data)/2] {
			data[i] = byte(i % 256)
		}

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
	case "RECEIVE":
		n, err := io.Copy(io.Discard, os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to receive data: %v", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stderr, "Received data: %d\n", n)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unrecognized mode: %s\n", mode)
		os.Exit(1)
	}

}
