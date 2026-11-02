package ui

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tosterabgx/kingsh/internal/chess"
)

const (
	squareW = 8
	squareH = 3
)

var (
	colLight = lipgloss.Color("#e4e6cb")
	colDark  = lipgloss.Color("#719352")

	colLightHighlighted = lipgloss.Color("#f6f6ed")
	colDarkHighlighted  = lipgloss.Color("#cdd9c2")

	colLightSelected = lipgloss.Color("#f1f27f")
	colDarkSelected  = lipgloss.Color("#b8c942")

	colLightHint = lipgloss.Color("#c4c6af")
	colDarkHint  = lipgloss.Color("#617e46")

	colPiece = lipgloss.Color("#000000")
)

var (
	solidGlyphs  = [...]string{" ", "♟", "♞", "♝", "♜", "♛", "♚"}
	hollowGlyphs = [...]string{" ", "♙", "♘", "♗", "♖", "♕", "♔"}
)

func (m model) View() tea.View {
	content := lipgloss.JoinHorizontal(lipgloss.Top, m.renderBoard(), "    ", m.renderPanel())
	v := tea.NewView(content)
	v.AltScreen = true
	v.WindowTitle = "kingsh"
	return v
}

func (m model) renderBoard() string {
	rows := make([]string, 0, 8*squareH)
	for rank := 7; rank >= 0; rank-- {
		lines := make([]string, squareH)
		for file := range 8 {
			for i, s := range m.renderSquare(chess.Sq(file, rank)) {
				lines[i] += s
			}
		}
		rows = append(rows, lines...)
	}
	return strings.Join(rows, "\n")
}

func (m model) renderSquare(sq chess.Square) []string {
	var rows []string

	var bg color.Color

	switch {
	case sq == m.selected:
		bg = colDarkSelected
		if sq.Light() {
			bg = colLightSelected
		}
	case sq == m.cursor:
		bg = colDarkHighlighted
		if sq.Light() {
			bg = colLightHighlighted
		}
	default:
		bg = colDark
		if sq.Light() {
			bg = colLight
		}
	}

	base := lipgloss.NewStyle().Background(bg).Foreground(colPiece)

	hintCol := colDarkHint
	if sq.Light() {
		hintCol = colLightHint
	}
	hint := lipgloss.NewStyle().Background(bg).Foreground(hintCol)

	p := m.game.Board[sq]
	char := solidGlyphs[p.Kind]
	if p.Color == chess.White {
		char = hollowGlyphs[p.Kind]
	}

	target := m.selected != chess.NoSquare && m.game.CanMove(m.selected, sq)

	pad := strings.Repeat(" ", (squareW-2)/2)
	middle := base.Render(pad + char + " " + pad)
	if target && p.Empty() {
		middle = hint.Render(pad + "●" + " " + pad)
	}

	for i := 0; i < squareH; i++ {
		if i == squareH/2 {
			rows = append(rows, middle)
		} else {
			rows = append(rows, base.Render(strings.Repeat(" ", squareW)))
		}
	}
	return rows
}

func (m model) statusLine() string {
	g := m.game
	if m.promotion != nil {
		line := "Promote to:"
		for _, pk := range promotionKeys {
			line += " " + pk.key + " " + solidGlyphs[pk.kind]
		}
		return line + "  esc cancel"
	}

	if m.timeout {
		winner := g.Turn.Other()
		if !g.HasMatingMaterial(winner) {
			return "Draw: timeout vs insufficient material"
		}
		return winner.String() + " wins on time"
	}

	switch g.Status() {
	case chess.Checkmate:
		return "Checkmate, " + g.Turn.Other().String() + " wins"
	case chess.Stalemate:
		return "Stalemate"
	case chess.InsufficientMaterial:
		return "Draw by insufficient material"
	case chess.FiftyMoves:
		return "Draw by 50-move rule"
	case chess.Repetition:
		return "Draw by threefold repetition"
	}
	if g.InCheck(g.Turn) {
		return g.Turn.String() + " to move, check"
	}
	return g.Turn.String() + " to move"
}
