package api

import (
	"fmt"
	"net/http"
	"path"
	"pong/internal/game"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "web/index.html")
}

func createRoom(g *game.Game, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	room := game.NewRoom(g)
	g.Rooms[room.ID] = room
	g.Available[room.ID] = true

	go room.Run()

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

	http.ServeFile(w, r, "web/room.html")
}

func quickPlay(g *game.Game, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roomID, found := g.FindAvailableRoom()
	newUrl := fmt.Sprintf("/rooms/%s", roomID)
	if !found {
		room := game.NewRoom(g)
		g.Rooms[room.ID] = room
		g.Available[room.ID] = true

		newUrl = fmt.Sprintf("/rooms/%s", room.ID)

		go room.Run()
	}

	http.Redirect(w, r, newUrl, http.StatusSeeOther)
}

func NewHandler(g *game.Game) *http.ServeMux {
	handler := http.NewServeMux()

	fs := http.FileServer(http.Dir("./web/"))
	handler.Handle("/static/", http.StripPrefix("/static/", fs))

	handler.HandleFunc("/", serveHome)
	handler.HandleFunc("/quick-play", func(w http.ResponseWriter, r *http.Request) {
		quickPlay(g, w, r)
	})
	handler.HandleFunc("/create-room", func(w http.ResponseWriter, r *http.Request) {
		createRoom(g, w, r)
	})
	handler.HandleFunc("/rooms/", func(w http.ResponseWriter, r *http.Request) {
		serveRoom(g, w, r)
	})
	handler.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		game.ServeWs(g, w, r)
	})

	return handler
}
