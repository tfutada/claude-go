// WebSocket Client Example
// Demonstrates connecting to a WebSocket server

package main

import (
	"bufio"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func main() {
	// Connect to server
	conn, err := net.DialTimeout("tcp", "localhost:8082", 5*time.Second)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	// Perform WebSocket handshake
	if err := performHandshake(conn); err != nil {
		fmt.Printf("Handshake failed: %v\n", err)
		return
	}

	fmt.Println("WebSocket connection established!")
	fmt.Println("Type messages (or 'quit' to exit):")

	// Start goroutine to read server responses
	go readMessages(conn)

	// Read user input and send
	stdinReader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, err := stdinReader.ReadString('\n')
		if err != nil {
			return
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if input == "quit" {
			// Send close frame
			writeFrame(conn, []byte{}, 0x8)
			fmt.Println("Closing connection...")
			time.Sleep(500 * time.Millisecond)
			return
		}

		// Send text frame
		err = writeFrame(conn, []byte(input), 0x1)
		if err != nil {
			fmt.Printf("Send error: %v\n", err)
			return
		}
	}
}

func performHandshake(conn net.Conn) error {
	// Generate random key
	keyBytes := make([]byte, 16)
	rand.Read(keyBytes)
	key := base64.StdEncoding.EncodeToString(keyBytes)

	// Send upgrade request
	request := fmt.Sprintf(
		"GET / HTTP/1.1\r\n"+
			"Host: localhost:8082\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Key: %s\r\n"+
			"Sec-WebSocket-Version: 13\r\n"+
			"\r\n",
		key,
	)
	_, err := conn.Write([]byte(request))
	if err != nil {
		return err
	}

	// Read response
	reader := bufio.NewReader(conn)
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	if !strings.Contains(statusLine, "101") {
		return fmt.Errorf("expected 101 Switching Protocols, got: %s", statusLine)
	}

	// Read headers
	var acceptKey string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(strings.ToLower(line), "sec-websocket-accept:") {
			acceptKey = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		}
	}

	// Verify accept key
	expectedKey := computeAcceptKey(key)
	if acceptKey != expectedKey {
		return fmt.Errorf("invalid accept key: got %s, expected %s", acceptKey, expectedKey)
	}

	return nil
}

func computeAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key + websocketGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func readMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, opcode, err := readFrame(reader)
		if err != nil {
			return
		}

		switch opcode {
		case 0x1: // Text
			fmt.Printf("\n< %s\n> ", string(message))
		case 0x8: // Close
			fmt.Println("\nServer closed connection")
			return
		case 0x9: // Ping
			writeFrame(conn, message, 0xA)
		}
	}
}

func readFrame(reader *bufio.Reader) ([]byte, byte, error) {
	header := make([]byte, 2)
	_, err := io.ReadFull(reader, header)
	if err != nil {
		return nil, 0, err
	}

	opcode := header[0] & 0x0F
	masked := (header[1] & 0x80) != 0
	length := uint64(header[1] & 0x7F)

	if length == 126 {
		extended := make([]byte, 2)
		io.ReadFull(reader, extended)
		length = uint64(binary.BigEndian.Uint16(extended))
	} else if length == 127 {
		extended := make([]byte, 8)
		io.ReadFull(reader, extended)
		length = binary.BigEndian.Uint64(extended)
	}

	var maskKey []byte
	if masked {
		maskKey = make([]byte, 4)
		io.ReadFull(reader, maskKey)
	}

	payload := make([]byte, length)
	_, err = io.ReadFull(reader, payload)
	if err != nil {
		return nil, 0, err
	}

	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	return payload, opcode, nil
}

// writeFrame writes a masked WebSocket frame (client -> server must be masked)
func writeFrame(conn net.Conn, payload []byte, opcode byte) error {
	var frame []byte

	// First byte: FIN + opcode
	frame = append(frame, 0x80|opcode)

	// Second byte: MASK bit (1) + payload length
	length := len(payload)
	if length < 126 {
		frame = append(frame, 0x80|byte(length))
	} else if length < 65536 {
		frame = append(frame, 0x80|126)
		frame = append(frame, byte(length>>8), byte(length))
	} else {
		frame = append(frame, 0x80|127)
		for i := 7; i >= 0; i-- {
			frame = append(frame, byte(length>>(i*8)))
		}
	}

	// Generate and append mask key
	maskKey := make([]byte, 4)
	rand.Read(maskKey)
	frame = append(frame, maskKey...)

	// Mask and append payload
	maskedPayload := make([]byte, length)
	for i := range payload {
		maskedPayload[i] = payload[i] ^ maskKey[i%4]
	}
	frame = append(frame, maskedPayload...)

	_, err := conn.Write(frame)
	return err
}
