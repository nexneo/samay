package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nexneo/samay/tui"
)

func main() {
	p := tea.NewProgram(tui.CreateApp())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}
}
