package game

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

type Room struct {
	ID           RoomID
	Status       int
	PlayersCount int
	Players      [2]*Player
	Broadcast    chan []byte
	Join         chan *Player
	Leave        chan *Player
}

func NewRoom() Room {
	return Room{
		ID:        RoomID(uuid.NewString()),
		Status:    waiting,
		Broadcast: make(chan []byte),
		Join:      make(chan *Player),
		Leave:     make(chan *Player),
	}
}

func (r *Room) IsWaiting() bool {
	return r.Status == waiting
}

func (r *Room) Run(ctx context.Context) {
	defer func(rooms map[RoomID]*Room, roomId RoomID) {
		delete(rooms, roomId)
		fmt.Println("clear room")
	}(ctx.Value("game").(map[RoomID]*Room), r.ID)

  for loop := true; loop; {
		select {
		case player := <-r.Join:
			r.addPlayer(player)

		case player := <-r.Leave:
			r.removePlayer(player)
			close(player.Send)
			loop = !r.isEmpty()

		case message := <-r.Broadcast:
			players := r.Players[:]
			for i, player := range players {
				if player == nil {
					continue
				}

				select {
				case player.Send <- message:
				default:
					close(player.Send)
					players[i] = nil
				}
			}
		}
	}
}

func (r *Room) isEmpty() bool {
	return r.PlayersCount == 0
}

func (r *Room) addPlayer(p *Player) {
	r.PlayersCount += 1

	if r.Players[0] == nil {
		r.Players[0] = p
		return
	}
	if r.Players[1] == nil {
		r.Players[1] = p
		r.Status = ready
	}
}

func (r *Room) removePlayer(p *Player) {
	r.PlayersCount -= 1

	if p == r.Players[0] {
		r.Players[0] = nil
	} else {
		r.Players[1] = nil
	}
	if r.Status == ready {
		r.Status = waiting
	} else if r.Status == start {
		r.Status = finish
	}
}
