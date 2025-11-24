'use client';

import React, { useEffect, useRef, useState } from 'react';

export default function WebRTCChatPeers() {
  const pcRef = useRef<RTCPeerConnection | null>(null);
  const dcRef = useRef<RTCDataChannel | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  const [log, setLog] = useState<string[]>([]);
  const [msg, setMsg] = useState('');

  const logLine = (s: string) => setLog((prev) => [...prev, s]);

  const setup = () => {
    if (pcRef.current) return;

    // WebSocket signalling
    wsRef.current = new WebSocket('ws://localhost:8080/ws');
    wsRef.current.onmessage = async (ev) => {
      const data = JSON.parse(ev.data);

      const pc = pcRef.current!;
      if (data.type === 'offer') {
        await pc.setRemoteDescription(data);
        const answer = await pc.createAnswer();
        await pc.setLocalDescription(answer);
        wsRef.current?.send(JSON.stringify(answer));
      } else if (data.type === 'answer') {
        await pc.setRemoteDescription(data);
      } else if (data.candidate) {
        await pc.addIceCandidate(data.candidate);
      }
    };

    // Create WebRTC peer connection
    const pc = new RTCPeerConnection({
      iceServers: [{ urls: 'stun:stun.l.google.com:19302' }],
    });
    pcRef.current = pc;

    // ICE candidate -> send to signalling
    pc.onicecandidate = (e) => {
      if (e.candidate) {
        wsRef.current?.send(JSON.stringify({ candidate: e.candidate }));
      }
    };

    // Data channel (if we are the initiator)
    const dc = pc.createDataChannel('chat');
    dcRef.current = dc;

    dc.onopen = () => logLine('DataChannel open');
    dc.onmessage = (ev) => logLine('Peer: ' + ev.data);

    // If the other peer creates the DataChannel
    pc.ondatachannel = (ev) => {
      const d = ev.channel;
      d.onopen = () => logLine('DataChannel open (remote)');
      d.onmessage = (data) => logLine('Peer: ' + data.data);
      dcRef.current = d;
    };
  };

  // Browser A: create offer manually
  const makeOffer = async () => {
    if (!pcRef.current) {
      logLine('Error: Click "Connect Signalling" first');
      return;
    }
    const pc = pcRef.current;
    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);
    wsRef.current?.send(JSON.stringify(offer));
    logLine('Sent offer');
  };

  const sendMsg = () => {
    dcRef.current?.send(msg);
    logLine('Me: ' + msg);
    setMsg('');
  };

  const [connected, setConnected] = useState(false);

  const handleSetup = () => {
    setup();
    setConnected(true);
    logLine('Connected to signalling server');
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white p-8">
      <div className="max-w-2xl mx-auto">
        <h1 className="text-3xl font-bold mb-2">P2P WebRTC Chat</h1>

        <div className="bg-gray-800 rounded p-4 mb-6 text-sm">
          <p className="font-semibold text-yellow-400 mb-2">How to use:</p>
          <ol className="list-decimal list-inside space-y-1 text-gray-300">
            <li>Open this page in <strong>two browser windows</strong></li>
            <li>Click <strong>"Connect Signalling"</strong> in both windows</li>
            <li>Click <strong>"Start as Offerer"</strong> in <em>only one</em> window</li>
            <li>Wait for "DataChannel open" in both logs</li>
            <li>Start chatting!</li>
          </ol>
        </div>

        <div className="flex gap-3 mb-6">
          <button
            onClick={handleSetup}
            disabled={connected}
            className={`px-4 py-2 rounded font-medium transition ${
              connected
                ? 'bg-green-600 cursor-not-allowed'
                : 'bg-blue-600 hover:bg-blue-700'
            }`}
          >
            {connected ? 'Connected' : 'Connect Signalling'}
          </button>
          <button
            onClick={makeOffer}
            disabled={!connected}
            className={`px-4 py-2 rounded font-medium transition ${
              connected
                ? 'bg-purple-600 hover:bg-purple-700'
                : 'bg-gray-600 cursor-not-allowed'
            }`}
          >
            Start as Offerer
          </button>
        </div>

        <div className="flex gap-2 mb-6">
          <input
            value={msg}
            onChange={(e) => setMsg(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && sendMsg()}
            placeholder="Type a message..."
            className="flex-1 px-4 py-2 rounded bg-gray-800 border border-gray-700 focus:border-blue-500 focus:outline-none"
          />
          <button
            onClick={sendMsg}
            className="px-6 py-2 bg-blue-600 hover:bg-blue-700 rounded font-medium transition"
          >
            Send
          </button>
        </div>

        <div>
          <h3 className="text-lg font-semibold mb-2">Log</h3>
          <pre className="bg-black rounded p-4 h-80 overflow-y-auto text-sm font-mono text-green-400">
            {log.length ? log.join('\n') : 'Waiting for connection...'}
          </pre>
        </div>
      </div>
    </div>
  );
}
