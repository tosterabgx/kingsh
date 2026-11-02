package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/tosterabgx/kingsh/internal/chess"
	"github.com/tosterabgx/kingsh/internal/lichess"
	"github.com/tosterabgx/kingsh/internal/ui"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "online" {
		os.Exit(online(os.Args[2:]))
	}

	tc := flag.String("tc", "10+0", "time control: minutes+inc in seconds, e.g. 10+0")
	flag.Parse()

	base, inc, err := chess.ParseTimeControl(*tc)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}

	p := tea.NewProgram(ui.New(base, inc))
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func online(args []string) int {
	fs := flag.NewFlagSet("kingsh online", flag.ContinueOnError)
	tc := fs.String("tc", "10+0", "time control: minutes+inc in seconds, e.g. 10+0")
	casual := fs.Bool("casual", false, "play an unrated game")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	base, inc, err := chess.ParseTimeControl(*tc)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 2
	}
	if !lichess.IsRapidOrSlower(base, inc) {
		fmt.Fprintf(os.Stderr, "error: lichess only lets apps seek rapid or slower games (minutes + 40 × increment of at least 8 minutes, e.g. 10+0 or 5+5); %s is too fast\n", *tc)
		return 2
	}

	token := os.Getenv("LICHESS_TOKEN")
	if token == "" {
		fmt.Fprintln(os.Stderr, "error: set LICHESS_TOKEN to a lichess API token with the board:play scope.")
		fmt.Fprintln(os.Stderr, "Create one here:", lichess.TokenURL)
		return 2
	}

	seek := lichess.Seek{Time: base, Increment: inc, Rated: !*casual}
	p := tea.NewProgram(ui.NewOnline(lichess.NewClient(token), seek))
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return 0
}
