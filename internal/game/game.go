package game

type RoomID string

type Game struct {
  Rooms     map[RoomID]*Room
	available []RoomID
}

func NewGame() *Game {
  return &Game {
    Rooms: make(map[RoomID]*Room),
  }
}
