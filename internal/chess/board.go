package chess

type Color int8

const (
	White Color = iota
	Black
)

func (c Color) String() string {
	if c == White {
		return "White"
	}
	return "Black"
}

type Kind int8

const (
	NoKind Kind = iota
	Pawn
	Knight
	Bishop
	Rook
	Queen
	King
)

var values = [...]int{NoKind: 0, Pawn: 1, Knight: 3, Bishop: 3, Rook: 5, Queen: 9, King: 0}

type Piece struct {
	Kind  Kind
	Color Color
}

func (p Piece) Empty() bool { return p.Kind == NoKind }

func (p Piece) Value() int { return values[p.Kind] }

type Square int8

func Sq(file, rank int) Square { return Square(rank*8 + file) }
func (s Square) File() int     { return int(s) % 8 }
func (s Square) Rank() int     { return int(s) / 8 }
func (s Square) Light() bool   { return (s.File()+s.Rank())%2 == 1 }

type Board [64]Piece

func StartingBoard() Board {
	var b Board
	back := [...]Kind{Rook, Knight, Bishop, Queen, King, Bishop, Knight, Rook}
	for f, k := range back {
		b[Sq(f, 0)] = Piece{k, White}
		b[Sq(f, 1)] = Piece{Pawn, White}
		b[Sq(f, 6)] = Piece{Pawn, Black}
		b[Sq(f, 7)] = Piece{k, Black}
	}
	return b
}
