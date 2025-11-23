// Minimal HTTP/1.1 Server (no standard library http package)
// Demonstrates HTTP protocol fundamentals
//
// HTTP/1.1 characteristics:
// - Text-based protocol
// - Request-response model
// - Headers end with \r\n\r\n
// - Body length via Content-Length or chunked

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
	listener, err := net.Listen("tcp", ":8083")
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Println("HTTP Server listening on :8083")
	fmt.Println("Open http://localhost:8083 in browser")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			continue
		}

		go handleHTTP(conn)
	}
}

func handleHTTP(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Read request line: METHOD PATH HTTP/1.1
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	requestLine = strings.TrimSpace(requestLine)

	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		sendError(conn, 400, "Bad Request")
		return
	}

	method := parts[0]
	path := parts[1]
	// version := parts[2]

	fmt.Printf("[%s] %s %s\n", conn.RemoteAddr(), method, path)

	// Read headers until empty line
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
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

	// Read body if Content-Length present
	var body []byte
	if lengthStr, ok := headers["content-length"]; ok {
		length, _ := strconv.Atoi(lengthStr)
		body = make([]byte, length)
		reader.Read(body)
	}

	// Route request
	switch {
	case method == "GET" && path == "/":
		sendHTML(conn, indexPage())

	case method == "GET" && path == "/api/time":
		sendJSON(conn, fmt.Sprintf(`{"time": "%s"}`, time.Now().Format(time.RFC3339)))

	case method == "POST" && path == "/api/echo":
		sendJSON(conn, fmt.Sprintf(`{"echo": "%s"}`, string(body)))

	case method == "GET" && path == "/headers":
		// Echo back request headers
		var sb strings.Builder
		sb.WriteString("<html><body><h1>Request Headers</h1><pre>")
		for k, v := range headers {
			sb.WriteString(fmt.Sprintf("%s: %s\n", k, v))
		}
		sb.WriteString("</pre></body></html>")
		sendHTML(conn, sb.String())

	default:
		sendError(conn, 404, "Not Found")
	}
}

func sendResponse(conn net.Conn, status int, statusText string, contentType string, body string) {
	response := fmt.Sprintf(
		"HTTP/1.1 %d %s\r\n"+
			"Content-Type: %s\r\n"+
			"Content-Length: %d\r\n"+
			"Connection: close\r\n"+
			"\r\n"+
			"%s",
		status, statusText, contentType, len(body), body,
	)
	conn.Write([]byte(response))
}

func sendHTML(conn net.Conn, body string) {
	sendResponse(conn, 200, "OK", "text/html; charset=utf-8", body)
}

func sendJSON(conn net.Conn, body string) {
	sendResponse(conn, 200, "OK", "application/json", body)
}

func sendError(conn net.Conn, status int, message string) {
	body := fmt.Sprintf("<html><body><h1>%d %s</h1></body></html>", status, message)
	sendResponse(conn, status, message, "text/html; charset=utf-8", body)
}

func indexPage() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>Minimal HTTP Server</title>
    <style>
        body { font-family: monospace; max-width: 600px; margin: 50px auto; padding: 20px; }
        pre { background: #f4f4f4; padding: 10px; overflow-x: auto; }
        button { padding: 10px 20px; margin: 5px; cursor: pointer; }
        #result { margin-top: 20px; }
    </style>
</head>
<body>
    <h1>Minimal HTTP/1.1 Server</h1>
    <p>Built with raw TCP sockets in Go - no net/http package!</p>

    <h2>Endpoints:</h2>
    <ul>
        <li><code>GET /</code> - This page</li>
        <li><code>GET /api/time</code> - Current time as JSON</li>
        <li><code>POST /api/echo</code> - Echo POST body as JSON</li>
        <li><code>GET /headers</code> - Show request headers</li>
    </ul>

    <h2>Try it:</h2>
    <button onclick="getTime()">GET /api/time</button>
    <button onclick="postEcho()">POST /api/echo</button>
    <button onclick="getHeaders()">GET /headers</button>

    <div id="result">
        <h3>Result:</h3>
        <pre id="output">Click a button to test</pre>
    </div>

    <script>
        async function getTime() {
            const res = await fetch('/api/time');
            const data = await res.json();
            document.getElementById('output').textContent = JSON.stringify(data, null, 2);
        }

        async function postEcho() {
            const res = await fetch('/api/echo', {
                method: 'POST',
                body: 'Hello from browser!'
            });
            const data = await res.json();
            document.getElementById('output').textContent = JSON.stringify(data, null, 2);
        }

        async function getHeaders() {
            window.open('/headers', '_blank');
        }
    </script>
</body>
</html>`
}
