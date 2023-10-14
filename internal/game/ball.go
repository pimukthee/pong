package game

import "fmt"

const (
	ballSpeed = 5
)

type Ball struct {
	room   *Room `json:"-"`
	Y      int   `json:"y"`
	X      int   `json:"x"`
	Dy     int   `json:"dy"`
	Dx     int   `json:"dx"`
	width  int   `json:"-"`
	height int   `json:"-"`
}

func NewBall(room *Room) *Ball {
	return &Ball{
		room: room,
		X:    boardWidth / 2,
		Y:    boardHeight / 2,
		Dx:   -ballSpeed,
		Dy:   3,
    width: grid,
    height: grid,
	}
}

func (ball *Ball) move() {
	newY := ball.Y + ball.Dy
	newX := ball.X + ball.Dx

	player1 := ball.room.Players[0]
	player2 := ball.room.Players[1]

	if ball.checkCollision(player1) {
		ball.Dx *= -1
		ball.X = player1.X + player1.width

		return
	}
	if ball.checkCollision(player2) {
		ball.Dx *= -1
		ball.X = player2.X - ball.width

		return
	}

	// bounce horizontal boundaries
	if newY >= grid && newY <= boardHeight-2*grid {
		ball.Y = newY
		ball.X = newX
	} else {
		ball.Dy *= -1
		ball.X = newX
	}

	// bounce vertical boundaries
	if newX >= 0 && newX <= boardWidth-grid {
		ball.Y = newY
		ball.X = newX
	} else {
		ball.Dx *= -1
		ball.Y = newY
	}
}
func (ball *Ball) checkCollision(player *Player) bool {
	return (ball.X < player.X+player.width &&
		ball.X+ball.width > player.X &&
		ball.Y < player.Y+player.height &&
		ball.Y+ball.height > player.Y)
}
