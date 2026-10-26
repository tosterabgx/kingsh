package chess

type Move struct{ From, To Square }

type Game struct {
	Board    Board
	Turn     Color
	Last     *Move
	Captured [2][]Piece
}

func NewGame() *Game {
	return &Game{Board: StartingBoard(), Turn: White}
}
