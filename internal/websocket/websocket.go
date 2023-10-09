package websocket

import (
	"log"
	"net/http"
	"path"
	"pong/internal/game"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func ServeWs(g *game.Game, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	roomID := game.RoomID(path.Base(r.URL.Path))
	room, ok := g.Rooms[roomID]
	if !ok {
		log.Println("room not found")
		return
	}

	if !room.IsWaiting() {
		log.Println("room is not available")
		return
	}

	player := &game.Player{Room: room, Conn: conn, Send: make(chan []byte)}
	player.Room.Join <- player

	go player.WritePump()
	go player.ReadPump()
}
