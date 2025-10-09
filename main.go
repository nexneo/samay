package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nexneo/samay/tui"
	"github.com/nexneo/samay/util/version"
)

func main() {
	showVersion := flag.Bool("version", false, "print the samay version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.String())
		return
	}

	p := tea.NewProgram(tui.CreateApp())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}
}
