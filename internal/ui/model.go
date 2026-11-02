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
	online    *online
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
	if m.online != nil {
		return tea.Batch(tick(), m.online.start())
	}
	return tick()
}

func (m model) over() bool {
	if m.online != nil {
		return m.online.finished()
	}
	return m.timeout || m.game.Status() != chess.Playing
}

func (m model) canMove() bool {
	if m.online != nil {
		o := m.online
		return o.active() && m.game.Turn == o.me && len(o.played) == len(o.confirmed)
	}
	return !m.over()
}

func (m model) flipped() bool {
	return m.online != nil && m.online.me == chess.Black
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.online != nil {
		if cmd, ok := m.updateOnline(msg); ok {
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tickMsg:
		if m.clock.Update(time.Time(msg)) && m.online == nil && !m.timeout {
			m.flagFell()
		}
		return m, tick()
	case tea.KeyPressMsg:
		key := msg.String()
		if m.online != nil {
			if cmd, ok := m.onlineKey(key); ok {
				return m, cmd
			}
		}
		if m.promotion != nil && key != "ctrl+c" {
			return m, m.choosePromotion(key)
		}

		switch key {
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
			return m, m.activate()
		case "esc":
			m.selected = chess.NoSquare
		}
	}

	return m, nil
}

func (m *model) moveCursor(dx, dy int) {
	if m.flipped() {
		dx, dy = -dx, -dy
	}
	f := max(0, min(7, m.cursor.File()+dx))
	r := max(0, min(7, m.cursor.Rank()+dy))
	m.cursor = chess.Sq(f, r)
}

func (m *model) activate() tea.Cmd {
	if !m.canMove() {
		return nil
	}
	from, to := m.selected, m.cursor
	p := m.game.Board[to]
	switch {
	case !p.Empty() && p.Color == m.game.Turn:
		m.selected = to
	case from != chess.NoSquare:
		if m.game.IsPromotion(from, to) && m.game.CanMove(from, to) {
			m.promotion = &chess.Move{From: from, To: to}
			return nil
		}
		m.selected = chess.NoSquare
		return m.play(from, to, chess.Queen)
	}
	return nil
}

func (m *model) choosePromotion(key string) tea.Cmd {
	if key == "esc" {
		m.promotion = nil
		return nil
	}

	for _, pk := range promotionKeys {
		if key == pk.key {
			from, to := m.promotion.From, m.promotion.To
			m.promotion = nil
			m.selected = chess.NoSquare
			return m.play(from, to, pk.kind)
		}
	}
	return nil
}

func (m *model) play(from, to chess.Square, promo chess.Kind) tea.Cmd {
	if m.online != nil {
		return m.playOnline(from, to, promo)
	}

	now := time.Now()
	if m.clock.Update(now) {
		m.flagFell()
		return nil
	}
	mover := m.game.Turn
	if !m.game.MovePromote(from, to, promo) {
		return nil
	}
	m.clock.Press(now, mover)
	if m.game.Status() != chess.Playing {
		m.clock.Stop()
	}
	return nil
}

func (m *model) flagFell() {
	m.timeout = true
	m.selected = chess.NoSquare
	m.promotion = nil
}
