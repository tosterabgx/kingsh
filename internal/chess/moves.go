package chess

var (
	knightSteps = [][2]int{{1, 2}, {2, 1}, {2, -1}, {1, -2}, {-1, -2}, {-2, -1}, {-2, 1}, {-1, 2}}
	kingSteps   = [][2]int{{0, 1}, {1, 1}, {1, 0}, {1, -1}, {0, -1}, {-1, -1}, {-1, 0}, {-1, 1}}
	rookDirs    = [][2]int{{0, 1}, {1, 0}, {0, -1}, {-1, 0}}
	bishopDirs  = [][2]int{{1, 1}, {1, -1}, {-1, -1}, {-1, 1}}
	queenDirs   = append(append([][2]int{}, rookDirs...), bishopDirs...)
)

func (g *Game) possibleMoves(from Square) []Square {
	p := g.Board[from]
	switch p.Kind {
	case Pawn:
		return g.pawnMoves(from, p.Color)
	case Knight:
		return g.stepMoves(from, p.Color, knightSteps)
	case King:
		return g.stepMoves(from, p.Color, kingSteps)
	case Bishop:
		return g.slideMoves(from, p.Color, bishopDirs)
	case Rook:
		return g.slideMoves(from, p.Color, rookDirs)
	case Queen:
		return g.slideMoves(from, p.Color, queenDirs)
	}

	return nil
}

func (g *Game) CanMove(from, to Square) bool {
	for _, s := range g.LegalMoves(from) {
		if s == to {
			return true
		}
	}
	return false
}

func onBoard(file, rank int) bool {
	return file >= 0 && file < 8 && rank >= 0 && rank < 8
}

func (g *Game) pawnMoves(from Square, c Color) []Square {
	dir, startRank := 1, 1
	if c == Black {
		dir, startRank = -1, 6
	}
	f, r := from.File(), from.Rank()
	var out []Square

	if onBoard(f, r+dir) && g.Board[Sq(f, r+dir)].Empty() {
		out = append(out, Sq(f, r+dir))
		if r == startRank && g.Board[Sq(f, r+2*dir)].Empty() {
			out = append(out, Sq(f, r+2*dir))
		}
	}

	for _, df := range []int{-1, 1} {
		if !onBoard(f+df, r+dir) {
			continue
		}
		to := Sq(f+df, r+dir)
		if t := g.Board[to]; !t.Empty() && t.Color != c {
			out = append(out, to)
		}
	}
	return out
}

func (g *Game) stepMoves(from Square, c Color, steps [][2]int) []Square {
	var out []Square
	for _, s := range steps {
		f, r := from.File()+s[0], from.Rank()+s[1]
		if !onBoard(f, r) {
			continue
		}
		to := Sq(f, r)
		if t := g.Board[to]; t.Empty() || t.Color != c {
			out = append(out, to)
		}
	}
	return out
}

func (g *Game) slideMoves(from Square, c Color, dirs [][2]int) []Square {
	var out []Square
	for _, d := range dirs {
		f, r := from.File()+d[0], from.Rank()+d[1]
		for onBoard(f, r) {
			to := Sq(f, r)
			t := g.Board[to]
			if !t.Empty() {
				if t.Color != c {
					out = append(out, to)
				}
				break
			}
			out = append(out, to)
			f, r = f+d[0], r+d[1]
		}
	}
	return out
}
