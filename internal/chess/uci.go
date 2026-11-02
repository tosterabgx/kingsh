package chess

import (
	"fmt"
	"strings"
)

func (s Square) String() string {
	if s < 0 || s > 63 {
		return "-"
	}
	return string(rune('a'+s.File())) + string(rune('1'+s.Rank()))
}

func ParseSquare(name string) (Square, error) {
	if len(name) != 2 || name[0] < 'a' || name[0] > 'h' || name[1] < '1' || name[1] > '8' {
		return NoSquare, fmt.Errorf("bad square %q", name)
	}
	return Sq(int(name[0]-'a'), int(name[1]-'1')), nil
}

var promoLetters = map[Kind]string{Knight: "n", Bishop: "b", Rook: "r", Queen: "q"}

func UCI(from, to Square, promo Kind) string {
	return from.String() + to.String() + promoLetters[promo]
}

func (g *Game) MoveUCI(move string) error {
	if len(move) != 4 && len(move) != 5 {
		return fmt.Errorf("bad move %q", move)
	}
	from, err := ParseSquare(move[:2])
	if err != nil {
		return fmt.Errorf("bad move %q: %w", move, err)
	}
	to, err := ParseSquare(move[2:4])
	if err != nil {
		return fmt.Errorf("bad move %q: %w", move, err)
	}

	promo := Queen
	if len(move) == 5 {
		promo = NoKind
		for k, letter := range promoLetters {
			if move[4:] == letter {
				promo = k
			}
		}
		if promo == NoKind {
			return fmt.Errorf("bad promotion in %q", move)
		}
	}

	if !g.MovePromote(from, to, promo) {
		return fmt.Errorf("illegal move %q", move)
	}
	return nil
}

func GameFromUCI(moves []string) (*Game, error) {
	g := NewGame()
	for _, mv := range moves {
		if err := g.MoveUCI(strings.TrimSpace(mv)); err != nil {
			return nil, err
		}
	}
	return g, nil
}
