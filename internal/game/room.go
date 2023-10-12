package game

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	none = iota
	waiting
	ready
	start
	finish
)

const (
	tickRate     = 60
	tickInterval = time.Second / tickRate

	width        = 750
	height       = 585
	grid         = 15
	playerHeight = grid * 5
	maxHeight    = height - grid - playerHeight
)

type gameState struct {
	Player1 PlayerState `json:"player1"`
	Player2 PlayerState `json:"player2"`
}

type Room struct {
	ID           RoomID
	Status       int
	PlayersCount int
	Players      [2]*Player
	Broadcast    chan gameState
	Join         chan *Player
	Leave        chan *Player
}

func NewRoom() Room {
	return Room{
		ID:        RoomID(uuid.NewString()),
		Status:    waiting,
		Broadcast: make(chan gameState),
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

	done := make(chan struct{})
	begin := make(chan struct{})

	go r.updateState(begin, done)

	for loop := true; loop; {
		select {
		case player := <-r.Join:
			r.addPlayer(player)
			fmt.Println("JOIN")
			begin <- struct{}{}
		case player := <-r.Leave:
			r.removePlayer(player)
			close(player.Send)
			loop = !r.isEmpty()

		case message := <-r.Broadcast:
			players := r.Players[:]
			for i := range players {
				if players[i] == nil {
					continue
				}

				select {
				case players[i].Send <- message:
				default:
					close(players[i].Send)
					players[i] = nil
				}
			}
		}
	}
	done <- struct{}{}
}

func (r *Room) updateState(begin chan struct{}, done chan struct{}) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	<-begin
	for {
		select {
		case <-ticker.C:
			r.Players[0].updatePosition()
			r.Broadcast <- gameState{Player1: r.Players[0].GetState()}
		case <-done:
			return
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
