// gateway_errors.go demonstrates 502 Bad Gateway and 504 Gateway Timeout errors
//
// Architecture:
//   Client -> Reverse Proxy (:8080) -> Upstream Server (:9090)
//
// 502 Bad Gateway: Upstream returns invalid/malformed response
// 504 Gateway Timeout: Upstream takes too long to respond
//
// Run: go run gateway_errors.go
// Test:
//   curl http://localhost:8080/normal    # 200 OK
//   curl http://localhost:8080/slow      # 504 Gateway Timeout
//   curl http://localhost:8080/crash     # 502 Bad Gateway

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func main() {
	// Start upstream server
	go startUpstreamServer()
	time.Sleep(100 * time.Millisecond)

	// Start reverse proxy
	startReverseProxy()
}

// Upstream server simulates various failure scenarios
func startUpstreamServer() {
	mux := http.NewServeMux()

	// Normal response
	mux.HandleFunc("/normal", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello from upstream!"))
	})

	// Slow response - causes 504 Gateway Timeout
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[Upstream] Sleeping for 10 seconds...")
		time.Sleep(10 * time.Second)
		w.Write([]byte("Finally done!"))
	})

	// Crash - causes 502 Bad Gateway (connection reset)
	mux.HandleFunc("/crash", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[Upstream] Crashing connection...")
		// Get underlying connection and close it abruptly
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
			return
		}
		conn, _, err := hijacker.Hijack()
		if err != nil {
			return
		}
		// Close without sending any response - causes 502
		conn.Close()
	})

	// Invalid response - also causes 502
	mux.HandleFunc("/invalid", func(w http.ResponseWriter, r *http.Request) {
		log.Println("[Upstream] Sending malformed response...")
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			return
		}
		conn, _, err := hijacker.Hijack()
		if err != nil {
			return
		}
		// Write garbage that's not valid HTTP
		conn.Write([]byte("THIS IS NOT A VALID HTTP RESPONSE\r\n\r\n"))
		conn.Close()
	})

	log.Println("[Upstream] Starting on :9090")
	http.ListenAndServe(":9090", mux)
}

// Reverse proxy with timeout settings
func startReverseProxy() {
	upstream, _ := url.Parse("http://localhost:9090")

	proxy := httputil.NewSingleHostReverseProxy(upstream)

	// Custom transport with short timeout for demonstration
	proxy.Transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ResponseHeaderTimeout: 3 * time.Second, // Timeout waiting for response headers
	}

	// Custom error handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("[Proxy] Error: %v", err)

		// Determine error type
		if isTimeout(err) {
			// 504 Gateway Timeout
			w.WriteHeader(http.StatusGatewayTimeout)
			fmt.Fprintf(w, "504 Gateway Timeout: upstream server took too long\nError: %v\n", err)
		} else {
			// 502 Bad Gateway
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintf(w, "502 Bad Gateway: upstream server returned invalid response\nError: %v\n", err)
		}
	}

	// Wrap proxy to log requests
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Proxy] Request: %s %s", r.Method, r.URL.Path)
		proxy.ServeHTTP(w, r)
	})

	log.Println("[Proxy] Starting reverse proxy on :8080")
	log.Println("")
	log.Println("Test commands:")
	log.Println("  curl http://localhost:8080/normal   # 200 OK")
	log.Println("  curl http://localhost:8080/slow     # 504 Gateway Timeout (wait 3s)")
	log.Println("  curl http://localhost:8080/crash    # 502 Bad Gateway")
	log.Println("  curl http://localhost:8080/invalid  # 502 Bad Gateway")
	log.Println("")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}

// isTimeout checks if error is a timeout error
func isTimeout(err error) bool {
	if err == nil {
		return false
	}

	// Check for context deadline exceeded
	if err.Error() == "context deadline exceeded" {
		return true
	}

	// Check for net.Error timeout
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}

	// Check wrapped errors
	if unwrapped := err.(interface{ Unwrap() error }); unwrapped != nil {
		return isTimeout(unwrapped.Unwrap())
	}

	// Check error message for timeout indicators
	errStr := err.Error()
	if contains(errStr, "timeout") || contains(errStr, "deadline") {
		return true
	}

	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Alternative: Simple demonstration without reverse proxy
func simpleDemo() {
	// Direct server that returns 502/504
	mux := http.NewServeMux()

	mux.HandleFunc("/502", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		io.WriteString(w, "502 Bad Gateway\n")
	})

	mux.HandleFunc("/504", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGatewayTimeout)
		io.WriteString(w, "504 Gateway Timeout\n")
	})

	http.ListenAndServe(":8080", mux)
}
