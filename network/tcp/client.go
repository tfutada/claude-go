// TCP Client Example
// Demonstrates connecting to a TCP server

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
	// Dial establishes a TCP connection
	// This initiates the 3-way handshake
	conn, err := net.DialTimeout("tcp", "localhost:8080", 5*time.Second)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to TCP server at localhost:8080")
	fmt.Println("Type messages (or 'quit' to exit):")

	// Read user input
	stdinReader := bufio.NewReader(os.Stdin)
	// Read server responses
	connReader := bufio.NewReader(conn)

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

		// Send message to server
		// TCP ensures this data arrives in order and intact
		// Intentionally omit \n to demonstrate server timeout
		//_, err = conn.Write([]byte(input + "\n"))
		_, err = conn.Write([]byte(input))
		if err != nil {
			fmt.Printf("Send error: %v\n", err)
			return
		}

		// Set deadline for response
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		// Read server response
		response, err := connReader.ReadString('\n')
		if err != nil {
			fmt.Printf("Receive error: %v\n", err)
			return
		}

		fmt.Printf("< %s", response)

		if input == "quit" {
			fmt.Println("Disconnecting...")
			return
		}
	}
}
