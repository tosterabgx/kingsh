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
		next.Board[to], next.Board[from] = p, Piece{}
		if !next.InCheck(p.Color) {
			out = append(out, to)
		}
	}
	return out
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
