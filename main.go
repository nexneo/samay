package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nexneo/samay/data"
	"github.com/nexneo/samay/tui"
	"github.com/nexneo/samay/util/version"
)

func main() {
	showVersion := flag.Bool("version", false, "print the samay version and exit")
	dbOverride := flag.String("database", "", "override the database location for this run")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.String())
		return
	}

	dbPath, err := data.ResolveDatabasePathWithOverride(*dbOverride)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve database path: %v\n", err)
		os.Exit(1)
	}
	if err := data.OpenDatabase(dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "open database: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if data.DB == nil {
			return
		}
		if err := data.DB.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "close database: %v\n", err)
		}
	}()

	p := tea.NewProgram(tui.CreateApp())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error: %v\n", err)
		os.Exit(1)
	}

	if data.DB != nil && data.DB.Path() != "" {
		fmt.Printf("Database located at: %s\n", data.DB.Path())
	}
}
