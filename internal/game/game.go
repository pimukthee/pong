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
	Available map[RoomID]bool
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
    Available: make(map[RoomID]bool),
	}
}

func (g *Game) FindAvailableRoom() (RoomID, bool) {
  for roomID := range g.Available {
    if g.Available[roomID] {
      return roomID, true
    }  
  }

  return "", false
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
