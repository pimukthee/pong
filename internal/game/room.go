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

	boardWidth   = 750
	boardHeight  = 585
	grid         = 15
	playerHeight = grid * 5
	maxHeight    = boardHeight - grid - playerHeight

	left  = Seat(false)
	right = Seat(true)
)

type Seat bool

type gameState struct {
	Player1 PlayerState `json:"player1"`
	Player2 PlayerState `json:"player2"`
	Ball    Ball        `json:"ball"`
}

type Room struct {
	ID           RoomID
	Status       int
	PlayersCount int
	Players      [2]*Player
	Ball         *Ball
	Broadcast    chan gameState
	Join         chan *Player
	Leave        chan *Player
}

func NewRoom() Room {
	return Room{
		ID:        RoomID(uuid.NewString()),
		Status:    waiting,
		Ball:      NewBall(),
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
			if r.Status == ready {
				begin <- struct{}{}
			}

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
			if r.Status == ready {
				r.Players[0].updatePosition()
				r.Players[1].updatePosition()
        r.Ball.move()

        r.Broadcast <- gameState{Player1: r.Players[0].GetState(), Player2: r.Players[1].GetState(), Ball: *r.Ball}
			}
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

	i := r.getAvailableSeat()
	r.Players[i] = p

	if r.PlayersCount == 2 {
		r.Status = ready
	}
}

func (r *Room) getAvailableSeat() int {
	if r.Players[0] == nil {
		return 0
	}

	return 1
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
