// Package main demonstrates the practical power of duck typing:
// Using interfaces with third-party types you don't control.
//
// Key insight: Consumer defines interface, not provider.
// This enables testing/mocking without modifying original code.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Reader is OUR interface - we define what we need
// os.File, strings.Reader, bytes.Buffer all satisfy this
// without knowing about our interface!
type Reader interface {
	Read([]byte) (int, error)
}

// CountBytes works with ANY Reader
// - os.File (real file)
// - strings.Reader (for testing)
// - bytes.Buffer (in-memory)
// - net.Conn (network)
// - http.Response.Body (HTTP)
func CountBytes(r Reader) (int, error) {
	buf := make([]byte, 1024)
	total := 0

	for {
		n, err := r.Read(buf)
		total += n
		if err == io.EOF {
			break
		}
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// === Traditional OOP Problem ===
// In Java/C#, you'd need:
//   interface Reader { int read(byte[] buf); }
//   class File implements Reader { ... }  // Can't modify!
//
// You can't make java.io.File implement YOUR interface
// without wrapper classes or adapters.

// === Go Solution ===
// os.File already has Read() method
// It automatically satisfies our Reader interface!

func main() {
	fmt.Println("=== Duck Typing with Third-Party Types ===\n")

	// 1. Use with real file (os.File)
	fmt.Println("1. With os.File (real file):")
	// Create a temp file to demonstrate
	tmpFile, _ := os.CreateTemp("", "demo*.txt")
	tmpFile.WriteString("This is content in a real file on disk!")
	tmpFile.Seek(0, 0) // reset to beginning

	count, _ := CountBytes(tmpFile)
	fmt.Printf("   Temp file has %d bytes\n", count)
	tmpFile.Close()
	os.Remove(tmpFile.Name())

	// 2. Use with strings.Reader (for testing!)
	fmt.Println("\n2. With strings.Reader (testing/mocking):")
	testData := "Hello, this is test data for our function"
	stringReader := strings.NewReader(testData)
	count, _ = CountBytes(stringReader)
	fmt.Printf("   Test string has %d bytes\n", count)

	// 3. The power: same function, different sources
	fmt.Println("\n3. Polymorphism in action:")

	// All these satisfy Reader without modification:
	sources := []struct {
		name   string
		reader Reader
	}{
		{"strings.Reader", strings.NewReader("short")},
		{"strings.Reader", strings.NewReader("a]longer string here")},
	}

	for _, src := range sources {
		count, _ := CountBytes(src.reader)
		fmt.Printf("   %s: %d bytes\n", src.name, count)
	}

	// 4. Why this matters for testing
	fmt.Println("\n=== Why This Matters ===")
	fmt.Println("Traditional OOP:")
	fmt.Println("  - os.File doesn't implement YOUR interface")
	fmt.Println("  - Need wrapper/adapter classes")
	fmt.Println("  - Tight coupling to concrete types")
	fmt.Println("")
	fmt.Println("Go Duck Typing:")
	fmt.Println("  - Define interface where YOU need it")
	fmt.Println("  - os.File automatically satisfies it")
	fmt.Println("  - Easy testing with strings.Reader")
	fmt.Println("  - No modification to third-party code")

	// 5. Real-world example: HTTP handler testing
	fmt.Println("\n=== Real-World: HTTP Testing ===")
	fmt.Println("// Your interface")
	fmt.Println("type ResponseWriter interface {")
	fmt.Println("    Write([]byte) (int, error)")
	fmt.Println("    WriteHeader(int)")
	fmt.Println("}")
	fmt.Println("")
	fmt.Println("// http.ResponseWriter satisfies this!")
	fmt.Println("// httptest.ResponseRecorder also satisfies this!")
	fmt.Println("// → Easy testing without mocking frameworks")
}
