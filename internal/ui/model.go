package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/tosterabgx/kingsh/internal/chess"
)

type model struct {
	game     *chess.Game
	cursor   chess.Square
	selected chess.Square
}

func New() model {
	return model{game: chess.NewGame(), cursor: chess.Sq(4, 1)}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up":
			m.moveCursor(0, 1)
		case "down":
			m.moveCursor(0, -1)
		case "left":
			m.moveCursor(-1, 0)
		case "right":
			m.moveCursor(1, 0)
		case "enter":
			m.selected = m.cursor
		}
	}

	return m, nil
}

func (m *model) moveCursor(dx, dy int) {
	f := max(0, min(7, m.cursor.File()+dx))
	r := max(0, min(7, m.cursor.Rank()+dy))
	m.cursor = chess.Sq(f, r)
}
