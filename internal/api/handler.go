package api

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"pong/internal/game"
	"pong/internal/websocket"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "template/index.html")
}

func createRoom(g *game.Game, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	room := game.NewRoom()
	g.Rooms[room.ID] = &room
	c := context.WithValue(context.Background(), "game", g.Rooms)
	go room.Run(c)

	newUrl := fmt.Sprintf("/rooms/%s", room.ID)

	http.Redirect(w, r, newUrl, http.StatusSeeOther)
}

func serveRoom(g *game.Game, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomID := game.RoomID(path.Base(r.URL.Path))
	room, ok := g.Rooms[roomID]
	if !ok {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	if !room.IsWaiting() {
		http.Error(w, "Room is not available", http.StatusBadRequest)
		return
	}

	http.ServeFile(w, r, "template/room.html")
}

func NewHandler(game *game.Game) *http.ServeMux {
	handler := http.NewServeMux()

	handler.HandleFunc("/", serveHome)
	handler.HandleFunc("/create-room", func(w http.ResponseWriter, r *http.Request) {
		createRoom(game, w, r)
	})
	handler.HandleFunc("/rooms/", func(w http.ResponseWriter, r *http.Request) {
		serveRoom(game, w, r)
	})
	handler.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
    websocket.ServeWs(game, w, r)
	})

	return handler
}
