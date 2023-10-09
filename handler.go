package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "template/index.html")
}

func createRoom(rooms Rooms, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	room := newRoom()
	rooms.rooms[room.id] = room
	rooms.available = append(rooms.available, room.id)

	newUrl := fmt.Sprintf("/rooms/%s", room.id)

	http.Redirect(w, r, newUrl, http.StatusSeeOther)
}

func serveRoom(rooms Rooms, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
  roomID := roomID(path.Base(r.URL.Path))
  if _, ok := rooms.rooms[roomID]; !ok {
    http.Error(w, "Room not found", http.StatusNotFound)
    return
  }

	http.ServeFile(w, r, "template/room.html")
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	room := newRoom()
	client := &Client{room: &room, conn: conn, send: make(chan []byte)}
	client.room.join <- client

	go client.writePump()
	go client.readPump()
}

func newHandler(rooms Rooms) *http.ServeMux {
	handler := http.NewServeMux()
	handler.HandleFunc("/", serveHome)
	handler.HandleFunc("/create-room", func(w http.ResponseWriter, r *http.Request) {
		createRoom(rooms, w, r)
	})
	handler.HandleFunc("/rooms/", func(w http.ResponseWriter, r *http.Request) {
		serveRoom(rooms, w, r)
	})
	handler.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(w, r)
	})
	return handler
}
