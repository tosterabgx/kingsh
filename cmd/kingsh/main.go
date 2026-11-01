package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/tosterabgx/kingsh/internal/chess"
	"github.com/tosterabgx/kingsh/internal/ui"
)

func main() {
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
