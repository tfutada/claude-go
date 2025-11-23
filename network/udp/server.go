// UDP Server Example
// Demonstrates connectionless, unreliable (but fast) communication
//
// UDP characteristics:
// - Connectionless (no handshake)
// - Unreliable (no guarantees)
// - Unordered (packets may arrive out of order)
// - No flow control
// - Lower overhead, faster
// - Message boundaries preserved (datagram-based)

package main

import (
	"fmt"
	"net"
)

func main() {
	// ListenUDP creates a UDP endpoint
	// No connection established - just ready to receive
	addr, err := net.ResolveUDPAddr("udp", ":8081")
	if err != nil {
		fmt.Printf("Address resolution error: %v\n", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("UDP Server listening on :8081")
	fmt.Println("Waiting for datagrams...")

	// Buffer for incoming data
	// UDP preserves message boundaries - each Read gets one datagram
	buffer := make([]byte, 1024)

	for {
		// ReadFromUDP receives a single datagram
		// Returns the number of bytes and sender's address
		// No connection state - each message is independent
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Read error: %v\n", err)
			continue
		}

		message := string(buffer[:n])
		fmt.Printf("[%s] Received (%d bytes): %s\n", remoteAddr, n, message)

		// Send response back to the sender
		// We must specify the address since there's no connection
		response := fmt.Sprintf("Server received %d bytes", n)
		_, err = conn.WriteToUDP([]byte(response), remoteAddr)
		if err != nil {
			fmt.Printf("Write error: %v\n", err)
			// Unlike TCP, we continue even if write fails
			// UDP doesn't guarantee delivery anyway
		}

		// Note: If this write fails, the client won't know
		// There's no acknowledgment in UDP
	}
}
