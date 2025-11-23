// UDP Client Example
// Demonstrates sending datagrams to a UDP server

package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	// Resolve server address
	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:8081")
	if err != nil {
		fmt.Printf("Address resolution error: %v\n", err)
		return
	}

	// DialUDP "connects" to the server
	// Note: This doesn't actually establish a connection!
	// It just sets the default destination for Write()
	// and filters incoming packets to only this address
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Printf("Failed to dial: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("UDP client ready to send to localhost:8081")
	fmt.Println("Type messages (or 'quit' to exit):")

	stdinReader := bufio.NewReader(os.Stdin)
	buffer := make([]byte, 1024)

	for {
		fmt.Print("> ")
		input, err := stdinReader.ReadString('\n')
		if err != nil {
			fmt.Printf("Input error: %v\n", err)
			return
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if input == "quit" {
			fmt.Println("Exiting...")
			return
		}

		// Send datagram
		// Unlike TCP, this is a single message unit
		// Either the whole message arrives or nothing
		_, err = conn.Write([]byte(input))
		if err != nil {
			fmt.Printf("Send error: %v\n", err)
			continue
		}

		// Set timeout for response
		// Since UDP has no built-in timeout, we must handle it
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))

		// Try to read response
		n, err := conn.Read(buffer)
		if err != nil {
			// Timeout or other error
			// In UDP, this could mean:
			// - Packet was lost
			// - Server didn't respond
			// - Response was lost
			fmt.Printf("No response (packet may be lost): %v\n", err)
			continue
		}

		fmt.Printf("< %s\n", string(buffer[:n]))
	}
}
