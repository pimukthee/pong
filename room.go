package main

import "github.com/google/uuid"

const (
	None = iota
	Waiting
	Start
)

type roomID string

type Room struct {
	id        roomID
	status    int
	players   map[*Player]bool
	broadcast chan []byte
	join      chan *Player
	leave     chan *Player
}

func newRoom() Room {
	return Room{
		id:        roomID(uuid.NewString()),
		status:    Waiting,
		broadcast: make(chan []byte),
		join:      make(chan *Player),
		leave:     make(chan *Player),
		players:   make(map[*Player]bool),
	}
}

func (r *Room) run() {
	for {
		select {
		case client := <-r.join:
			r.players[client] = true
		case client := <-r.leave:
			if _, ok := r.players[client]; ok {
				delete(r.players, client)
				close(client.send)
			}
		case message := <-r.broadcast:
			for client := range r.players {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(r.players, client)
				}
			}
		}
	}
}
