# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Info

Go module: `claude-go`

## Build and Run Commands

```bash
# Run
go run .

# Build
go build -o claude-go .

# Test
go test ./...

# Format
go fmt ./...

# Lint (if installed)
golangci-lint run
```

## Architecture

Go examples organized by topic:

- `alignment/` - Struct field alignment and memory padding
- `asm/` - Assembly output analysis with go:noinline
- `channel/` - Channel fundamentals and patterns
- `ducktyping/` - Interface and duck typing patterns
- `embedding/` - Struct embedding examples
- `escape/` - Escape analysis examples
- `fatpointer/` - Interface internal representation
- `goroutine/` - Goroutine scheduler, GMP model, preemption
- `io/` - File I/O and io.Reader patterns
- `network/` - Network programming (TCP, UDP, WebSocket, HTTP)
- `slice/` - Slice internals and behavior

## TCP Server/Client Example

```bash
# Terminal 1: Start server
go run ./network/tcp/server.go

# Terminal 2: Start client
go run ./network/tcp/client.go
```

Server uses `SetReadDeadline` (30s) to prevent hanging on incomplete messages.
Client intentionally omits `\n` to demonstrate timeout behavior.

## OpenAI Streaming Example

```bash
# OpenAI
export OPENAI_API_KEY=sk-...
go run ./network/http/openai_stream.go

# Azure OpenAI
export OPENAI_API_KEY=your-azure-key
export OPENAI_API_BASE=https://{resource}.openai.azure.com/openai/v1
go run ./network/http/openai_stream.go
```

Environment variables:
- `OPENAI_API_KEY` (required) - API key for OpenAI or Azure
- `OPENAI_API_BASE` (optional) - Base URL, defaults to `https://api.openai.com/v1`
