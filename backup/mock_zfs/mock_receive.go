package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	n, err := io.Copy(io.Discard, os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to receive data: %v", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Received data: %d\n", n)
	os.Exit(0)
}
