package game

const (
	ballSpeed = 5
)

type Ball struct {
	Y  int `json:"y"`
	X  int `json:"x"`
	Dy int `json:"dy"`
	Dx int `json:"dx"`
}

func NewBall() *Ball {
	return &Ball{
		X:  boardWidth / 2,
		Y:  boardHeight / 2,
		Dx: -ballSpeed,
		Dy: 3,
	}
}

func (ball *Ball) move() {
	newY := ball.Y + ball.Dy
	newX := ball.X + ball.Dx

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
