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

func createPrivateRoom(g *game.Game, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	room := game.NewRoom(g, true)
	g.Rooms[room.ID] = room

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
		room := game.NewRoom(g, false)
		g.Rooms[room.ID] = room
		g.Available[room.ID] = true

		newUrl = fmt.Sprintf("/rooms/%s", room.ID)

		go room.Run()
	}

	http.Redirect(w, r, newUrl, http.StatusSeeOther)
}

func noCache(h http.Handler) http.Handler {
  fn := func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Cache-Control", "no-cache, private, max-age=0")
    
    h.ServeHTTP(w, r)
  }

  return http.HandlerFunc(fn)
}

func NewHandler(g *game.Game) *http.ServeMux {
	handler := http.NewServeMux()

	fs := http.FileServer(http.Dir("./web/"))
	handler.Handle("/static/", noCache(http.StripPrefix("/static/", fs)))

	handler.HandleFunc("/", serveHome)
	handler.HandleFunc("/quick-play", func(w http.ResponseWriter, r *http.Request) {
		quickPlay(g, w, r)
	})
	handler.HandleFunc("/create-private-room", func(w http.ResponseWriter, r *http.Request) {
		createPrivateRoom(g, w, r)
	})
	handler.HandleFunc("/rooms/", func(w http.ResponseWriter, r *http.Request) {
		serveRoom(g, w, r)
	})
	handler.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		game.ServeWs(g, w, r)
	})

	return handler
}
