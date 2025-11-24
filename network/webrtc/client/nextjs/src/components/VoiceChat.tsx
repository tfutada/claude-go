'use client';

import React, { useRef, useState } from 'react';

export default function VoiceChat() {
  const pcRef = useRef<RTCPeerConnection | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const remoteAudioRef = useRef<HTMLAudioElement | null>(null);
  const localStreamRef = useRef<MediaStream | null>(null);

  const [log, setLog] = useState<string[]>([]);
  const [connected, setConnected] = useState(false);
  const [inCall, setInCall] = useState(false);
  const [muted, setMuted] = useState(false);

  const logLine = (s: string) => setLog((prev) => [...prev, s]);

  const setup = async () => {
    if (pcRef.current) return;

    // Get microphone access
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      localStreamRef.current = stream;
      logLine('Microphone access granted');
    } catch (err) {
      logLine('Error: Microphone access denied');
      return;
    }

    // WebSocket signalling
    wsRef.current = new WebSocket('ws://localhost:8080/ws');
    wsRef.current.onopen = () => {
      setConnected(true);
      logLine('Connected to signalling server');
    };
    wsRef.current.onmessage = async (ev) => {
      const data = JSON.parse(ev.data);

      const pc = pcRef.current!;
      if (data.type === 'offer') {
        await pc.setRemoteDescription(data);
        const answer = await pc.createAnswer();
        await pc.setLocalDescription(answer);
        wsRef.current?.send(JSON.stringify(answer));
        logLine('Received offer, sent answer');
      } else if (data.type === 'answer') {
        await pc.setRemoteDescription(data);
        logLine('Received answer');
      } else if (data.candidate) {
        await pc.addIceCandidate(data.candidate);
      }
    };

    // Create WebRTC peer connection
    const pc = new RTCPeerConnection({
      iceServers: [{ urls: 'stun:stun.l.google.com:19302' }],
    });
    pcRef.current = pc;

    // Add local audio track to connection
    localStreamRef.current!.getTracks().forEach((track) => {
      pc.addTrack(track, localStreamRef.current!);
    });

    // ICE candidate -> send to signalling
    pc.onicecandidate = (e) => {
      if (e.candidate) {
        wsRef.current?.send(JSON.stringify({ candidate: e.candidate }));
      }
    };

    // Receive remote audio
    pc.ontrack = (e) => {
      logLine('Receiving remote audio');
      if (remoteAudioRef.current) {
        remoteAudioRef.current.srcObject = e.streams[0];
      }
      setInCall(true);
    };

    pc.onconnectionstatechange = () => {
      logLine(`Connection state: ${pc.connectionState}`);
      if (pc.connectionState === 'connected') {
        setInCall(true);
      } else if (pc.connectionState === 'disconnected' || pc.connectionState === 'failed') {
        setInCall(false);
      }
    };
  };

  const makeOffer = async () => {
    if (!pcRef.current) {
      logLine('Error: Click "Connect" first');
      return;
    }
    const pc = pcRef.current;
    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);
    wsRef.current?.send(JSON.stringify(offer));
    logLine('Sent offer');
  };

  const toggleMute = () => {
    if (localStreamRef.current) {
      const audioTrack = localStreamRef.current.getAudioTracks()[0];
      if (audioTrack) {
        audioTrack.enabled = !audioTrack.enabled;
        setMuted(!audioTrack.enabled);
        logLine(audioTrack.enabled ? 'Unmuted' : 'Muted');
      }
    }
  };

  const hangUp = () => {
    if (pcRef.current) {
      pcRef.current.close();
      pcRef.current = null;
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    if (localStreamRef.current) {
      localStreamRef.current.getTracks().forEach((track) => track.stop());
      localStreamRef.current = null;
    }
    setConnected(false);
    setInCall(false);
    logLine('Call ended');
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white p-8">
      <div className="max-w-2xl mx-auto">
        <h1 className="text-3xl font-bold mb-2">P2P Voice Chat</h1>

        <div className="bg-gray-800 rounded p-4 mb-6 text-sm">
          <p className="font-semibold text-yellow-400 mb-2">How to use:</p>
          <ol className="list-decimal list-inside space-y-1 text-gray-300">
            <li>Open this page in <strong>two browser windows</strong></li>
            <li>Click <strong>"Connect"</strong> in both (allow microphone)</li>
            <li>Click <strong>"Start Call"</strong> in <em>only one</em> window</li>
            <li>Start talking!</li>
          </ol>
        </div>

        <div className="flex gap-3 mb-6">
          <button
            onClick={setup}
            disabled={connected}
            className={`px-4 py-2 rounded font-medium transition ${
              connected
                ? 'bg-green-600 cursor-not-allowed'
                : 'bg-blue-600 hover:bg-blue-700'
            }`}
          >
            {connected ? 'Connected' : 'Connect'}
          </button>
          <button
            onClick={makeOffer}
            disabled={!connected || inCall}
            className={`px-4 py-2 rounded font-medium transition ${
              connected && !inCall
                ? 'bg-purple-600 hover:bg-purple-700'
                : 'bg-gray-600 cursor-not-allowed'
            }`}
          >
            Start Call
          </button>
          {inCall && (
            <>
              <button
                onClick={toggleMute}
                className={`px-4 py-2 rounded font-medium transition ${
                  muted
                    ? 'bg-yellow-600 hover:bg-yellow-700'
                    : 'bg-gray-600 hover:bg-gray-700'
                }`}
              >
                {muted ? 'Unmute' : 'Mute'}
              </button>
              <button
                onClick={hangUp}
                className="px-4 py-2 rounded font-medium transition bg-red-600 hover:bg-red-700"
              >
                Hang Up
              </button>
            </>
          )}
        </div>

        {inCall && (
          <div className="bg-green-900/50 rounded p-4 mb-6 text-center">
            <span className="text-green-400 font-semibold">
              🎤 Call in progress {muted && '(Muted)'}
            </span>
          </div>
        )}

        <audio ref={remoteAudioRef} autoPlay />

        <div>
          <h3 className="text-lg font-semibold mb-2">Log</h3>
          <pre className="bg-black rounded p-4 h-60 overflow-y-auto text-sm font-mono text-green-400">
            {log.length ? log.join('\n') : 'Waiting for connection...'}
          </pre>
        </div>
      </div>
    </div>
  );
}
