package main

import (
	"context"
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
	id           roomID
	status       int
	playersCount int
	players      [2]*Player
	broadcast    chan []byte
	join         chan *Player
	leave        chan *Player
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

func (r *Room) run(ctx context.Context) {
	defer func(rooms map[roomID]*Room, roomId roomID) {
		delete(rooms, roomId)
		fmt.Println("clear room")
	}(ctx.Value("rooms").(map[roomID]*Room), r.id)

  loop := true
	for loop {
		select {
		case player := <-r.join:
			r.addPlayer(player)

		case player := <-r.leave:
			r.removePlayer(player)
			close(player.send)
      loop = !r.isEmpty()

		case message := <-r.broadcast:
			players := r.players[:]
			for i, player := range players {
				if player == nil {
					continue
				}

				select {
				case player.send <- message:
				default:
					close(player.send)
					players[i] = nil
				}
			}
		}
	}
}

func (r *Room) isEmpty() bool {
	return r.playersCount == 0
}

func (r *Room) addPlayer(p *Player) {
  r.playersCount += 1

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
  r.playersCount -= 1

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
