# WebRTC P2P Chat

A peer-to-peer chat application using WebRTC data channels for direct browser-to-browser communication.

## Architecture

```
┌─────────────┐     WebSocket      ┌─────────────┐
│  Browser A  │◄──────────────────►│  Signaling  │
│  (Next.js)  │                    │   Server    │
└──────┬──────┘                    │    (Go)     │
       │                           └──────┬──────┘
       │ WebRTC Data Channel              │
       │ (P2P - no server)                │
       │                           ┌──────┴──────┐
       │                           │  Browser B  │
       └──────────────────────────►│  (Next.js)  │
                                   └─────────────┘
```

## Directory Structure

```
network/webrtc/
├── server/
│   ├── go/          # Go signaling server
│   │   └── main.go
│   └── rust/        # Rust signaling server
│       ├── Cargo.toml
│       └── src/main.rs
└── client/
    └── nextjs/      # React client
        └── src/
            ├── app/
            │   ├── page.tsx        # Text chat
            │   └── voice/page.tsx  # Voice chat
            └── components/
                ├── WebRTCChat.tsx   # Text chat component
                └── VoiceChat.tsx    # Voice chat component
```

## How It Works

### What are SDP and ICE?

**SDP (Session Description Protocol)** - Describes "what" to connect:
- Media capabilities (codecs, formats)
- Data channel configuration
- Security parameters (encryption keys)
- Think of it as a "business card" with your capabilities

**ICE (Interactive Connectivity Establishment)** - Describes "how" to connect:
- Network paths to reach the peer (IP addresses, ports)
- Multiple candidates: local, server-reflexive (via STUN), relay (via TURN)
- Browsers test each path to find the best one
- Think of it as "here are all the ways you can reach me"

```
SDP Offer:  "I support opus audio, VP8 video, data channels..."
ICE:        "You can reach me at 192.168.1.5:54321 or 203.0.113.5:12345..."
```

### Why WebSocket for Signaling?

Before browsers can establish a direct P2P connection, they need to exchange SDP and ICE candidates. This requires **server-to-browser push** - the server must push messages to browsers when they arrive from other peers.

- **HTTP** - Request/response only, browser must poll
- **WebSocket** - Bidirectional, server can push anytime

### Signaling Server = Dumb Relay (Like a Network Hub)

The signaling server acts like an **IP router or network hub in broadcast mode**:

- Receives messages from one client
- Forwards to all other clients (excluding sender)
- **Doesn't understand or modify the payload**
- Just routes based on connection info

```
Client A ──┐
           │
Client B ──┼── Hub ── broadcasts to all except sender
           │
Client C ──┘
```

**What it does NOT do:**
- Send client IDs to browsers
- Parse or understand WebRTC offers/answers
- Manage any WebRTC logic

### Connection Flow

```
Browser A                    Server                    Browser B
   │                           │                           │
   │──[connect]───────────────>│                           │
   │                           │<──[connect]───────────────│
   │                           │                           │
   │──[SDP offer]─────────────>│                           │
   │                           │──[SDP offer]─────────────>│
   │                           │                           │
   │                           │<──[SDP answer]────────────│
   │<──[SDP answer]────────────│                           │
   │                           │                           │
   │──[ICE candidates]────────>│──[ICE candidates]────────>│
   │<──[ICE candidates]────────│<──[ICE candidates]────────│
   │                           │                           │
   │<========== P2P Data Channel (no server) =============>│
```

1. **Connect Signalling** - Both browsers connect via WebSocket
2. **Start as Offerer** - Browser A creates SDP offer, server relays to B
3. **Answer** - Browser B creates SDP answer, server relays to A
4. **ICE Exchange** - Both exchange network path candidates
5. **P2P Established** - Direct connection, server no longer needed

### Data Channel

Once WebRTC connection is established:

- Named "chat" in this implementation
- Messages flow directly between browsers (no server)
- Low latency, direct peer connection

## Running the App

### Prerequisites

- Go 1.21+
- Node.js 18+
- npm

### Start Signaling Server

**Go:**
```bash
cd network/webrtc/server/go
go run main.go
# Server starts on :8080
```

**Rust:**
```bash
cd network/webrtc/server/rust
cargo run
# Server starts on :8080
```

### Start Next.js Client

```bash
cd network/webrtc/client/nextjs
npm install
npm run dev
# Client starts on :3000
```

## Usage

### Text Chat (http://localhost:3000)

1. Open in **two browser windows**
2. Click **"Connect Signalling"** in both windows
3. Click **"Start as Offerer"** in **only one** window
4. Wait for "DataChannel open" in both logs
5. Type messages and click Send (or press Enter)

### Voice Chat (http://localhost:3000/voice)

1. Open in **two browser windows**
2. Click **"Connect"** in both (allow microphone access)
3. Click **"Start Call"** in **only one** window
4. Start talking!
5. Use **Mute/Unmute** and **Hang Up** buttons as needed

## Code Explanation

### Go Signaling Server (`server/go/main.go`)

```go
// Key structures
type message struct {
    data   []byte
    sender *websocket.Conn
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan message)
```

**Key behaviors:**

- Accepts WebSocket connections on `/ws`
- Broadcasts messages to all clients **except the sender**
- Simple relay - no message parsing or validation

### React Client (`client/nextjs/src/components/WebRTCChat.tsx`)

**State management:**
```typescript
const pcRef = useRef<RTCPeerConnection | null>(null);  // WebRTC connection
const dcRef = useRef<RTCDataChannel | null>(null);     // Data channel
const wsRef = useRef<WebSocket | null>(null);          // Signaling WebSocket
```

**Connection flow:**

1. `setup()` - Creates WebSocket and RTCPeerConnection
2. `makeOffer()` - Initiates connection (offerer only)
3. `onmessage` handler - Processes offer/answer/ICE from peer
4. `sendMsg()` - Sends chat message via data channel

**ICE candidate handling:**
```typescript
pc.onicecandidate = (e) => {
    if (e.candidate) {
        wsRef.current?.send(JSON.stringify({ candidate: e.candidate }));
    }
};
```

## WebRTC Concepts

| Term | Description |
|------|-------------|
| **SDP** | Session Description Protocol - describes media/connection |
| **Offer** | Initial connection proposal from peer A |
| **Answer** | Response to offer from peer B |
| **ICE** | Interactive Connectivity Establishment |
| **ICE Candidate** | Potential network path (IP/port) |
| **STUN** | Server to discover public IP (uses Google's) |
| **Data Channel** | Low-level bidirectional data stream |

## Future Plans

- [x] Rust signaling server implementation
- [ ] Multiple chat rooms
- [ ] File transfer support
- [ ] Video/audio streaming
