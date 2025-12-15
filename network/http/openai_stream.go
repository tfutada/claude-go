// OpenAI streaming API example - demonstrates SSE parsing.
//
// Uses chunked transfer encoding with SSE format.
// Each chunk contains: data: {"choices":[{"delta":{"content":"token"}}]}
//
// Usage:
//   export OPENAI_API_KEY=sk-...
//   export OPENAI_API_BASE=https://api.openai.com/v1  # optional, default
//   go run ./network/http/openai_stream.go
//
// Azure OpenAI:
//   export OPENAI_API_KEY=your-azure-key
//   export OPENAI_API_BASE=https://{resource}.openai.azure.com/openai/deployments/{deployment}?api-version=2024-02-15-preview
//   go run ./network/http/openai_stream.go
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}


func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: OPENAI_API_KEY not set")
		os.Exit(1)
	}

	apiBase := os.Getenv("OPENAI_API_BASE")
	if apiBase == "" {
		apiBase = "https://api.openai.com/v1"
	}
	endpoint := apiBase + "/chat/completions"

	// Build request
	reqBody := ChatRequest{
		Model: "gpt-4o-mini",
		Messages: []Message{
			{Role: "user", Content: "Count from 1 to 5 slowly, one number per line."},
		},
		Stream: true,
	}

	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("api-key", apiKey) // Azure OpenAI uses this header

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("Transfer-Encoding: %s\n\n", resp.Header.Get("Transfer-Encoding"))

	if resp.StatusCode != http.StatusOK {
		fmt.Println("API error:", resp.Status)
		os.Exit(1)
	}

	// Dump raw SSE body
	fmt.Println("=== Raw SSE Body ===")
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			fmt.Println("[empty line]")
		} else {
			fmt.Println(line)
		}
	}
	fmt.Println("=== End ===")
}
