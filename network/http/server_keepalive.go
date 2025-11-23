// HTTP/1.1 Server with Keep-Alive support
// Demonstrates persistent connections - multiple requests per TCP connection
//
// Keep-Alive benefits:
// - Avoids TCP handshake overhead per request
// - Reduces latency for multiple requests
// - More efficient resource usage

package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	keepAliveTimeout = 30 * time.Second
	maxRequests      = 100 // Max requests per connection
)

func main() {
	listener, err := net.Listen("tcp", ":8084")
	if err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Println("HTTP Server (Keep-Alive) listening on :8084")
	fmt.Println("Open http://localhost:8084 in browser")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			continue
		}

		go handleHTTPKeepAlive(conn)
	}
}

func handleHTTPKeepAlive(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	reader := bufio.NewReader(conn)
	requestCount := 0

	for {
		// Set read deadline for keep-alive timeout
		conn.SetReadDeadline(time.Now().Add(keepAliveTimeout))

		// Read request line
		requestLine, err := reader.ReadString('\n')
		if err != nil {
			if requestCount > 0 {
				fmt.Printf("[%s] Connection closed after %d requests\n", clientAddr, requestCount)
			}
			return
		}
		requestLine = strings.TrimSpace(requestLine)

		parts := strings.Split(requestLine, " ")
		if len(parts) != 3 {
			sendErrorKA(conn, 400, "Bad Request", false)
			return
		}

		method := parts[0]
		path := parts[1]
		requestCount++

		fmt.Printf("[%s] #%d %s %s\n", clientAddr, requestCount, method, path)

		// Read headers
		headers := make(map[string]string)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}

			colonIdx := strings.Index(line, ":")
			if colonIdx > 0 {
				key := strings.TrimSpace(line[:colonIdx])
				value := strings.TrimSpace(line[colonIdx+1:])
				headers[strings.ToLower(key)] = value
			}
		}

		// Read body if present
		var body []byte
		if lengthStr, ok := headers["content-length"]; ok {
			length, _ := strconv.Atoi(lengthStr)
			body = make([]byte, length)
			reader.Read(body)
		}

		// Check if client wants to close connection
		connectionHeader := strings.ToLower(headers["connection"])
		keepAlive := connectionHeader != "close" && requestCount < maxRequests

		// Route and respond
		switch {
		case method == "GET" && path == "/":
			sendHTMLKA(conn, indexPageKA(), keepAlive)

		case method == "GET" && path == "/api/time":
			sendJSONKA(conn, fmt.Sprintf(`{"time": "%s", "request": %d}`, time.Now().Format(time.RFC3339), requestCount), keepAlive)

		case method == "POST" && path == "/api/echo":
			sendJSONKA(conn, fmt.Sprintf(`{"echo": "%s", "request": %d}`, string(body), requestCount), keepAlive)

		case method == "GET" && path == "/api/stats":
			sendJSONKA(conn, fmt.Sprintf(`{"requests_on_connection": %d, "client": "%s"}`, requestCount, clientAddr), keepAlive)

		default:
			sendErrorKA(conn, 404, "Not Found", keepAlive)
		}

		// Close if client requested or max requests reached
		if !keepAlive {
			fmt.Printf("[%s] Closing connection after %d requests\n", clientAddr, requestCount)
			return
		}
	}
}

func sendResponseKA(conn net.Conn, status int, statusText string, contentType string, body string, keepAlive bool) {
	connectionValue := "keep-alive"
	extraHeaders := fmt.Sprintf("Keep-Alive: timeout=%d, max=%d\r\n", int(keepAliveTimeout.Seconds()), maxRequests)
	if !keepAlive {
		connectionValue = "close"
		extraHeaders = ""
	}

	response := fmt.Sprintf(
		"HTTP/1.1 %d %s\r\n"+
			"Content-Type: %s\r\n"+
			"Content-Length: %d\r\n"+
			"Connection: %s\r\n"+
			"%s"+
			"\r\n"+
			"%s",
		status, statusText, contentType, len(body), connectionValue, extraHeaders, body,
	)
	conn.Write([]byte(response))
}

func sendHTMLKA(conn net.Conn, body string, keepAlive bool) {
	sendResponseKA(conn, 200, "OK", "text/html; charset=utf-8", body, keepAlive)
}

func sendJSONKA(conn net.Conn, body string, keepAlive bool) {
	sendResponseKA(conn, 200, "OK", "application/json", body, keepAlive)
}

func sendErrorKA(conn net.Conn, status int, message string, keepAlive bool) {
	body := fmt.Sprintf("<html><body><h1>%d %s</h1></body></html>", status, message)
	sendResponseKA(conn, status, message, "text/html; charset=utf-8", body, keepAlive)
}

func indexPageKA() string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>HTTP Server with Keep-Alive</title>
    <style>
        body { font-family: monospace; max-width: 600px; margin: 50px auto; padding: 20px; }
        pre { background: #f4f4f4; padding: 10px; overflow-x: auto; }
        button { padding: 10px 20px; margin: 5px; cursor: pointer; }
        #result { margin-top: 20px; }
        .info { background: #e8f4e8; padding: 10px; border-left: 3px solid #4a4; margin: 10px 0; }
    </style>
</head>
<body>
    <h1>HTTP/1.1 Server with Keep-Alive</h1>

    <div class="info">
        <strong>Keep-Alive enabled!</strong><br>
        Multiple requests reuse the same TCP connection.<br>
        Timeout: 30s, Max requests: 100
    </div>

    <h2>Endpoints:</h2>
    <ul>
        <li><code>GET /</code> - This page</li>
        <li><code>GET /api/time</code> - Current time + request count</li>
        <li><code>POST /api/echo</code> - Echo POST body</li>
        <li><code>GET /api/stats</code> - Connection stats</li>
    </ul>

    <h2>Test Keep-Alive:</h2>
    <button onclick="getTime()">GET /api/time</button>
    <button onclick="getStats()">GET /api/stats</button>
    <button onclick="multipleRequests()">Send 5 requests</button>

    <div id="result">
        <h3>Result:</h3>
        <pre id="output">Click a button to test.
Watch the server logs to see request counts on the same connection.</pre>
    </div>

    <script>
        async function getTime() {
            const res = await fetch('/api/time');
            const data = await res.json();
            document.getElementById('output').textContent = JSON.stringify(data, null, 2);
        }

        async function getStats() {
            const res = await fetch('/api/stats');
            const data = await res.json();
            document.getElementById('output').textContent = JSON.stringify(data, null, 2);
        }

        async function multipleRequests() {
            const results = [];
            for (let i = 0; i < 5; i++) {
                const res = await fetch('/api/time');
                const data = await res.json();
                results.push(data);
            }
            document.getElementById('output').textContent =
                'All requests used same TCP connection:\n\n' +
                JSON.stringify(results, null, 2);
        }
    </script>
</body>
</html>`
}
