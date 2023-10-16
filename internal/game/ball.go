package game

import (
	"math"
)

const (
	ballSpeed      = 10
	maxBounceAngle = 5 * math.Pi / 12

	leftSpeed  = -1
	rightSpeed = 1
)

type Ball struct {
	room     *Room `json:"-"`
	Y        int   `json:"y"`
	X        int   `json:"x"`
	Dy       int   `json:"dy"`
	Dx       int   `json:"dx"`
	IsMoving bool  `json:"isMoving"`
	width    int   `json:"-"`
	height   int   `json:"-"`
}

func NewBall(room *Room) *Ball {
	return &Ball{
		room:   room,
		X:      boardWidth/2 - grid/2,
		Y:      boardHeight/2 - grid/2,
		Dx:     -ballSpeed,
		Dy:     0,
		width:  grid,
		height: grid,
	}
}

func (ball *Ball) reset(currentDirection Seat) {
	ball.IsMoving = false

	if currentDirection == right {
		player := ball.room.Players[0]
		ball.X = player.X + player.width + 1
		ball.Y = player.Y + player.height/2 - ball.height/2
		ball.Dx = ballSpeed
		ball.Dy = 0

		return
	}

	player := ball.room.Players[1]
	ball.X = player.X - ball.width - 1
	ball.Y = player.Y + player.height/2 - ball.height/2
	ball.Dx = -ballSpeed
	ball.Dy = 0
}

func (ball *Ball) move() bool {
	room := ball.room
	if !ball.IsMoving {
		if room.Players[room.turn] == nil {
			ball.Dy = 0
		} else {
			ball.Dy = room.Players[room.turn].Dy
		}
		ball.Y += ball.Dy
		return false
	}

	newY := ball.Y + ball.Dy
	newX := ball.X + ball.Dx

	player1 := ball.room.Players[0]
	player2 := ball.room.Players[1]

	if ball.checkCollision(player1) {
		ball.adjustSpeedAfterCollideWithPaddle(player1)
		ball.X = player1.X + player1.width

		return false
	}
	if ball.checkCollision(player2) {
		ball.adjustSpeedAfterCollideWithPaddle(player2)
		ball.X = player2.X - ball.width

		return false
	}

	if newX < 0 {
		ball.room.Players[1].Score++

		return true
	}
	if newX > boardWidth-ball.width {
		ball.room.Players[0].Score++

		return true
	}

	// bounce horizontal boundaries
	if newY >= grid && newY <= boardHeight-2*grid {
		ball.Y = newY
		ball.X = newX
	} else {
		ball.Dy *= -1
		ball.X = newX
	}

	ball.X = newX
	ball.Y = newY

	return false
}

func (ball *Ball) getScoredPlayer() *Player {
	newX := ball.X + ball.Dx
	if newX < 0 {
		return ball.room.Players[1]
	}
	if newX > boardWidth-ball.width {
		return ball.room.Players[0]
	}

	return nil
}

func (ball *Ball) adjustSpeedAfterCollideWithPaddle(player *Player) {
	bounceAngle := ball.calculateBounceAngle(player)

	ball.Dx = int(float64(ballSpeed) * math.Cos(bounceAngle))
	ball.Dy = int(float64(ballSpeed) * -math.Sin(bounceAngle))
	if ball.Dx == 0 {
		ball.Dx = 1
	}

	if player.Seat == right {
		ball.Dx *= -1
	}
}

func (ball *Ball) calculateBounceAngle(player *Player) float64 {
	relativeIntersectY := (float64(player.Y) + float64(player.height)/2.0) - (float64(ball.Y) + float64(ball.height/2.0))
	normalized := relativeIntersectY / (float64(player.height) / 2.0)

	return normalized * maxBounceAngle
}

func (ball *Ball) checkCollision(player *Player) bool {
	return (ball.X < player.X+player.width &&
		ball.X+ball.width > player.X &&
		ball.Y < player.Y+player.height &&
		ball.Y+ball.height > player.Y)
}
