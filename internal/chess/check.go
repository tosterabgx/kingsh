package chess

type Status int8

const (
	Playing Status = iota
	Checkmate
	Stalemate
)

func (g *Game) LegalMoves(from Square) []Square {
	p := g.Board[from]
	if p.Empty() {
		return nil
	}
	var out []Square
	for _, to := range g.possibleMoves(from) {
		next := Game{Board: g.Board}
		next.Board.play(from, to, Queen)
		if !next.InCheck(p.Color) {
			out = append(out, to)
		}
	}
	return append(out, g.castleMoves(from)...)
}

func (g *Game) castleMoves(from Square) []Square {
	p := g.Board[from]
	r := 0
	if p.Color == Black {
		r = 7
	}
	if p.Kind != King || from != Sq(4, r) || g.InCheck(p.Color) {
		return nil
	}
	enemy := p.Color.Other()

	sides := []struct {
		side     int
		rookFile int
		empty    []int
		cross    []int
	}{
		{KingSide, 7, []int{5, 6}, []int{5, 6}},
		{QueenSide, 0, []int{1, 2, 3}, []int{3, 2}},
	}

	var out []Square
	for _, s := range sides {
		if !g.CanCastle[p.Color][s.side] || g.Board[Sq(s.rookFile, r)] != (Piece{Rook, p.Color}) {
			continue
		}
		ok := true
		for _, f := range s.empty {
			if !g.Board[Sq(f, r)].Empty() {
				ok = false
			}
		}
		for _, f := range s.cross {
			if ok && g.Attacked(Sq(f, r), enemy) {
				ok = false
			}
		}
		if ok {
			out = append(out, Sq(s.cross[len(s.cross)-1], r))
		}

	}
	return out
}

func (g *Game) Attacked(sq Square, by Color) bool {
	for s := Square(0); s < 64; s++ {
		p := g.Board[s]
		if p.Empty() || p.Color != by {
			continue
		}
		if p.Kind == Pawn {
			dir := 1
			if by == Black {
				dir = -1
			}
			if sq.Rank() == s.Rank()+dir && abs(sq.File()-s.File()) == 1 {
				return true
			}
			continue
		}
		for _, to := range g.possibleMoves(s) {
			if to == sq {
				return true
			}
		}
	}
	return false
}

func (g *Game) InCheck(c Color) bool {
	king := NoSquare
	for s := Square(0); s < 64; s++ {
		if p := g.Board[s]; p.Kind == King && p.Color == c {
			king = s
			break
		}
	}
	if king == NoSquare {
		return false
	}

	for s := Square(0); s < 64; s++ {
		p := g.Board[s]
		if p.Empty() || p.Color == c {
			continue
		}
		for _, to := range g.possibleMoves(s) {
			if to == king {
				return true
			}
		}
	}
	return false
}

func (g *Game) Status() Status {
	for s := Square(0); s < 64; s++ {
		if p := g.Board[s]; !p.Empty() && p.Color == g.Turn && len(g.LegalMoves(s)) > 0 {
			return Playing
		}
	}
	if g.InCheck(g.Turn) {
		return Checkmate
	}
	return Stalemate
}
