package chess

func (g *Game) LegalMoves(from Square) []Square {
	p := g.Board[from]
	switch p.Kind {
	case NoKind:
		return nil
	}

	return nil
}
