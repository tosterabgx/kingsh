package ui

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/tosterabgx/kingsh/internal/chess"
	"github.com/tosterabgx/kingsh/internal/lichess"
)

type online struct {
	client *lichess.Client
	seek   lichess.Seek

	ctx      context.Context
	stop     context.CancelFunc
	stopSeek context.CancelFunc
	msgs     chan tea.Msg

	gameID       string
	me           chess.Color
	white, black lichess.Player
	confirmed    []string
	played       []string
	status       string
	winner       string
	drawOffered  bool
	drawSent     bool
	opponentGone bool
	confirm      string
	err          string
}

type (
	gameStartMsg struct {
		id    string
		color chess.Color
	}
	gameEventMsg     lichess.GameEvent
	gameStreamEndMsg struct{ err error }
	eventStreamEnd   struct{ err error }
	seekEndedMsg     struct{ err error }
	reseekMsg        struct{}
	reconnectMsg     struct{}
	moveFailedMsg    struct{ err error }
	actionFailedMsg  struct{ err error }
)

func NewOnline(client *lichess.Client, seek lichess.Seek) model {
	m := New(seek.Time, seek.Increment)
	ctx, stop := context.WithCancel(context.Background())
	m.online = &online{client: client, seek: seek, ctx: ctx, stop: stop, msgs: make(chan tea.Msg, 64)}
	return m
}

func (o *online) start() tea.Cmd {
	go o.streamEvents()
	o.startSeek()
	return o.next()
}

func (o *online) next() tea.Cmd {
	return func() tea.Msg {
		select {
		case msg := <-o.msgs:
			return msg
		case <-o.ctx.Done():
			return nil
		}
	}
}

func (o *online) send(msg tea.Msg) {
	select {
	case o.msgs <- msg:
	case <-o.ctx.Done():
	}
}

func (o *online) streamEvents() {
	err := o.client.StreamEvents(o.ctx, func(e lichess.Event) {
		if e.Type == "gameStart" {
			color := chess.White
			if e.Game.Color == "black" {
				color = chess.Black
			}
			o.send(gameStartMsg{e.Game.ID, color})
		}
	})
	o.send(eventStreamEnd{err})
}

func (o *online) startSeek() {
	ctx, cancel := context.WithCancel(o.ctx)
	o.stopSeek = cancel
	go func() {
		err := o.client.SeekGame(ctx, o.seek)
		if ctx.Err() == nil {
			o.send(seekEndedMsg{err})
		}
	}()
}

func (o *online) streamGame() {
	err := o.client.StreamGame(o.ctx, o.gameID, func(e lichess.GameEvent) {
		o.send(gameEventMsg(e))
	})
	o.send(gameStreamEndMsg{err})
}

func (o *online) seeking() bool { return o.gameID == "" }

func (o *online) active() bool { return o.gameID != "" && o.status == "started" }

func (o *online) finished() bool {
	return o.gameID != "" && o.status != "" && o.status != "created" && o.status != "started"
}

func (o *online) player(c chess.Color) lichess.Player {
	if c == chess.White {
		return o.white
	}
	return o.black
}

func (m *model) updateOnline(msg tea.Msg) (tea.Cmd, bool) {
	o := m.online
	switch msg := msg.(type) {
	case gameStartMsg:
		if o.gameID == "" {
			o.gameID, o.me = msg.id, msg.color
			o.stopSeek()
			m.cursor = chess.Sq(4, 1)
			if o.me == chess.Black {
				m.cursor = chess.Sq(4, 6)
			}
			go o.streamGame()
		}
		return o.next(), true

	case gameEventMsg:
		e := lichess.GameEvent(msg)
		switch e.Type {
		case "gameFull":
			o.white, o.black = e.White, e.Black
			if e.Clock != nil {
				m.clock.Increment = time.Duration(e.Clock.Increment) * time.Millisecond
			}
			m.applyState(e.State)
		case "gameState":
			m.applyState(e.GameState)
		case "opponentGone":
			o.opponentGone = e.Gone
		}
		return o.next(), true

	case gameStreamEndMsg:
		if !o.finished() && o.ctx.Err() == nil {
			o.err = "connection to the game lost, reconnecting…"
			return tea.Batch(o.next(), tea.Tick(2*time.Second, func(time.Time) tea.Msg { return reconnectMsg{} })), true
		}
		return o.next(), true

	case reconnectMsg:
		if !o.finished() && o.ctx.Err() == nil {
			go o.streamGame()
		}
		return nil, true

	case eventStreamEnd:
		if o.seeking() && o.ctx.Err() == nil {
			o.err = errText("lost the connection to lichess", msg.err)
		}
		return o.next(), true

	case seekEndedMsg:
		switch {
		case msg.err != nil && o.seeking():
			o.err = errText("could not seek a game", msg.err)
		case o.seeking():
			return tea.Batch(o.next(), tea.Tick(3*time.Second, func(time.Time) tea.Msg { return reseekMsg{} })), true
		}
		return o.next(), true

	case reseekMsg:
		if o.seeking() && o.ctx.Err() == nil {
			o.startSeek()
		}
		return nil, true

	case moveFailedMsg:
		o.err = errText("move refused", msg.err)
		m.rebuild(o.confirmed)
		return nil, true

	case actionFailedMsg:
		o.err = errText("lichess", msg.err)
		return nil, true
	}
	return nil, false
}

func (m *model) onlineKey(key string) (tea.Cmd, bool) {
	o := m.online
	if key == "ctrl+c" {
		o.stop()
		return tea.Quit, true
	}

	if o.confirm != "" {
		what := o.confirm
		o.confirm = ""
		switch {
		case what == "quit" && key == "q":
			o.stop()
			return tea.Quit, true
		case what == "resign" && key == "R":
			return o.resign(), true
		}
		return nil, true
	}

	switch key {
	case "q":
		if o.active() {
			o.confirm = "quit"
			return nil, true
		}
		o.stop()
		return tea.Quit, true
	case "R":
		if o.active() {
			o.confirm = "resign"
		}
		return nil, true
	case "D":
		if o.active() {
			o.drawSent = true
			return o.action(o.client.OfferDraw), true
		}
		return nil, true
	}
	return nil, false
}

func (o *online) resign() tea.Cmd {
	if len(o.confirmed) < 2 {
		return o.action(o.client.Abort)
	}
	return o.action(o.client.Resign)
}

func (o *online) action(do func(context.Context, string) error) tea.Cmd {
	id := o.gameID
	return func() tea.Msg {
		if err := do(o.ctx, id); err != nil {
			return actionFailedMsg{err}
		}
		return nil
	}
}

func (m *model) playOnline(from, to chess.Square, promo chess.Kind) tea.Cmd {
	o := m.online
	if !m.game.IsPromotion(from, to) {
		promo = chess.NoKind
	}
	uci := chess.UCI(from, to, promo)
	mover := m.game.Turn
	if promo == chess.NoKind {
		promo = chess.Queen
	}
	if !m.game.MovePromote(from, to, promo) {
		return nil
	}
	o.played = append(slices.Clone(o.played), uci)
	o.err = ""
	m.clock.Press(time.Now(), mover)
	m.clock.Running = len(o.played) >= 2

	id := o.gameID
	return func() tea.Msg {
		if err := o.client.Move(o.ctx, id, uci); err != nil {
			return moveFailedMsg{err}
		}
		return nil
	}
}

func (m *model) applyState(s lichess.GameState) {
	o := m.online
	moves := strings.Fields(s.Moves)
	pending := len(o.played) == len(moves)+1 && slices.Equal(o.played[:len(moves)], moves)

	switch {
	case slices.Equal(moves, o.played), pending:
	case len(moves) == len(o.played)+1 && slices.Equal(moves[:len(o.played)], o.played):
		if err := m.game.MoveUCI(moves[len(moves)-1]); err != nil {
			o.err = "can't follow the game: " + err.Error()
			m.rebuild(moves)
		} else {
			o.played = moves
		}
	default:
		m.rebuild(moves)
	}
	o.confirmed = moves
	if !pending && o.err != "" && strings.HasPrefix(o.err, "connection") {
		o.err = ""
	}

	o.status, o.winner = s.Status, s.Winner
	if o.me == chess.White {
		o.drawOffered, o.drawSent = s.BDraw, s.WDraw
	} else {
		o.drawOffered, o.drawSent = s.WDraw, s.BDraw
	}

	running := s.Status == "started" && len(o.played) >= 2
	m.clock.Set(time.Duration(s.WTime)*time.Millisecond, time.Duration(s.BTime)*time.Millisecond,
		m.game.Turn, running, time.Now())

	if m.selected != chess.NoSquare {
		if p := m.game.Board[m.selected]; p.Empty() || p.Color != o.me {
			m.selected = chess.NoSquare
		}
	}
	if o.finished() {
		m.selected = chess.NoSquare
		m.promotion = nil
		o.confirm = ""
	}
}

func (m *model) rebuild(moves []string) {
	o := m.online
	g, err := chess.GameFromUCI(moves)
	if err != nil {
		o.err = "can't follow the game: " + err.Error()
		return
	}
	m.game = g
	o.played = moves
	m.selected = chess.NoSquare
	m.promotion = nil
}

func (m model) onlineStatus() string {
	o := m.online
	switch {
	case o.confirm == "quit":
		return "Press q again to leave (your clock keeps running on lichess)"
	case o.confirm == "resign":
		if len(o.confirmed) < 2 {
			return "Press R again to abort the game"
		}
		return "Press R again to resign"
	case o.seeking():
		if o.err != "" {
			return o.err
		}
		rated := "casual"
		if o.seek.Rated {
			rated = "rated"
		}
		return fmt.Sprintf("Seeking a %s %s game…", rated, timeControl(o.seek))
	case o.finished():
		return m.onlineResult()
	case o.err != "":
		return o.err
	case o.opponentGone:
		return "Your opponent left the game"
	case o.drawOffered:
		return "Draw offered: D to accept"
	}

	var s string
	if m.game.Turn == o.me {
		s = "Your move"
	} else {
		s = "Waiting for your opponent"
	}
	if m.game.InCheck(m.game.Turn) {
		s += ", check"
	}
	if o.drawSent {
		s += " · draw offered"
	}
	return s
}

func (m model) onlineResult() string {
	o := m.online
	winner := chess.White
	if o.winner == "black" {
		winner = chess.Black
	}
	name := func(c chess.Color) string { return o.player(c).Name }

	switch o.status {
	case "mate":
		return "Checkmate, " + name(winner) + " wins"
	case "resign":
		return name(winner.Other()) + " resigned, " + name(winner) + " wins"
	case "outoftime":
		if o.winner == "" {
			return "Draw: timeout vs insufficient material"
		}
		return name(winner) + " wins on time"
	case "timeout":
		return name(winner.Other()) + " left the game, " + name(winner) + " wins"
	case "stalemate":
		return "Stalemate"
	case "draw":
		switch m.game.Status() {
		case chess.InsufficientMaterial:
			return "Draw by insufficient material"
		case chess.FiftyMoves:
			return "Draw by 50-move rule"
		case chess.Repetition:
			return "Draw by threefold repetition"
		}
		return "Draw"
	case "aborted", "noStart":
		return "Game aborted"
	}
	if o.winner != "" {
		return "Game over, " + name(winner) + " wins"
	}
	return "Game over"
}

func (m model) onlineHelp() string {
	o := m.online
	switch {
	case o.seeking():
		return "q cancel"
	case o.finished():
		return "q quit"
	}
	return "R resign · D draw · q leave"
}

func timeControl(s lichess.Seek) string {
	return fmt.Sprintf("%s+%d", strings.TrimSuffix(fmt.Sprintf("%.1f", s.Time.Minutes()), ".0"), int(s.Increment/time.Second))
}

func errText(what string, err error) string {
	if err == nil {
		return what
	}
	return what + ": " + err.Error()
}
