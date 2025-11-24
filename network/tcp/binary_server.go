//go:build ignore

// TCP Binary Server Example
// Demonstrates length-prefixed binary protocol
//
// Protocol format:
// [4 bytes: message length (BigEndian uint32)][N bytes: message data]
//
// This approach works for any data type (text, images, protobuf, etc.)

package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", ":8081")
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Println("Binary TCP Server listening on :8081")
	fmt.Println("Protocol: [4-byte length][data]")
	fmt.Println("Waiting for connections...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			continue
		}

		go handleBinaryConnection(conn)
	}
}

func handleBinaryConnection(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("[%s] Client connected\n", clientAddr)

	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		// Step 1: Read message length (4 bytes)
		data, err := receiveMessage(conn)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("[%s] Client disconnected\n", clientAddr)
			} else {
				fmt.Printf("[%s] Read error: %v\n", clientAddr, err)
			}
			return
		}

		fmt.Printf("[%s] Received %d bytes: %s\n", clientAddr, len(data), string(data))

		// Echo back with prefix
		response := append([]byte("Server received: "), data...)

		// Send response
		if err := sendMessage(conn, response); err != nil {
			fmt.Printf("[%s] Write error: %v\n", clientAddr, err)
			return
		}

		// Check for quit command
		if string(data) == "quit" {
			fmt.Printf("[%s] Client requested disconnect\n", clientAddr)
			return
		}
	}
}

// receiveMessage reads a length-prefixed message
func receiveMessage(conn net.Conn) ([]byte, error) {
	// Read 4-byte length header
	var length uint32
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	// Sanity check to prevent huge allocations
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
