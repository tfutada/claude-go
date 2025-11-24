package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[*websocket.Conn]bool)

type message struct {
	data   []byte
	sender *websocket.Conn
}

var broadcast = make(chan message)

func main() {
	http.HandleFunc("/ws", handleWS)

	go relay()

	log.Println("Signalling server on :8080")
	http.ListenAndServe(":8080", nil)
}

func handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	clients[conn] = true
	log.Println("client connected")

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("error reading:", err)
			delete(clients, conn)
			conn.Close()
			return
		}
		// broadcast SDP/ICE to all other clients
		broadcast <- message{data: msg, sender: conn}
	}
}

func relay() {
	for {
		msg := <-broadcast
		for c := range clients {
			// Don't send back to sender
			if c != msg.sender {
				c.WriteMessage(websocket.TextMessage, msg.data)
			}
		}
	}
}
