package main

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	none = iota
	waiting
	ready
	start
	finish
)

type roomID string

type Room struct {
	id        roomID
	status    int
	players   [2]*Player
	broadcast chan []byte
	join      chan *Player
	leave     chan *Player
}

func newRoom() Room {
	return Room{
		id:        roomID(uuid.NewString()),
		status:    waiting,
		broadcast: make(chan []byte),
		join:      make(chan *Player),
		leave:     make(chan *Player),
	}
}

func (r *Room) IsWaiting() bool {
	return r.status == waiting
}

func (r *Room) run() {
	for {
		select {
		case player := <-r.join:
			r.addPlayer(player)
		case player := <-r.leave:
			r.removePlayer(player)
			close(player.send)
		case message := <-r.broadcast:
			for i, player := range r.players {
				select {
				case player.send <- message:
				default:
					close(player.send)
					r.players[i] = nil
				}
			}
		}
	}
}

func (r *Room) addPlayer(p *Player) {
	if r.players[0] == nil {
		r.players[0] = p
		return
	}
	if r.players[1] == nil {
		r.players[1] = p
		r.status = ready
	}
}

func (r *Room) removePlayer(p *Player) {
	if p == r.players[0] {
		r.players[0] = nil
	} else {
		r.players[1] = nil
	}
	if r.status == ready {
		r.status = waiting
	} else if r.status == start {
		r.status = finish
	}
}
