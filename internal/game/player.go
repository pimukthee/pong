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
  maxSpeed = 6
)

type Action struct {
	Up   bool `json:"up"`
	Down bool `json:"down"`
}

type PlayerID string

type Player struct {
	ID     PlayerID
	Room   *Room
	Conn   *websocket.Conn
	Send   chan gameState
	action Action
	Y      int
	Dy     int
}

type PlayerState struct {
	ID PlayerID `json:"id"`
	Y  int      `json:"y"`
	Dy int      `json:"dy"`
}

func NewPlayer(room *Room, conn *websocket.Conn) *Player {
	return &Player{
		ID:   PlayerID(uuid.NewString()),
		Room: room,
		Conn: conn,
		Send: make(chan gameState, 1),
		Y:    height/2 - playerHeight/2,
	}
}

func (p *Player) GetState() PlayerState {
	return PlayerState{
		ID: p.ID,
		Y:  p.Y,
		Dy: p.Dy,
	}
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
		err := p.Conn.ReadJSON(&p.action)
		if err != nil {
			log.Println(err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
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
		case state, ok := <-p.Send:

			p.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				p.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := p.Conn.WriteJSON(state)

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
	if p.action.Up && !p.action.Down {
		p.pressUp()
	} else if p.action.Down && !p.action.Up {
		p.pressDown()
	} else if !p.action.Up && !p.action.Down {
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
