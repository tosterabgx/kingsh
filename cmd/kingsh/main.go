package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/tosterabgx/kingsh/internal/ui"
)

func main() {

	p := tea.NewProgram(ui.New())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
