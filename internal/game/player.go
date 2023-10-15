package game

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10

	acceleration = 3
	maxSpeed     = 6
)

type Action struct {
	Up     bool `json:"up"`
	Down   bool `json:"down"`
	Start  bool `json:"start"`
	Replay bool `json:"replay"`
}

type PlayerID string

type Player struct {
	ID     PlayerID        `json:"id"`
	Room   *Room           `json:"-"`
	Conn   *websocket.Conn `json:"-"`
	Send   chan Message    `json:"-"`
	Action Action          `json:"-"`
	Seat   Seat            `json:"seat"`
	width  int             `json:"-"`
	height int             `json:"-"`
	Score  int             `json:"score"`
	X      int             `json:"x"`
	Y      int             `json:"y"`
	Dy     int             `json:"dy"`
}

func NewPlayer(room *Room, conn *websocket.Conn) *Player {
	seat := left
	if room.getAvailableSeat() == 1 {
		seat = right
	}

	player := &Player{
		ID:     PlayerID(uuid.NewString()),
		Room:   room,
		Conn:   conn,
		Send:   make(chan Message, 1),
		Seat:   seat,
		Y:      boardHeight/2 - playerHeight/2,
		width:  grid,
		height: playerHeight,
	}

	if seat == right {
		player.X = boardWidth - grid*3
	} else {
		player.X = grid * 2
	}

	return player
}

func (p *Player) reset() {
	p.resetPosition()
	p.Score = 0
}

func (p *Player) resetPosition() {
	p.Y = boardHeight/2 - playerHeight/2
	p.Dy = 0
}

func (p *Player) isMaxScoreReached() bool {
	return p.Score >= maxScore
}

func (p *Player) ReadAction() {
	defer func() {
		p.Room.Leave <- p
		p.Conn.Close()
	}()

	p.Conn.SetReadDeadline(time.Now().Add(pongWait))
	p.Conn.SetPongHandler(func(string) error {
		p.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		err := p.Conn.ReadJSON(&p.Action)
		if err != nil {
			log.Println(err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		if p.Action.Replay && (p.Room.Status == ready || p.Room.Status == finish) {
			p.Room.Status = pause
			p.Room.Broadcast <- Message{
				Type: "replay",
			}

			continue
		}

		if p.Action.Start && (p.Room.Status == ready || p.Room.Status == pause) {
			if p.Room.Status == ready {
				p.Room.pause <- struct{}{}
			}
			p.Room.Status = start
		}
	}
}

func (p *Player) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		p.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-p.Send:

			p.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				p.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := p.Conn.WriteJSON(msg)
			if err != nil {
				log.Println(err)
				return
			}

		case <-ticker.C:
			p.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := p.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func (p *Player) updatePosition() {
	p.updateSpeed()
	p.Y = p.move()
}

func (p *Player) move() int {
	newY := p.Y + p.Dy
	if newY <= maxHeight && newY >= grid {
		return newY
	}

	return p.Y
}

func (p *Player) updateSpeed() {
	if p.Action.Up && !p.Action.Down {
		p.pressUp()
	} else if p.Action.Down && !p.Action.Up {
		p.pressDown()
	} else if !p.Action.Up && !p.Action.Down {
		p.stop()
	}
}

func (p *Player) pressUp() {
	p.Dy = max(-maxSpeed, p.Dy-acceleration)
}

func (p *Player) pressDown() {
	p.Dy = min(maxSpeed, p.Dy+acceleration)
}

func (p *Player) stop() {
	if p.Dy > 0 {
		p.Dy -= acceleration
	} else if p.Dy < 0 {
		p.Dy += acceleration
	}
}
