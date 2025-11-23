// WebSocket Server Example
// Demonstrates the upgrade from HTTP to WebSocket protocol
//
// WebSocket characteristics:
// - Starts as HTTP, upgrades to WebSocket
// - Bidirectional communication
// - Message-based (not stream-based like TCP)
// - Low overhead binary framing
// - Persistent connection

package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// WebSocket GUID for handshake (RFC 6455)
const websocketGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func main() {
	listener, err := net.Listen("tcp", ":8082")
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Println("WebSocket Server listening on :8082")
	fmt.Println("Connect with: ws://localhost:8082")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			continue
		}

		go handleWebSocket(conn)
	}
}

func handleWebSocket(conn net.Conn) {
	defer conn.Close()
	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("[%s] Connection received\n", clientAddr)

	reader := bufio.NewReader(conn)

	// Step 1: Read HTTP Upgrade request
	request, err := http.ReadRequest(reader)
	if err != nil {
		fmt.Printf("[%s] Failed to read request: %v\n", clientAddr, err)
		return
	}

	// Step 2: Validate WebSocket upgrade request
	if !isWebSocketUpgrade(request) {
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}

	// Step 3: Get the Sec-WebSocket-Key
	key := request.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}

	// Step 4: Calculate accept key (SHA1 hash of key + GUID, base64 encoded)
	acceptKey := computeAcceptKey(key)

	// Step 5: Send upgrade response
	response := fmt.Sprintf(
		"HTTP/1.1 101 Switching Protocols\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Accept: %s\r\n"+
			"\r\n",
		acceptKey,
	)
	conn.Write([]byte(response))

	fmt.Printf("[%s] WebSocket connection established\n", clientAddr)

	// Step 6: Now communicate using WebSocket frames
	for {
		// Read WebSocket frame
		message, opcode, err := readFrame(reader)
		if err != nil {
			fmt.Printf("[%s] Read error: %v\n", clientAddr, err)
			return
		}

		switch opcode {
		case 0x1: // Text frame
			fmt.Printf("[%s] Received: %s\n", clientAddr, string(message))

			// Echo back
			response := fmt.Sprintf("Server received: %s", string(message))
			err = writeFrame(conn, []byte(response), 0x1)
			if err != nil {
				fmt.Printf("[%s] Write error: %v\n", clientAddr, err)
				return
			}

		case 0x8: // Close frame
			fmt.Printf("[%s] Close frame received\n", clientAddr)
			// Send close frame back
			writeFrame(conn, []byte{}, 0x8)
			return

		case 0x9: // Ping frame
			fmt.Printf("[%s] Ping received\n", clientAddr)
			// Respond with pong
			writeFrame(conn, message, 0xA)

		case 0xA: // Pong frame
			fmt.Printf("[%s] Pong received\n", clientAddr)
		}
	}
}

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.ToLower(r.Header.Get("Upgrade")) == "websocket" &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

func computeAcceptKey(key string) string {
	h := sha1.New()
	h.Write([]byte(key + websocketGUID))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// readFrame reads a WebSocket frame
// Frame format:
//
//	0                   1                   2                   3
//	0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//
// +-+-+-+-+-------+-+-------------+-------------------------------+
// |F|R|R|R| opcode|M| Payload len |    Extended payload length    |
// |I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
// |N|V|V|V|       |S|             |   (if payload len==126/127)   |
// | |1|2|3|       |K|             |                               |
// +-+-+-+-+-------+-+-------------+-------------------------------+
func readFrame(reader *bufio.Reader) ([]byte, byte, error) {
	// Read first 2 bytes
	header := make([]byte, 2)
	_, err := io.ReadFull(reader, header)
	if err != nil {
		return nil, 0, err
	}

	// Parse header
	// fin := (header[0] & 0x80) != 0  // Final fragment
	opcode := header[0] & 0x0F        // Opcode
	masked := (header[1] & 0x80) != 0 // Is masked?
	length := uint64(header[1] & 0x7F)

	// Extended payload length
	if length == 126 {
		extended := make([]byte, 2)
		io.ReadFull(reader, extended)
		length = uint64(binary.BigEndian.Uint16(extended))
	} else if length == 127 {
		extended := make([]byte, 8)
		io.ReadFull(reader, extended)
		length = binary.BigEndian.Uint64(extended)
	}

	// Read masking key (client -> server is always masked)
	var maskKey []byte
	if masked {
		maskKey = make([]byte, 4)
		io.ReadFull(reader, maskKey)
	}

	// Read payload
	payload := make([]byte, length)
	_, err = io.ReadFull(reader, payload)
	if err != nil {
		return nil, 0, err
	}

	// Unmask payload
	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	return payload, opcode, nil
}

// writeFrame writes a WebSocket frame (server -> client, not masked)
func writeFrame(conn net.Conn, payload []byte, opcode byte) error {
	var frame []byte

	// First byte: FIN + opcode
	frame = append(frame, 0x80|opcode)

	// Second byte: payload length (server doesn't mask)
	length := len(payload)
	if length < 126 {
		frame = append(frame, byte(length))
	} else if length < 65536 {
		frame = append(frame, 126)
		frame = append(frame, byte(length>>8), byte(length))
	} else {
		frame = append(frame, 127)
		for i := 7; i >= 0; i-- {
			frame = append(frame, byte(length>>(i*8)))
		}
	}

	// Append payload
	frame = append(frame, payload...)

	_, err := conn.Write(frame)
	return err
}
