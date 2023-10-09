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
	clients   map[*Client]bool
	broadcast chan []byte
	join      chan *Client
	leave     chan *Client
}

func newRoom() Room {
	return Room{
		id:        roomID(uuid.NewString()),
		status:    Start,
		broadcast: make(chan []byte),
		join:      make(chan *Client),
		leave:     make(chan *Client),
		clients:   make(map[*Client]bool),
	}
}

func (r *Room) exist() bool {
	return r.status != None
}

func (r *Room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
		case client := <-r.leave:
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.send)
			}
		case message := <-r.broadcast:
			for client := range r.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(r.clients, client)
				}
			}
		}
	}
}
