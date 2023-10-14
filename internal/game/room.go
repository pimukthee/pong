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
	pause
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
	Player1      *Player `json:"player1"`
	Player2      *Player `json:"player2"`
	Ball         Ball    `json:"ball"`
	ScoredPlayer *Player `json:"scoredPlayer"`
}

type Room struct {
	ID           RoomID
	Status       int
	PlayersCount int
	Players      [2]*Player
	Ball         *Ball
	Broadcast    chan Message
	Join         chan *Player
	Leave        chan *Player
	pause        chan struct{}
	done         chan struct{}
}

func NewRoom() *Room {
	room := Room{
		ID:        RoomID(uuid.NewString()),
		Status:    waiting,
		Broadcast: make(chan Message),
		Join:      make(chan *Player),
		Leave:     make(chan *Player),
		pause:     make(chan struct{}),
		done:      make(chan struct{}),
	}
	room.Ball = NewBall(&room)

	return &room
}

func (r *Room) IsWaiting() bool {
	return r.Status == waiting
}

func (r *Room) Run(ctx context.Context) {
	defer func(rooms map[RoomID]*Room, roomId RoomID) {
		delete(rooms, roomId)
		fmt.Println("clear room")
	}(ctx.Value("game").(map[RoomID]*Room), r.ID)

	go r.updateState()

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
			for i := range players {
				select {
				case players[i].Send <- message:
				default:
					close(players[i].Send)
					players[i] = nil
				}
			}
		}
	}
	r.done <- struct{}{}
}

func (r *Room) updateState() {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	<-r.pause

	for {
		select {
		case <-ticker.C:
			if r.Status == start {
				r.Players[0].updatePosition()
				r.Players[1].updatePosition()

				var scoredPlayer *Player
        var shouldEnd bool
				if r.Ball.move() {
					scoredPlayer = r.Ball.getScoredPlayer()
          shouldEnd = r.shouldEnd(scoredPlayer)
					r.reset(scoredPlayer)
				}

				msg := Message{
					Type: "update",
					Data: gameState{Player1: r.Players[0], Player2: r.Players[1], Ball: *r.Ball, ScoredPlayer: scoredPlayer},
				}

				r.Broadcast <- msg

        if shouldEnd {
          r.endRound(scoredPlayer) 
        }

			}
		case <-r.done:
			return
		}
	}
}

func (r *Room) shouldEnd(scoredPlayer *Player) bool {
  return scoredPlayer.isMaxScoreReached()
} 

func (r *Room) endRound(winner *Player) {
  r.Status = finish

  msg := Message {
    Type: "finish",
    Data: winner, 
  }
  r.Broadcast <-msg
}

func (r *Room) reset(scoredPlayer *Player) {
	r.Status = pause
	for i := range r.Players {
		r.Players[i].reset()
	}
	r.Ball.reset(scoredPlayer)
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
