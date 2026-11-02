package ui

import (
	"fmt"
	"image/color"
	"slices"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/tosterabgx/kingsh/internal/chess"
)

const clockW = 14

var (
	colMuted = lipgloss.Color("#989795")

	colWhiteClockBg     = lipgloss.Color("#ffffff")
	colWhiteClockFg     = lipgloss.Color("#262421")
	colWhiteClockIdleBg = lipgloss.Color("#989795")
	colWhiteClockIdleFg = lipgloss.Color("#4b4847")
	colBlackClockBg     = lipgloss.Color("#262421")
	colBlackClockFg     = lipgloss.Color("#ffffff")
	colBlackClockIdleBg = lipgloss.Color("#3c3a38")
	colBlackClockIdleFg = lipgloss.Color("#8b8987")
	colLowClockBg       = lipgloss.Color("#c33a2c")
	colLowClockFg       = lipgloss.Color("#ffffff")
)

func (m model) renderPanel() string {
	h := 8 * squareH

	bottomColor := chess.White
	if m.flipped() {
		bottomColor = chess.Black
	}
	topColor := bottomColor.Other()

	top := append(m.clockBox(topColor), "", m.nameLine(topColor), m.capturedLine(topColor))
	bottom := []string{m.capturedLine(bottomColor), m.nameLine(bottomColor), ""}
	bottom = append(bottom, m.clockBox(bottomColor)...)

	middle := make([]string, max(0, h-len(top)-len(bottom)))
	if len(middle) > 0 {
		mid := (len(middle) - 1) / 2
		middle[mid] = m.statusLine()
		if help := m.helpLine(); help != "" && mid+2 < len(middle) {
			middle[mid+2] = lipgloss.NewStyle().Foreground(colMuted).Render(help)
		}
	}

	lines := slices.Concat(top, middle, bottom)
	return strings.Join(lines[:h], "\n")
}

func (m model) nameLine(c chess.Color) string {
	name := lipgloss.NewStyle().Bold(true).Render(c.String())
	if m.online == nil || m.online.seeking() {
		return name
	}

	p := m.online.player(c)
	title := ""
	if p.Title != "" {
		title = p.Title + " "
	}
	name = lipgloss.NewStyle().Bold(true).Render(title + p.Name)
	if p.Rating > 0 {
		name += lipgloss.NewStyle().Foreground(colMuted).Render(" (" + strconv.Itoa(p.Rating) + ")")
	}
	return name
}

func (m model) helpLine() string {
	if m.online == nil {
		return ""
	}
	return m.onlineHelp()
}

func (m model) capturedLine(c chess.Color) string {
	taken := slices.Clone(m.game.Captured[c])
	slices.SortFunc(taken, func(a, b chess.Piece) int { return int(a.Kind) - int(b.Kind) })

	b := ""
	for _, p := range taken {
		if p.Color == chess.White {
			b += hollowGlyphs[p.Kind]
		} else {
			b += solidGlyphs[p.Kind]
		}
	}
	if lead := material(m.game.Captured[c]) - material(m.game.Captured[c.Other()]); lead > 0 {
		b += " +" + strconv.Itoa(lead)
	}
	return lipgloss.NewStyle().Foreground(colMuted).Render(b)
}

func material(pieces []chess.Piece) int {
	n := 0
	for _, p := range pieces {
		n += p.Value()
	}
	return n
}

func (m model) clockBox(c chess.Color) []string {
	left := m.clock.Left[c]
	active := m.clock.Running && m.clock.Active == c

	var bg, fg color.Color
	switch {
	case c == chess.White && active:
		bg, fg = colWhiteClockBg, colWhiteClockFg
	case c == chess.White:
		bg, fg = colWhiteClockIdleBg, colWhiteClockIdleFg
	case active:
		bg, fg = colBlackClockBg, colBlackClockFg
	default:
		bg, fg = colBlackClockIdleBg, colBlackClockIdleFg
	}
	if left == 0 || (active && left < 20*time.Second) {
		bg, fg = colLowClockBg, colLowClockFg
	}

	st := lipgloss.NewStyle().Background(bg).Foreground(fg).Bold(active).Width(clockW)
	t := formatClock(left)
	return []string{
		st.Render(""),
		st.Align(lipgloss.Right).Render(t + "  "),
		st.Render(""),
	}
}

func formatClock(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d < 20*time.Second {
		tenths := int(d / (100 * time.Millisecond))
		return fmt.Sprintf("0:%02d.%d", tenths/10, tenths%10)
	}

	s := int((d + time.Second - 1) / time.Second)
	if s >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", s/3600, s/60%60, s%60)
	}
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}
