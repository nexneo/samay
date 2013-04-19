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
	content string

	help,
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

	if len(os.Args) < 3 {
		cmdflags.Parse(os.Args)
	} else {
		cmdflags.Parse(os.Args[2:])
		// if -p flag not set, check if second argument is Project name
		if name == "" && !strings.HasPrefix(os.Args[2], "-") {
			name = os.Args[2]
			cmdflags.Parse(os.Args[3:])
		}
	}

	if help || cmd == "help" {
		usage()
		os.Exit(0)
	}

	if name == "" && cmd != "report" {
		log.Println("Error: Project name required.\n")
		usage()
		os.Exit(0)
	}
}

// Bind variables with command line flags
func initflags(cmd string) {
	cmdflags = flag.NewFlagSet(cmd, flag.ExitOnError)
	cmdflags.StringVar(&name, "p", "", "Project Name")
	cmdflags.StringVar(&content, "m", "", "Log content")

	cmdflags.IntVar(&month, "r", int(time.Now().Month()), "Report Month")
	cmdflags.IntVar(&idx, "i", -1, "Index(#)")

	cmdflags.DurationVar(&duration, "d", time.Hour, "Entry Duration")

	cmdflags.BoolVar(&billable, "bill", true, "Billable")
	cmdflags.BoolVar(&help, "h", false, "This Help")
}

func usage() {
	fmt.Fprintf(os.Stderr,
		"Usage: samay [command] [project] [options]\n\nCommands:\n%s%s%s%s%s%s%s\n\nOptions:\n",
		"  start (s) : Start timer\n",
		"  stop  (p) : Stop Timer\n",
		"  entry (e) : Direct time log entry\n",
		"  remove    : Remove project along with all entries\n",
		"  log       : Show log for project\n",
		"  del       : Delete specific entry (specifiy -#)\n",
		"  report    : Report for all projects\n")
	cmdflags.Parse([]string{})
	cmdflags.PrintDefaults()
}
