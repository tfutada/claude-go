use std::collections::HashMap;
use std::net::SocketAddr;
use std::sync::{Arc, Mutex};
use std::sync::atomic::{AtomicUsize, Ordering};

use axum::{
    extract::State,
    extract::ws::{WebSocket, WebSocketUpgrade, Message},
    response::IntoResponse,
    routing::get,
    Router,
};
use futures::{StreamExt};
use tokio::sync::mpsc;

static NEXT_CLIENT_ID: AtomicUsize = AtomicUsize::new(1);

#[derive(Clone)]
struct AppState {
    // Map client ID to their mpsc sender
    clients: Arc<Mutex<HashMap<usize, mpsc::UnboundedSender<Message>>>>,
}

#[tokio::main]
async fn main() {
    let state = AppState {
        clients: Arc::new(Mutex::new(HashMap::new())),
    };

    let app = Router::new()
        .route("/ws", get(ws_handler))
        .with_state(state);

    let addr: SocketAddr = "0.0.0.0:8080".parse().unwrap();
    println!("Rust signalling server listening on {addr}");
    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}

// HTTP -> WebSocket upgrade handler
async fn ws_handler(
    ws: WebSocketUpgrade,
    State(state): State<AppState>,
) -> impl IntoResponse {
    ws.on_upgrade(move |socket| handle_socket(socket, state))
}

// Per-connection handler
async fn handle_socket(stream: WebSocket, state: AppState) {
    use tokio_stream::wrappers::UnboundedReceiverStream;

    let client_id = NEXT_CLIENT_ID.fetch_add(1, Ordering::Relaxed);
    let (mut ws_sender, mut ws_receiver) = stream.split();

    // mpsc channel to send messages *to* this client
    let (tx, rx) = mpsc::unbounded_channel::<Message>();
    let rx_stream = UnboundedReceiverStream::new(rx);

    // store sender in global client list
    {
        let mut clients = state.clients.lock().unwrap();
        clients.insert(client_id, tx);
        println!("Client {client_id} connected, total clients: {}", clients.len());
    }

    // Task: forward messages from rx_stream to WebSocket
    let forward_to_ws = tokio::spawn(async move {
        let mut rx_stream = rx_stream;
        while let Some(msg) = rx_stream.next().await {
            if ws_sender.send(msg).await.is_err() {
                break;
            }
        }
    });

    // Read from WebSocket and broadcast to all clients except sender
    while let Some(Ok(msg)) = ws_receiver.next().await {
        match msg {
            Message::Text(_) | Message::Binary(_) => {
                broadcast(&state, msg, client_id).await;
            }
            Message::Close(_) => {
                break;
            }
            _ => {}
        }
    }

    // connection is closing; remove this client
    {
        let mut clients = state.clients.lock().unwrap();
        clients.remove(&client_id);
        println!("Client {client_id} disconnected, total clients: {}", clients.len());
    }
    forward_to_ws.abort();
}

// broadcast a message to all connected clients except sender
async fn broadcast(state: &AppState, msg: Message, sender_id: usize) {
    let mut to_remove = Vec::new();

    let clients = state.clients.lock().unwrap();
    for (&id, client_tx) in clients.iter() {
        // Don't send back to sender
        if id == sender_id {
            continue;
        }
        if client_tx.send(msg.clone()).is_err() {
            // client disconnected, mark for removal
            to_remove.push(id);
        }
    }
    drop(clients);

    // remove disconnected clients
    if !to_remove.is_empty() {
        let mut clients = state.clients.lock().unwrap();
        for id in to_remove {
            clients.remove(&id);
        }
    }
}
