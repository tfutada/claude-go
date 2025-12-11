// Demonstrates syscall overhead: comparing small reads vs buffered reads
package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

const (
	fileSize  = 10 * 1024 * 1024 // 10MB test file
	smallRead = 1                // 1 byte per syscall (worst case)
	largeRead = 4096             // 4KB buffered read
)

func main() {
	filename := "testfile.bin"

	// Create test file
	fmt.Println("Creating 10MB test file...")
	createTestFile(filename, fileSize)
	defer os.Remove(filename)

	// Test 1: Read 1 byte at a time (many syscalls)
	fmt.Println("\n--- Test 1: 1-byte reads (unbuffered) ---")
	duration1 := benchmarkSmallReads(filename, smallRead, 1024*1024) // read 1MB
	fmt.Printf("Read 1MB with 1-byte chunks: %v\n", duration1)
	fmt.Printf("Syscalls: ~1,000,000\n")

	// Test 2: Read 4KB at a time
	fmt.Println("\n--- Test 2: 4KB reads (unbuffered) ---")
	duration2 := benchmarkSmallReads(filename, largeRead, 1024*1024) // read 1MB
	fmt.Printf("Read 1MB with 4KB chunks: %v\n", duration2)
	fmt.Printf("Syscalls: ~256\n")

	// Test 3: bufio.Reader (buffered)
	fmt.Println("\n--- Test 3: 1-byte reads with bufio.Reader ---")
	duration3 := benchmarkBufferedReads(filename, 1024*1024) // read 1MB
	fmt.Printf("Read 1MB with bufio (1-byte app reads): %v\n", duration3)
	fmt.Printf("Actual syscalls: ~256 (bufio uses 4KB internal buffer)\n")

	// Summary
	fmt.Println("\n=== Summary ===")
	fmt.Printf("1-byte unbuffered: %v (baseline)\n", duration1)
	fmt.Printf("4KB unbuffered:    %v (%.1fx faster)\n", duration2, float64(duration1)/float64(duration2))
	fmt.Printf("bufio buffered:    %v (%.1fx faster)\n", duration3, float64(duration1)/float64(duration3))
}

func createTestFile(filename string, size int) {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Write in chunks for efficiency
	chunk := make([]byte, 4096)
	for i := range chunk {
		chunk[i] = byte(i % 256)
	}

	written := 0
	for written < size {
		n, err := f.Write(chunk)
		if err != nil {
			panic(err)
		}
		written += n
	}
}

// Direct f.Read() - each call = 1 syscall
func benchmarkSmallReads(filename string, chunkSize, totalBytes int) time.Duration {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	buf := make([]byte, chunkSize)
	read := 0

	start := time.Now()
	for read < totalBytes {
		n, err := f.Read(buf)
		if err != nil {
			break
		}
		read += n
	}
	return time.Since(start)
}

// bufio.Reader - internal 4KB buffer reduces syscalls
func benchmarkBufferedReads(filename string, totalBytes int) time.Duration {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	reader := bufio.NewReader(f) // default 4KB buffer
	read := 0

	start := time.Now()
	for read < totalBytes {
		_, err := reader.ReadByte() // 1 byte at a time from buffer
		if err != nil {
			break
		}
		read++
	}
	return time.Since(start)
}
