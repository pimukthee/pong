package main

import (
	"log"
	"net/http"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func newHandler(hub *Hub) *http.ServeMux {
	handler := http.NewServeMux()
	handler.HandleFunc("/", serveHome)
	handler.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	return handler
}
