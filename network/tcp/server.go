// TCP Server Example
// Demonstrates connection-oriented, reliable communication
//
// TCP characteristics:
// - Connection-oriented (3-way handshake)
// - Reliable delivery (acknowledgments, retransmission)
// - Ordered data (sequence numbers)
// - Flow control (sliding window)
// - Error checking (checksums)

package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

func main() {
	// Listen on TCP port 8080
	// "tcp" specifies the protocol
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Println("TCP Server listening on :8080")
	fmt.Println("Waiting for connections...")

	for {
		// Accept blocks until a client connects
		// This is where the 3-way handshake completes:
		// 1. Client sends SYN
		// 2. Server responds with SYN-ACK
		// 3. Client sends ACK
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			continue
		}

		// Handle each connection in a goroutine
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("[%s] Client connected\n", clientAddr)

	// Set read deadline to prevent hanging connections
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	reader := bufio.NewReader(conn)

	for {
		// Read until newline - TCP is stream-based
		// Data may arrive in chunks, but bufio handles this
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("[%s] Connection closed: %v\n", clientAddr, err)
			return
		}

		message = strings.TrimSpace(message)
		fmt.Printf("[%s] Received: %s\n", clientAddr, message)

		// Echo back with modification
		response := fmt.Sprintf("Server received: %s\n", message)

		// Write sends data reliably
		// TCP guarantees delivery or reports error
		_, err = conn.Write([]byte(response))
		if err != nil {
			fmt.Printf("[%s] Write error: %v\n", clientAddr, err)
			return
		}

		// Reset deadline for next message
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		if message == "quit" {
			fmt.Printf("[%s] Client requested disconnect\n", clientAddr)
			return
		}
	}
}
