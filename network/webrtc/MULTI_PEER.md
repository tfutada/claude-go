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
  const [callingPeers, setCallingPeers] = useState<Set<string>>(new Set());

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
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: true,
        audio: true
      });
      localStreamRef.current = stream;
      if (localVideoRef.current) {
        localVideoRef.current.srcObject = stream;
      }
    } catch (err) {
      console.error('Failed to get media devices:', err);
      alert('Camera/microphone access denied');
      return;
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
          // Existing peers when we joined (no auto-connect)
          setPeers(msg.peers);
          break;

        case 'peer-joined':
          // New peer joined - add to list
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
          setCallingPeers(prev => {
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
      setCallingPeers(prev => {
        const next = new Set(prev);
        next.delete(peerId);
        return next;
      });
    };

    pc.onconnectionstatechange = () => {
      if (pc.connectionState === 'connected') {
        setConnectedPeers(prev => new Set([...prev, peerId]));
        setCallingPeers(prev => {
          const next = new Set(prev);
          next.delete(peerId);
          return next;
        });
      } else if (pc.connectionState === 'disconnected' || pc.connectionState === 'failed') {
        setConnectedPeers(prev => {
          const next = new Set(prev);
          next.delete(peerId);
          return next;
        });
        setCallingPeers(prev => {
          const next = new Set(prev);
          next.delete(peerId);
          return next;
        });
      }
    };

    peerConnectionsRef.current.set(peerId, { pc });
    return pc;
  };

  const callPeer = async (peerId: string) => {
    if (connectedPeers.has(peerId) || callingPeers.has(peerId)) return;

    setCallingPeers(prev => new Set([...prev, peerId]));

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

  const hangUp = (peerId: string) => {
    const peerConn = peerConnectionsRef.current.get(peerId);
    if (peerConn) {
      peerConn.pc.close();
      peerConnectionsRef.current.delete(peerId);
    }
    setConnectedPeers(prev => {
      const next = new Set(prev);
      next.delete(peerId);
      return next;
    });
    setCallingPeers(prev => {
      const next = new Set(prev);
      next.delete(peerId);
      return next;
    });
  };

  const handleOffer = async (fromPeerId: string, sdp: string) => {
    // Check if we already have a connection (e.g., we also called them)
    let peerConn = peerConnectionsRef.current.get(fromPeerId);
    if (peerConn) {
      // Already have connection - this is a glare condition
      // Use peer ID comparison to decide who wins (lower ID is offerer)
      if (myPeerId < fromPeerId) {
        // We should be the offerer, ignore their offer
        return;
      }
      // They should be the offerer, close our connection and accept theirs
      peerConn.pc.close();
      peerConnectionsRef.current.delete(fromPeerId);
    }

    // Auto-accept incoming calls
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

      {/* Peer List with Call Buttons */}
      <div style={{ marginBottom: '20px' }}>
        <h3>Peers in Room ({peers.length})</h3>
        {peers.length === 0 ? (
          <p>No other peers connected</p>
        ) : (
          <table style={{ borderCollapse: 'collapse', width: '100%', maxWidth: '500px' }}>
            <thead>
              <tr>
                <th style={{ textAlign: 'left', padding: '8px', borderBottom: '1px solid #ccc' }}>Peer ID</th>
                <th style={{ textAlign: 'left', padding: '8px', borderBottom: '1px solid #ccc' }}>Status</th>
                <th style={{ textAlign: 'left', padding: '8px', borderBottom: '1px solid #ccc' }}>Action</th>
              </tr>
            </thead>
            <tbody>
              {peers.map(peerId => (
                <tr key={peerId}>
                  <td style={{ padding: '8px', borderBottom: '1px solid #eee' }}>
                    {peerId}
                  </td>
                  <td style={{
                    padding: '8px',
                    borderBottom: '1px solid #eee',
                    color: connectedPeers.has(peerId) ? 'green' : callingPeers.has(peerId) ? 'orange' : 'gray'
                  }}>
                    {connectedPeers.has(peerId) ? 'Connected' : callingPeers.has(peerId) ? 'Calling...' : 'Available'}
                  </td>
                  <td style={{ padding: '8px', borderBottom: '1px solid #eee' }}>
                    {connectedPeers.has(peerId) ? (
                      <button
                        onClick={() => hangUp(peerId)}
                        style={{
                          padding: '4px 12px',
                          backgroundColor: '#dc3545',
                          color: 'white',
                          border: 'none',
                          borderRadius: '4px',
                          cursor: 'pointer'
                        }}
                      >
                        Hang Up
                      </button>
                    ) : callingPeers.has(peerId) ? (
                      <button
                        onClick={() => hangUp(peerId)}
                        style={{
                          padding: '4px 12px',
                          backgroundColor: '#6c757d',
                          color: 'white',
                          border: 'none',
                          borderRadius: '4px',
                          cursor: 'pointer'
                        }}
                      >
                        Cancel
                      </button>
                    ) : (
                      <button
                        onClick={() => callPeer(peerId)}
                        style={{
                          padding: '4px 12px',
                          backgroundColor: '#28a745',
                          color: 'white',
                          border: 'none',
                          borderRadius: '4px',
                          cursor: 'pointer'
                        }}
                      >
                        Call
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
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

        {/* Remote videos - only show connected peers */}
        {Array.from(connectedPeers).map(peerId => (
          <div key={peerId}>
            <p>{peerId}</p>
            <video
              ref={el => {
                if (el) {
                  remoteVideosRef.current.set(peerId, el);
                  // Set stream if already available
                  const peerConn = peerConnectionsRef.current.get(peerId);
                  if (peerConn?.stream) {
                    el.srcObject = peerConn.stream;
                  }
                }
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

## Frame Exchange with 3 Peers (Manual Call Buttons)

```
═══════════════════════════════════════════════════════════════════
PHASE 1: All Peers Join Room (No Auto-Connect)
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

UI State at this point:
┌─────────────────────────────────────────────┐
│ Browser A sees:                             │
│ ┌─────────┬───────────┬─────────┐           │
│ │ Peer ID │ Status    │ Action  │           │
│ ├─────────┼───────────┼─────────┤           │
│ │ B       │ Available │ [Call]  │           │
│ │ C       │ Available │ [Call]  │           │
│ └─────────┴───────────┴─────────┘           │
└─────────────────────────────────────────────┘

═══════════════════════════════════════════════════════════════════
PHASE 2: A clicks "Call" on B (Manual Action)
═══════════════════════════════════════════════════════════════════

Browser A → Server:
┌──────────────────────────────────────┐
│ { "type": "offer",                   │
│   "from": "A", "to": "B",            │
│   "sdp": "..." }                     │
└──────────────────────────────────────┘

Server → Browser B (targeted):
┌──────────────────────────────────────┐
│ { "type": "offer",                   │
│   "from": "A", "to": "B",            │
│   "sdp": "..." }                     │
└──────────────────────────────────────┘

Browser B auto-accepts and sends answer:
┌──────────────────────────────────────┐
│ { "type": "answer",                  │
│   "from": "B", "to": "A",            │
│   "sdp": "..." }                     │
└──────────────────────────────────────┘

Server → Browser A (targeted):
┌──────────────────────────────────────┐
│ { "type": "answer",                  │
│   "from": "B", "to": "A",            │
│   "sdp": "..." }                     │
└──────────────────────────────────────┘

ICE candidates exchanged...

UI State after A-B connected:
┌─────────────────────────────────────────────┐
│ Browser A sees:                             │
│ ┌─────────┬───────────┬───────────┐         │
│ │ Peer ID │ Status    │ Action    │         │
│ ├─────────┼───────────┼───────────┤         │
│ │ B       │ Connected │ [Hang Up] │         │
│ │ C       │ Available │ [Call]    │         │
│ └─────────┴───────────┴───────────┘         │
└─────────────────────────────────────────────┘

═══════════════════════════════════════════════════════════════════
PHASE 3: C clicks "Call" on A (Manual Action)
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

Browser A auto-accepts and sends answer:
┌──────────────────────────────────────┐
│ { "type": "answer",                  │
│   "from": "A", "to": "C",            │
│   "sdp": "..." }                     │
└──────────────────────────────────────┘

ICE candidates exchanged...

═══════════════════════════════════════════════════════════════════
PHASE 4: C clicks "Call" on B (Manual Action)
═══════════════════════════════════════════════════════════════════

Browser C → Server:
┌──────────────────────────────────────┐
│ { "type": "offer",                   │
│   "from": "C", "to": "B",            │
│   "sdp": "..." }                     │
└──────────────────────────────────────┘

... same flow as above ...

═══════════════════════════════════════════════════════════════════
PHASE 5: Full Mesh (Each Peer Connected Individually)
═══════════════════════════════════════════════════════════════════

    A ════════════ B
     ╲           ╱
      ╲   P2P   ╱
       ╲       ╱
        ╲     ╱
           C

Final UI State on Browser C:
┌─────────────────────────────────────────────┐
│ Browser C sees:                             │
│ ┌─────────┬───────────┬───────────┐         │
│ │ Peer ID │ Status    │ Action    │         │
│ ├─────────┼───────────┼───────────┤         │
│ │ A       │ Connected │ [Hang Up] │         │
│ │ B       │ Connected │ [Hang Up] │         │
│ └─────────┴───────────┴───────────┘         │
│                                             │
│ Videos: [Me] [A] [B]                        │
└─────────────────────────────────────────────┘
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

- Peers see list of available peers with Call/Hang Up buttons
- Click "Call" to initiate connection to specific peer
- Incoming offers are auto-accepted (no reject UI)
- Each peer can have multiple independent P2P connections
- Server only routes messages, doesn't modify them
- Videos only appear for connected peers
- Clean disconnect sends `peer-left` to all
