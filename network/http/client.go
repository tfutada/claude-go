// Minimal HTTP/1.1 Client (no standard library http package)
// Demonstrates HTTP protocol fundamentals from client side
//
// Counterpart to server.go - tests against it
// Run: go run client.go (with server.go running on :8083)

package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func main() {
	baseURL := "localhost:8083"

	fmt.Println("=== Minimal HTTP/1.1 Client ===")
	fmt.Println("Testing against server at", baseURL)
	fmt.Println()

	// Test 1: GET /
	fmt.Println("--- Test 1: GET / ---")
	resp, err := httpGet(baseURL, "/")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printResponse(resp, true) // truncate body
	}

	// Test 2: GET /api/time
	fmt.Println("\n--- Test 2: GET /api/time ---")
	resp, err = httpGet(baseURL, "/api/time")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printResponse(resp, false)
	}

	// Test 3: POST /api/echo
	fmt.Println("\n--- Test 3: POST /api/echo ---")
	resp, err = httpPost(baseURL, "/api/echo", "Hello from raw TCP client!")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printResponse(resp, false)
	}

	// Test 4: GET /headers
	fmt.Println("\n--- Test 4: GET /headers ---")
	resp, err = httpGet(baseURL, "/headers")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printResponse(resp, false)
	}

	// Test 5: GET /notfound (404)
	fmt.Println("\n--- Test 5: GET /notfound (expect 404) ---")
	resp, err = httpGet(baseURL, "/notfound")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		printResponse(resp, false)
	}
}

// HTTPResponse holds parsed response
type HTTPResponse struct {
	StatusCode   int
	StatusText   string
	Headers      map[string]string
	Body         string
	ResponseTime time.Duration
}

// httpGet performs GET request using raw TCP
func httpGet(host, path string) (*HTTPResponse, error) {
	return httpRequest(host, "GET", path, nil, "")
}

// httpPost performs POST request using raw TCP
func httpPost(host, path, body string) (*HTTPResponse, error) {
	headers := map[string]string{
		"Content-Type": "text/plain",
	}
	return httpRequest(host, "POST", path, headers, body)
}

// httpRequest builds and sends HTTP request over TCP
func httpRequest(host, method, path string, headers map[string]string, body string) (*HTTPResponse, error) {
	start := time.Now()

	// Connect via TCP
	conn, err := net.DialTimeout("tcp", host, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	// Set read/write deadline
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	// Build request
	// Request line: METHOD PATH HTTP/1.1
	var req strings.Builder
	req.WriteString(fmt.Sprintf("%s %s HTTP/1.1\r\n", method, path))

	// Host header (required in HTTP/1.1)
	req.WriteString(fmt.Sprintf("Host: %s\r\n", host))

	// User-Agent
	req.WriteString("User-Agent: RawTCPClient/1.0\r\n")

	// Connection header
	req.WriteString("Connection: close\r\n")

	// Content-Length for body
	if body != "" {
		req.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(body)))
	}

	// Custom headers
	for k, v := range headers {
		req.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	// End of headers
	req.WriteString("\r\n")

	// Body
	if body != "" {
		req.WriteString(body)
	}

	// Send request
	_, err = conn.Write([]byte(req.String()))
	if err != nil {
		return nil, fmt.Errorf("write failed: %w", err)
	}

	// Parse response
	return parseResponse(conn, start)
}

// parseResponse reads and parses HTTP response
func parseResponse(conn net.Conn, start time.Time) (*HTTPResponse, error) {
	reader := bufio.NewReader(conn)

	// Read status line: HTTP/1.1 STATUS TEXT
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read status failed: %w", err)
	}
	statusLine = strings.TrimSpace(statusLine)

	// Parse status line
	parts := strings.SplitN(statusLine, " ", 3)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid status line: %s", statusLine)
	}

	statusCode, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid status code: %s", parts[1])
	}

	statusText := ""
	if len(parts) == 3 {
		statusText = parts[2]
	}

	// Read headers
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("read header failed: %w", err)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break // End of headers
		}

		colonIdx := strings.Index(line, ":")
		if colonIdx > 0 {
			key := strings.TrimSpace(line[:colonIdx])
			value := strings.TrimSpace(line[colonIdx+1:])
			headers[strings.ToLower(key)] = value
		}
	}

	// Read body based on Content-Length
	var body string
	if lengthStr, ok := headers["content-length"]; ok {
		length, _ := strconv.Atoi(lengthStr)
		if length > 0 {
			bodyBytes := make([]byte, length)
			n, err := reader.Read(bodyBytes)
			if err != nil && n == 0 {
				return nil, fmt.Errorf("read body failed: %w", err)
			}
			body = string(bodyBytes[:n])
		}
	}

	return &HTTPResponse{
		StatusCode:   statusCode,
		StatusText:   statusText,
		Headers:      headers,
		Body:         body,
		ResponseTime: time.Since(start),
	}, nil
}

// printResponse displays response info
func printResponse(resp *HTTPResponse, truncate bool) {
	fmt.Printf("Status: %d %s\n", resp.StatusCode, resp.StatusText)
	fmt.Printf("Response Time: %v\n", resp.ResponseTime)
	fmt.Println("Headers:")
	for k, v := range resp.Headers {
		fmt.Printf("  %s: %s\n", k, v)
	}

	body := resp.Body
	if truncate && len(body) > 200 {
		body = body[:200] + "... (truncated)"
	}
	fmt.Printf("Body:\n%s\n", body)
}
