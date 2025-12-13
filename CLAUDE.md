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
