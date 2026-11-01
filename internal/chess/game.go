package chess

type Move struct{ From, To Square }

const (
	KingSide  = 0
	QueenSide = 1
)

type Game struct {
	Board    Board
	Turn     Color
	Last     *Move
	Captured [2][]Piece

	CanCastle [2][2]bool
	EnPassant Square
}

func NewGame() *Game {
	return &Game{
		Board:     StartingBoard(),
		Turn:      White,
		CanCastle: [2][2]bool{{true, true}, {true, true}},
		EnPassant: NoSquare,
	}
}

func (g *Game) IsPromotion(from, to Square) bool {
	return g.Board[from].Kind == Pawn && (to.Rank() == 0 || to.Rank() == 7)
}

func (g *Game) Move(from, to Square) bool {
	return g.MovePromote(from, to, Queen)
}

func (g *Game) MovePromote(from, to Square, promo Kind) bool {
	p := g.Board[from]
	if p.Empty() || p.Color != g.Turn || !g.CanMove(from, to) {
		return false
	}
	if g.IsPromotion(from, to) && (promo < Knight || promo > Queen) {
		return false
	}

	if captured := g.Board.play(from, to, promo); !captured.Empty() {
		g.Captured[g.Turn] = append(g.Captured[g.Turn], captured)
	}
	g.updateCastling(from, to)

	g.EnPassant = NoSquare
	if p.Kind == Pawn && abs(to.Rank()-from.Rank()) == 2 {
		g.EnPassant = Sq(from.File(), (from.Rank()+to.Rank())/2)
	}

	g.Last = &Move{from, to}
	g.Turn = g.Turn.Other()
	return true
}

func (b *Board) play(from, to Square, promo Kind) Piece {
	p := b[from]
	captured := b[to]
	r := from.Rank()

	switch {
	case p.Kind == Pawn && from.File() != to.File() && captured.Empty():
		passed := Sq(to.File(), r)
		captured, b[passed] = b[passed], Piece{}
	case p.Kind == King && to.File()-from.File() == 2:
		b[Sq(5, r)], b[Sq(7, r)] = b[Sq(7, r)], Piece{}
	case p.Kind == King && from.File()-to.File() == 2:
		b[Sq(3, r)], b[Sq(0, r)] = b[Sq(0, r)], Piece{}
	}

	if p.Kind == Pawn && (to.Rank() == 0 || to.Rank() == 7) {
		p.Kind = promo
	}
	b[to], b[from] = p, Piece{}
	return captured
}

func (g *Game) updateCastling(from, to Square) {
	for _, c := range []Color{White, Black} {
		r := 0
		if c == Black {
			r = 7
		}
		if from == Sq(4, r) {
			g.CanCastle[c] = [2]bool{}
		}
		if from == Sq(7, r) || to == Sq(7, r) {
			g.CanCastle[c][KingSide] = false
		}
		if from == Sq(0, r) || to == Sq(0, r) {
			g.CanCastle[c][QueenSide] = false
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
