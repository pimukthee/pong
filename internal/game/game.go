package game

import (
	"log"
	"net/http"
	"path"

	"github.com/gorilla/websocket"
)

const (
  maxScore = 1
)

type RoomID string

type Game struct {
	Rooms     map[RoomID]*Room
	available []RoomID
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Message struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

type InitMessage struct {
	ID   PlayerID `json:"id"`
	Seat Seat     `json:"seat"`
}

func NewGame() *Game {
	return &Game{
		Rooms: make(map[RoomID]*Room),
	}
}

func ServeWs(g *Game, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	roomID := RoomID(path.Base(r.URL.Path))
	room, ok := g.Rooms[roomID]
	if !ok {
		log.Println("room not found")
		return
	}

	if !room.IsWaiting() {
		log.Println("room is not available")
		return
	}

	player := NewPlayer(room, conn)
	player.Room.Join <- player

	// send player id back to the client
	conn.WriteJSON(Message{
		Type: "init",
		Data: InitMessage{
			ID:   player.ID,
			Seat: player.Seat,
		},
	})

	go player.WritePump()
	go player.ReadAction()
}
