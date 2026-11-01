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

func (g *Game) Move(from, to Square) bool {
	p := g.Board[from]
	if p.Empty() || p.Color != g.Turn || !g.CanMove(from, to) {
		return false
	}

	if target := g.Board[to]; !target.Empty() {
		g.Captured[g.Turn] = append(g.Captured[g.Turn], target)
	}

	// TODO: promotion
	if p.Kind == Pawn && (to.Rank() == 0 || to.Rank() == 7) {
		p.Kind = Queen
	}

	g.Board[to], g.Board[from] = p, Piece{}
	g.Last = &Move{from, to}
	g.Turn = g.Turn.Other()
	return true
}
