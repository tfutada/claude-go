# Multi-Peer WebRTC Implementation

## Overview

Current implementation broadcasts to all peers. This document outlines changes needed for proper multi-peer support with peer targeting.

## Changes Required

### 1. Go Signaling Server (`server/go/main.go`)

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client with ID
type Client struct {
	ID   string
	Conn *websocket.Conn
}

var (
	clients   = make(map[string]*websocket.Conn) // ID → connection
	clientsMu sync.RWMutex
)

// SignalMessage with peer targeting
type SignalMessage struct {
	Type      string `json:"type"`                // offer, answer, candidate, join, peer-joined, peer-left, peer-list
	From      string `json:"from,omitempty"`      // sender ID
	To        string `json:"to,omitempty"`        // target peer ID
	PeerID    string `json:"peerId,omitempty"`    // for join/peer-joined/peer-left
	Peers     []string `json:"peers,omitempty"`   // for peer-list
	SDP       string `json:"sdp,omitempty"`
	Candidate any    `json:"candidate,omitempty"`
}

type message struct {
	data   []byte
	sender string
}

var broadcast = make(chan message, 256)

func main() {
	http.HandleFunc("/ws", handleWS)

	go relay()

	log.Println("Signaling server on :8080")
	http.ListenAndServe(":8080", nil)
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	// Wait for join message to get peer ID
	_, msg, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return
	}

	var joinMsg SignalMessage
	if err := json.Unmarshal(msg, &joinMsg); err != nil || joinMsg.Type != "join" {
		conn.Close()
		return
	}

	peerID := joinMsg.PeerID
	log.Printf("client joined: %s", peerID)

	// Send existing peers to new client
	clientsMu.RLock()
	existingPeers := make([]string, 0, len(clients))
	for id := range clients {
		existingPeers = append(existingPeers, id)
	}
	clientsMu.RUnlock()

	peerListMsg, _ := json.Marshal(SignalMessage{
		Type:  "peer-list",
		Peers: existingPeers,
	})
	conn.WriteMessage(websocket.TextMessage, peerListMsg)

	// Add new client
	clientsMu.Lock()
	clients[peerID] = conn
	clientsMu.Unlock()

	// Notify others about new peer
	joinNotify, _ := json.Marshal(SignalMessage{
		Type:   "peer-joined",
		PeerID: peerID,
	})
	broadcast <- message{data: joinNotify, sender: peerID}

	// Read loop
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("client disconnected: %s", peerID)
			break
		}
		broadcast <- message{data: msg, sender: peerID}
	}

	// Cleanup
	clientsMu.Lock()
	delete(clients, peerID)
	clientsMu.Unlock()
	conn.Close()

	// Notify others about peer leaving
	leaveNotify, _ := json.Marshal(SignalMessage{
		Type:   "peer-left",
		PeerID: peerID,
	})
	broadcast <- message{data: leaveNotify, sender: peerID}
}

func relay() {
	for msg := range broadcast {
		var signal SignalMessage
		if err := json.Unmarshal(msg.data, &signal); err != nil {
			continue
		}

		clientsMu.RLock()
		if signal.To != "" {
			// Targeted message: send to specific peer
			if conn, ok := clients[signal.To]; ok {
				conn.WriteMessage(websocket.TextMessage, msg.data)
			}
		} else {
			// Broadcast to all others
			for id, conn := range clients {
				if id != msg.sender {
					conn.WriteMessage(websocket.TextMessage, msg.data)
				}
			}
		}
		clientsMu.RUnlock()
	}
}
```

### 2. React Client (`client/nextjs/src/app/page.tsx`)

```tsx
'use client';

import { useEffect, useRef, useState } from 'react';

interface PeerConnection {
  pc: RTCPeerConnection;
  stream?: MediaStream;
}

export default function Home() {
  const [myPeerId] = useState(() => Math.random().toString(36).substr(2, 9));
  const [peers, setPeers] = useState<string[]>([]);
  const [connectedPeers, setConnectedPeers] = useState<Set<string>>(new Set());

  const wsRef = useRef<WebSocket | null>(null);
  const localStreamRef = useRef<MediaStream | null>(null);
  const peerConnectionsRef = useRef<Map<string, PeerConnection>>(new Map());

  const localVideoRef = useRef<HTMLVideoElement>(null);
  const remoteVideosRef = useRef<Map<string, HTMLVideoElement>>(new Map());

  useEffect(() => {
    initWebSocket();
    return () => {
      wsRef.current?.close();
      localStreamRef.current?.getTracks().forEach(t => t.stop());
      peerConnectionsRef.current.forEach(pc => pc.pc.close());
    };
  }, []);

  const initWebSocket = async () => {
    // Get local media first
    const stream = await navigator.mediaDevices.getUserMedia({
      video: true,
      audio: true
    });
    localStreamRef.current = stream;
    if (localVideoRef.current) {
      localVideoRef.current.srcObject = stream;
    }

    // Connect to signaling server
    const ws = new WebSocket('ws://localhost:8080/ws');
    wsRef.current = ws;

    ws.onopen = () => {
      // Send join message with our ID
      ws.send(JSON.stringify({ type: 'join', peerId: myPeerId }));
    };

    ws.onmessage = async (event) => {
      const msg = JSON.parse(event.data);

      switch (msg.type) {
        case 'peer-list':
          // Existing peers when we joined
          setPeers(msg.peers);
          // Create offers to all existing peers
          for (const peerId of msg.peers) {
            await createOffer(peerId);
          }
          break;

        case 'peer-joined':
          // New peer joined - add to list, they will send us an offer
          setPeers(prev => [...prev, msg.peerId]);
          break;

        case 'peer-left':
          // Peer left - cleanup
          setPeers(prev => prev.filter(id => id !== msg.peerId));
          setConnectedPeers(prev => {
            const next = new Set(prev);
            next.delete(msg.peerId);
            return next;
          });
          const pc = peerConnectionsRef.current.get(msg.peerId);
          if (pc) {
            pc.pc.close();
            peerConnectionsRef.current.delete(msg.peerId);
          }
          break;

        case 'offer':
          await handleOffer(msg.from, msg.sdp);
          break;

        case 'answer':
          await handleAnswer(msg.from, msg.sdp);
          break;

        case 'candidate':
          await handleCandidate(msg.from, msg.candidate);
          break;
      }
    };
  };

  const createPeerConnection = (peerId: string): RTCPeerConnection => {
    const pc = new RTCPeerConnection({
      iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
    });

    // Add local tracks
    localStreamRef.current?.getTracks().forEach(track => {
      pc.addTrack(track, localStreamRef.current!);
    });

    // Handle ICE candidates
    pc.onicecandidate = (event) => {
      if (event.candidate) {
        wsRef.current?.send(JSON.stringify({
          type: 'candidate',
          from: myPeerId,
          to: peerId,
          candidate: event.candidate
        }));
      }
    };

    // Handle remote stream
    pc.ontrack = (event) => {
      const [stream] = event.streams;
      const peerConn = peerConnectionsRef.current.get(peerId);
      if (peerConn) {
        peerConn.stream = stream;
      }
      // Update video element
      const videoEl = remoteVideosRef.current.get(peerId);
      if (videoEl) {
        videoEl.srcObject = stream;
      }
      setConnectedPeers(prev => new Set([...prev, peerId]));
    };

    pc.onconnectionstatechange = () => {
      if (pc.connectionState === 'connected') {
        setConnectedPeers(prev => new Set([...prev, peerId]));
      } else if (pc.connectionState === 'disconnected' || pc.connectionState === 'failed') {
        setConnectedPeers(prev => {
          const next = new Set(prev);
          next.delete(peerId);
          return next;
        });
      }
    };

    peerConnectionsRef.current.set(peerId, { pc });
    return pc;
  };

  const createOffer = async (peerId: string) => {
    const pc = createPeerConnection(peerId);
    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);

    wsRef.current?.send(JSON.stringify({
      type: 'offer',
      from: myPeerId,
      to: peerId,
      sdp: offer.sdp
    }));
  };

  const handleOffer = async (fromPeerId: string, sdp: string) => {
    const pc = createPeerConnection(fromPeerId);
    await pc.setRemoteDescription({ type: 'offer', sdp });

    const answer = await pc.createAnswer();
    await pc.setLocalDescription(answer);

    wsRef.current?.send(JSON.stringify({
      type: 'answer',
      from: myPeerId,
      to: fromPeerId,
      sdp: answer.sdp
    }));
  };

  const handleAnswer = async (fromPeerId: string, sdp: string) => {
    const peerConn = peerConnectionsRef.current.get(fromPeerId);
    if (peerConn) {
      await peerConn.pc.setRemoteDescription({ type: 'answer', sdp });
    }
  };

  const handleCandidate = async (fromPeerId: string, candidate: RTCIceCandidateInit) => {
    const peerConn = peerConnectionsRef.current.get(fromPeerId);
    if (peerConn) {
      await peerConn.pc.addIceCandidate(candidate);
    }
  };

  return (
    <div style={{ padding: '20px' }}>
      <h1>WebRTC Multi-Peer</h1>

      {/* My Info */}
      <div style={{ marginBottom: '20px' }}>
        <strong>My ID:</strong> {myPeerId}
      </div>

      {/* Peer List */}
      <div style={{ marginBottom: '20px' }}>
        <h3>Peers in Room ({peers.length})</h3>
        {peers.length === 0 ? (
          <p>No other peers connected</p>
        ) : (
          <ul>
            {peers.map(peerId => (
              <li key={peerId} style={{
                color: connectedPeers.has(peerId) ? 'green' : 'orange'
              }}>
                {peerId} {connectedPeers.has(peerId) ? '(connected)' : '(connecting...)'}
              </li>
            ))}
          </ul>
        )}
      </div>

      {/* Videos */}
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: '10px' }}>
        {/* Local video */}
        <div>
          <p>Me ({myPeerId})</p>
          <video
            ref={localVideoRef}
            autoPlay
            muted
            playsInline
            style={{ width: '300px', background: '#000' }}
          />
        </div>

        {/* Remote videos */}
        {peers.map(peerId => (
          <div key={peerId}>
            <p>{peerId}</p>
            <video
              ref={el => {
                if (el) remoteVideosRef.current.set(peerId, el);
              }}
              autoPlay
              playsInline
              style={{ width: '300px', background: '#000' }}
            />
          </div>
        ))}
      </div>
    </div>
  );
}
```

## Frame Exchange with 3 Peers

```
═══════════════════════════════════════════════════════════════════
PHASE 1: Connection
═══════════════════════════════════════════════════════════════════

Browser A → Server:
┌──────────────────────────────────────┐
│ { "type": "join", "peerId": "A" }    │
└──────────────────────────────────────┘

Server → Browser A:
┌──────────────────────────────────────┐
│ { "type": "peer-list", "peers": [] } │
└──────────────────────────────────────┘

Browser B → Server:
┌──────────────────────────────────────┐
│ { "type": "join", "peerId": "B" }    │
└──────────────────────────────────────┘

Server → Browser B:
┌──────────────────────────────────────┐
│ { "type": "peer-list", "peers": ["A"] }│
└──────────────────────────────────────┘

Server → Browser A:
┌──────────────────────────────────────┐
│ { "type": "peer-joined", "peerId": "B" }│
└──────────────────────────────────────┘

Browser C → Server:
┌──────────────────────────────────────┐
│ { "type": "join", "peerId": "C" }    │
└──────────────────────────────────────┘

Server → Browser C:
┌──────────────────────────────────────┐
│ { "type": "peer-list", "peers": ["A", "B"] }│
└──────────────────────────────────────┘

Server → Browser A:
┌──────────────────────────────────────┐
│ { "type": "peer-joined", "peerId": "C" }│
└──────────────────────────────────────┘

Server → Browser B:
┌──────────────────────────────────────┐
│ { "type": "peer-joined", "peerId": "C" }│
└──────────────────────────────────────┘

═══════════════════════════════════════════════════════════════════
PHASE 2: C creates offers to existing peers (A and B)
═══════════════════════════════════════════════════════════════════

Browser C → Server:
┌──────────────────────────────────────┐
│ { "type": "offer",                   │
│   "from": "C", "to": "A",            │
│   "sdp": "..." }                     │
└──────────────────────────────────────┘

Server → Browser A (targeted):
┌──────────────────────────────────────┐
│ { "type": "offer",                   │
│   "from": "C", "to": "A",            │
│   "sdp": "..." }                     │
└──────────────────────────────────────┘

Browser C → Server:
┌──────────────────────────────────────┐
│ { "type": "offer",                   │
│   "from": "C", "to": "B",            │
│   "sdp": "..." }                     │
└──────────────────────────────────────┘

Server → Browser B (targeted):
┌──────────────────────────────────────┐
│ { "type": "offer",                   │
│   "from": "C", "to": "B",            │
│   "sdp": "..." }                     │
└──────────────────────────────────────┘

═══════════════════════════════════════════════════════════════════
PHASE 3: A and B send answers to C
═══════════════════════════════════════════════════════════════════

Browser A → Server:
┌──────────────────────────────────────┐
│ { "type": "answer",                  │
│   "from": "A", "to": "C",            │
│   "sdp": "..." }                     │
└──────────────────────────────────────┘

Server → Browser C (targeted):
┌──────────────────────────────────────┐
│ { "type": "answer",                  │
│   "from": "A", "to": "C",            │
│   "sdp": "..." }                     │
└──────────────────────────────────────┘

Browser B → Server:
┌──────────────────────────────────────┐
│ { "type": "answer",                  │
│   "from": "B", "to": "C",            │
│   "sdp": "..." }                     │
└──────────────────────────────────────┘

Server → Browser C (targeted):
┌──────────────────────────────────────┐
│ { "type": "answer",                  │
│   "from": "B", "to": "C",            │
│   "sdp": "..." }                     │
└──────────────────────────────────────┘

═══════════════════════════════════════════════════════════════════
PHASE 4: ICE Candidates (targeted to each peer)
═══════════════════════════════════════════════════════════════════

Browser C → Server (for A):
┌──────────────────────────────────────┐
│ { "type": "candidate",               │
│   "from": "C", "to": "A",            │
│   "candidate": {...} }               │
└──────────────────────────────────────┘

Browser C → Server (for B):
┌──────────────────────────────────────┐
│ { "type": "candidate",               │
│   "from": "C", "to": "B",            │
│   "candidate": {...} }               │
└──────────────────────────────────────┘

... (each peer sends candidates to each other peer) ...

═══════════════════════════════════════════════════════════════════
PHASE 5: Mesh Established
═══════════════════════════════════════════════════════════════════

    A ════════════ B
     ╲           ╱
      ╲   P2P   ╱
       ╲       ╱
        ╲     ╱
           C

3 P2P connections: A-B (existing), A-C, B-C
```

## Key Differences from Current Implementation

| Aspect | Current | Multi-Peer |
|--------|---------|------------|
| Client tracking | `map[*Conn]bool` | `map[string]*Conn` |
| Message routing | Broadcast all | Targeted by `to` field |
| Peer awareness | None | `peer-list`, `peer-joined`, `peer-left` |
| Connection init | Implicit | Explicit `join` message |
| UI peer list | None | Display all peers with status |

## Notes

- New peer sends offers to ALL existing peers
- Existing peers wait for offers from new peer
- Each peer maintains N-1 `RTCPeerConnection` objects
- Server only routes messages, doesn't modify them
- Clean disconnect sends `peer-left` to all
