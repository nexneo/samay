package main

import (
	"flag"
	"fmt"
	"github.com/nexneo/samay/data"
	"log"
	"os"
	"strings"
	"time"
)

var (
	cmd,
	name,
	typ,
	newName,
	content string

	billable bool

	duration time.Duration

	idx,
	month,
	week int

	cmdflags *flag.FlagSet
)

func main() {
	parseFlags()
	project := data.CreateProject(name)
	// fmt.Println("Running...", cmd)
	fn := commands[cmd]
	if fn != nil {
		err := fn(project)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		usage()
	}
}

func parseFlags() {
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	} else {
		cmd = "report"
	}

	initflags(cmd)

	if cmd == "help" {
		usage()
		os.Exit(0)
	}

	if len(os.Args) < 3 {
		cmdflags.Parse(os.Args)
	} else if cmd == "mv" {
		name = os.Args[2]
		newName = os.Args[3]
	} else {
		cmdflags.Parse(os.Args[2:])
		// if -p flag not set, check if second argument is Project name
		if name == "" && !strings.HasPrefix(os.Args[2], "-") {
			name = os.Args[2]
			cmdflags.Parse(os.Args[3:])
		}
	}

	if name == "" && cmd != "report" && cmd != "web" {
		log.Print("Error: Project name required.\n\n")
		usage()
		os.Exit(0)
	}
}

// Bind variables with command line flags
func initflags(cmd string) {
	cmdflags = flag.NewFlagSet(cmd, flag.ExitOnError)
	cmdflags.StringVar(&content, "m", "", "Description for entry")

	cmdflags.IntVar(&month, "r", int(time.Now().Month()), "Report Month")
	cmdflags.IntVar(&idx, "i", -1, "Entry index(#) in log")

	cmdflags.DurationVar(&duration, "d", time.Hour, "Entry Duration")

	cmdflags.BoolVar(&billable, "bill", true, "Billable")
}

func usage() {
	fmt.Fprintf(os.Stderr,
		"Usage: samay [command] [project] [options|new project]\n\nCommands:\n%s%s%s%s%s%s%s%s%s\n\nOptions:\n",
		"  start (s) : Start timer\n",
		"  stop  (p) : Stop Timer and entry with -m=(description) or uses $EDITOR\n",
		"  entry (e) : Direct entry with -d=(duration)\n",
		"  rm        : Remove all entries, or single with -i=(#)\n",
		"  mv        : Move all entries to new project\n",
		"  log       : Show log for project\n",
		"  show      : Show project or entry with -i=(#)\n",
		"  report    : Report for all projects\n",
		"  help      : Help\n")
	cmdflags.Parse([]string{})
	cmdflags.PrintDefaults()
}
