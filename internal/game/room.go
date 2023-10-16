package game

import (
	"fmt"
	"sync"
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

type command struct {
	operation string
	data      any
}

type Seat bool

type gameState struct {
	Player1      Player  `json:"player1"`
	Player2      Player  `json:"player2"`
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
	IsPrivate    bool
	turn         int
	mutex        sync.Mutex
	stateCh      chan command
	ready        chan struct{}
	pause        chan struct{}
	aborts       [2]chan struct{}
	game         *Game
}

func NewRoom(g *Game, isPrivate bool) *Room {
	room := Room{
		ID:        RoomID(uuid.NewString()),
		Status:    waiting,
		Broadcast: make(chan Message, 1),
		Join:      make(chan *Player),
		Leave:     make(chan *Player),
		IsPrivate: isPrivate,
		stateCh:   make(chan command),
		ready:     make(chan struct{}),
		pause:     make(chan struct{}),
		game:      g,
	}
	room.Ball = NewBall(&room)

	return &room
}

func (r *Room) IsWaiting() bool {
	return r.Status == waiting
}

func (r *Room) Run() {
	defer func() {
		delete(r.game.Rooms, r.ID)
		delete(r.game.Available, r.ID)
		fmt.Println("clear room")
	}()

	for i := 0; i < 2; i++ {
		r.aborts[i] = make(chan struct{})
	}

	go r.stateManager()
	go r.statusManager()

	for loop := true; loop; {
		select {
		case player := <-r.Join:
			r.stateCh <- command{operation: "add player", data: player}

		case player := <-r.Leave:
			r.stateCh <- command{operation: "remove player", data: player}

		case message := <-r.Broadcast:
			players := r.Players[:]
			for i := range players {
				if players[i] == nil {
					continue
				}
				select {
				case players[i].Send <- message:
				default:
					r.stateCh <- command{operation: "remove player", data: players[i]}
				}
			}

		case <-r.aborts[0]:
			loop = false
		}
	}
}

func (r *Room) statusManager() {
	for cmd := range r.stateCh {
		c := cmd.operation
		data := cmd.data

		switch c {
		case "add player":
			r.addPlayer(data.(*Player))
		case "remove player":
			player := data.(*Player)
			r.removePlayer(player)
			close(player.Send)
			if r.isEmpty() {
				r.aborts[0] <- struct{}{}
        r.aborts[1] <- struct{}{}
				return
			}
		case "update status":
			r.Status = data.(int)
		}
	}
}

func (r *Room) stateManager() {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop() 

	for {
		if r.Status == waiting {
			select {
			case <-r.ready:
				r.initRoom()
				r.stateCh <- command{operation: "update status", data: pause}
			case <-r.aborts[1]:
				return
			}
		}
		select {
		case <-ticker.C:
			r.mutex.Lock()
			if r.Status == waiting {
				r.mutex.Unlock()
				continue
			}

			r.Players[0].updatePosition()
			r.Players[1].updatePosition()

			var scoredPlayer *Player
			var shouldEnd bool
			if r.Ball.move() {
				scoredPlayer = r.Ball.getScoredPlayer()
				shouldEnd = r.shouldEnd(scoredPlayer)
				r.resetAfterScore(scoredPlayer)
			}

			msg := Message{
				Type: "update",
				Data: gameState{Player1: *r.Players[0], Player2: *r.Players[1], Ball: *r.Ball, ScoredPlayer: scoredPlayer},
			}

			r.Broadcast <- msg

			if shouldEnd {
				r.endRound(scoredPlayer)
			}
			r.mutex.Unlock()
		case <-r.pause:
			r.Ball.IsMoving = true
		case <-r.aborts[1]:
			return
		}
	}
}

func (r *Room) initRoom() {
	r.reset()
}

func (r *Room) shouldEnd(scoredPlayer *Player) bool {
	return scoredPlayer.isMaxScoreReached()
}

func (r *Room) endRound(winner *Player) {
	r.stateCh <- command{operation: "update status", data: finish}

	r.reset()

	msg := Message{
		Type: "finish",
		Data: winner,
	}
	r.Broadcast <- msg
}

func (r *Room) reset() {
	for i := range r.Players {
		if r.Players[i] != nil {
			r.Players[i].reset()
		}
	}
	r.turn = 0
	r.Ball.reset(Seat(right))
}

func (r *Room) resetAfterScore(scoredPlayer *Player) {
	r.stateCh <- command{operation: "update status", data: pause}

	for i := range r.Players {
		r.Players[i].resetPosition()
	}

	r.turn = 1
	if scoredPlayer.Seat == right {
		r.turn = 0
	}
	r.Ball.reset(scoredPlayer.Seat)
}

func (r *Room) isEmpty() bool {
	return r.PlayersCount == 0
}

func (r *Room) addPlayer(p *Player) {
	r.PlayersCount += 1

	i := r.getAvailableSeat()
	r.Players[i] = p

	if r.PlayersCount == 2 {
		r.game.Available[r.ID] = false
		r.Status = ready
		r.ready <- struct{}{}
		r.Broadcast <- Message{
			Type: "ready",
		}
	}
}

func (r *Room) getAvailableSeat() int {
	if r.Players[0] == nil {
		return 0
	}

	return 1
}

func (r *Room) removePlayer(p *Player) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.PlayersCount -= 1

	if p == r.Players[0] {
		r.Players[0] = nil
	} else {
		r.Players[1] = nil
	}

	if r.PlayersCount > 0 {
		if !r.IsPrivate {
			r.game.Available[r.ID] = true
		}
	}

	r.Broadcast <- Message{
		Type: "leave",
	}

	r.Status = waiting
}
