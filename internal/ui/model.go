package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/tosterabgx/kingsh/internal/chess"
)

type model struct {
	game      *chess.Game
	clock     chess.Clock
	cursor    chess.Square
	timeout   bool
	selected  chess.Square
	promotion *chess.Move
}

var promotionKeys = []struct {
	key  string
	kind chess.Kind
}{
	{"1", chess.Queen},
	{"2", chess.Rook},
	{"3", chess.Bishop},
	{"4", chess.Knight},
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func New(base, inc time.Duration) model {
	return model{
		game:     chess.NewGame(),
		clock:    chess.NewClock(base, inc),
		cursor:   chess.Sq(4, 1),
		selected: chess.NoSquare,
	}
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) over() bool {
	return m.timeout || m.game.Status() != chess.Playing
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		if !m.timeout && m.clock.Update(time.Time(msg)) {
			m.flagFell()
		}
		return m, tick()
	case tea.KeyPressMsg:
		if m.promotion != nil && msg.String() != "ctrl+c" {
			m.choosePromotion(msg.String())
			return m, nil
		}

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
		case "enter", "space":
			m.activate()
		case "esc":
			m.selected = chess.NoSquare
		}
	}

	return m, nil
}

func (m *model) moveCursor(dx, dy int) {
	f := max(0, min(7, m.cursor.File()+dx))
	r := max(0, min(7, m.cursor.Rank()+dy))
	m.cursor = chess.Sq(f, r)
}

func (m *model) activate() {
	if m.over() {
		return
	}
	from, to := m.selected, m.cursor
	p := m.game.Board[to]
	switch {
	case !p.Empty() && p.Color == m.game.Turn:
		m.selected = to
	case from != chess.NoSquare:
		if m.game.IsPromotion(from, to) && m.game.CanMove(from, to) {
			m.promotion = &chess.Move{From: from, To: to}
			return
		}
		m.play(from, to, chess.Queen)
		m.selected = chess.NoSquare
	}
}

func (m *model) choosePromotion(key string) {
	if key == "esc" {
		m.promotion = nil
		return
	}

	for _, pk := range promotionKeys {
		if key == pk.key {
			m.play(m.promotion.From, m.promotion.To, pk.kind)
			m.promotion = nil
			m.selected = chess.NoSquare
			return
		}
	}
}

func (m *model) play(from, to chess.Square, promo chess.Kind) {
	now := time.Now()
	if m.clock.Update(now) {
		m.flagFell()
		return
	}
	mover := m.game.Turn
	if !m.game.MovePromote(from, to, promo) {
		return
	}
	m.clock.Press(now, mover)
	if m.game.Status() != chess.Playing {
		m.clock.Stop()
	}
}

func (m *model) flagFell() {
	m.timeout = true
	m.selected = chess.NoSquare
	m.promotion = nil
}
