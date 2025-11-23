# TCP vs UDP Network Examples

Educational examples demonstrating the differences between TCP and UDP protocols in Go.

## Key Differences

| Feature | TCP | UDP |
|---------|-----|-----|
| Connection | Connection-oriented (3-way handshake) | Connectionless |
| Reliability | Guaranteed delivery | Best-effort delivery |
| Ordering | Preserves order | No ordering guarantee |
| Data format | Byte stream | Datagrams (messages) |
| Speed | Slower (overhead) | Faster (minimal overhead) |
| Use cases | HTTP, FTP, SSH, email | DNS, streaming, gaming |

## Running the Examples

### TCP Example

```bash
# Terminal 1: Start server
cd tcp && go run server.go

# Terminal 2: Start client
cd tcp && go run client.go
```

### UDP Example

```bash
# Terminal 1: Start server
cd udp && go run server.go

# Terminal 2: Start client
cd udp && go run client.go
```

## Concepts Demonstrated

### TCP (`net/tcp`)
- `net.Listen()` - Create listening socket
- `listener.Accept()` - Accept incoming connections (blocking)
- `conn.Read()/Write()` - Stream-based I/O
- Connection state management
- Deadline/timeout handling

### UDP (`net/udp`)
- `net.ListenUDP()` - Create UDP endpoint
- `conn.ReadFromUDP()` - Receive datagram with sender address
- `conn.WriteToUDP()` - Send datagram to specific address
- No connection state
- Manual timeout handling

## Go Network Programming Notes

```go
// TCP uses net.Conn interface
conn, _ := net.Dial("tcp", "host:port")
conn.Write(data)  // Stream write
conn.Read(buffer) // Stream read

// UDP uses net.UDPConn
conn, _ := net.DialUDP("udp", nil, addr)
conn.Write(data)       // Datagram write
conn.ReadFromUDP(buf)  // Datagram read with sender info
```
