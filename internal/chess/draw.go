package chess

type position struct {
	board     Board
	turn      Color
	canCastle [2][2]bool
	enPassant Square
}

func (g *Game) position() position {
	ep := NoSquare
	if g.enPassantPossible() {
		ep = g.EnPassant
	}
	return position{g.Board, g.Turn, g.CanCastle, ep}
}

func (g *Game) enPassantPossible() bool {
	if g.EnPassant == NoSquare {
		return false
	}
	r := 4
	if g.Turn == Black {
		r = 3
	}
	for _, df := range []int{-1, 1} {
		f := g.EnPassant.File() + df
		if !onBoard(f, r) {
			continue
		}
		from := Sq(f, r)
		if g.Board[from] == (Piece{Pawn, g.Turn}) && g.canEnPassant(from, g.EnPassant, g.Turn) {
			return true
		}
	}
	return false
}

func (g *Game) Repetitions() int {
	cur := g.position()
	n := 0
	for _, p := range g.history {
		if p == cur {
			n++
		}
	}
	return n
}

func (g *Game) InsufficientMaterial() bool {
	knights := 0
	bishopOnLight, bishopOnDark := false, false
	for s := Square(0); s < 64; s++ {
		switch g.Board[s].Kind {
		case Pawn, Rook, Queen:
			return false
		case Knight:
			knights++
		case Bishop:
			if s.Light() {
				bishopOnLight = true
			} else {
				bishopOnDark = true
			}
		}
	}
	bishops := bishopOnLight || bishopOnDark
	switch {
	case knights == 0:
		return !(bishopOnLight && bishopOnDark)
	case knights == 1:
		return !bishops
	}
	return false
}

func (g *Game) HasMatingMaterial(c Color) bool {
	minors := 0
	for s := Square(0); s < 64; s++ {
		p := g.Board[s]
		if p.Color != c {
			continue
		}
		switch p.Kind {
		case Pawn, Rook, Queen:
			return true
		case Knight, Bishop:
			minors++
		}
	}
	return minors >= 2
}
