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

func serveWs(room *Room, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{room: room, conn: conn, send: make(chan []byte)}
	client.room.register <- client

	go client.writePump()
	go client.readPump()
}

func newHandler(room *Room) *http.ServeMux {
	handler := http.NewServeMux()
	handler.HandleFunc("/", serveHome)
	handler.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(room, w, r)
	})
	return handler
}
