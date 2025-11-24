//go:build ignore

// TCP Binary Client Example
// Connects to binary_server.go using length-prefixed protocol
//
// Run server first: go run binary_server.go
// Then run client:  go run binary_client.go

package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func main() {
	// Connect to server
	conn, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to binary server on :8081")
	fmt.Println("Protocol: [4-byte length][data]")
	fmt.Println("Type messages (or 'quit' to exit):")

	reader := bufio.NewReader(os.Stdin)

	for {
		// Read user input
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Input error: %v\n", err)
			return
		}

		message := strings.TrimSpace(input)
		if message == "" {
			continue
		}

		// Send as binary (length-prefixed)
		if err := sendMessage(conn, []byte(message)); err != nil {
			fmt.Printf("Send error: %v\n", err)
			return
		}

		// Receive response
		response, err := receiveMessage(conn)
		if err != nil {
			fmt.Printf("Receive error: %v\n", err)
			return
		}

		fmt.Printf("Response (%d bytes): %s\n", len(response), string(response))

		if message == "quit" {
			break
		}
	}
}

// sendMessage sends a length-prefixed message
func sendMessage(conn net.Conn, data []byte) error {
	// Write 4-byte length header
	length := uint32(len(data))
	if err := binary.Write(conn, binary.BigEndian, length); err != nil {
		return err
	}

	// Write data
	_, err := conn.Write(data)
	return err
}

// receiveMessage reads a length-prefixed message
func receiveMessage(conn net.Conn) ([]byte, error) {
	// Read 4-byte length header
	var length uint32
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	// Sanity check
	if length > 10*1024*1024 { // 10MB max
		return nil, fmt.Errorf("message too large: %d bytes", length)
	}

	// Read exactly 'length' bytes
	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}

	return data, nil
}
