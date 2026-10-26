package ui

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/tosterabgx/kingsh/internal/chess"
)

const (
	squareW = 7
	squareH = 3
)

var (
	colLight         = lipgloss.Color("#e4e6cb")
	colDark          = lipgloss.Color("#719352")
	colLightSelected = lipgloss.Color("#eeef7f")
	colDarkSelected  = lipgloss.Color("#b3c542")
)

var (
	solidGlyphs  = [...]string{" ", "♟", "♞", "♝", "♜", "♛", "♚"}
	hollowGlyphs = [...]string{" ", "♙", "♘", "♗", "♖", "♕", "♔"}
)

func (m model) View() tea.View {
	v := tea.NewView(m.renderBoard())
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

	if sq.Light() {
		bg = colLight
		if sq == m.cursor {
			bg = colLightSelected
		}
	} else {
		bg = colDark
		if sq == m.cursor {
			bg = colDarkSelected
		}
	}

	base := lipgloss.NewStyle().Background(bg)

	for i := 0; i < squareH; i++ {
		rows = append(rows, base.Render(strings.Repeat(" ", squareW)))
	}
	return rows
}
