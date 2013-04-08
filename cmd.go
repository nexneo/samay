package main

import (
	"flag"
	"fmt"
	"github.com/nexneo/samay/data"
	"log"
	"os"
	"time"
)

var (
	cmd,
	name,
	typ,
	message string

	help,
	billable bool

	duration time.Duration

	month,
	week int

	cmdflags *flag.FlagSet
)

func main() {
	parseFlags()
	project := data.CreateProject(name)

	var timer *data.Timer
	var err error

	switch cmd {

	case "start", "s":
		data.Save(project)
		timer = data.CreateTimer(project)
		err = data.Save(timer)

	case "stop", "p":
		if yes, timer := project.OnClock(); yes {
			entry := project.CreateEntry(message, billable)

			if err = timer.Stop(entry); err == nil {
				fmt.Printf("%.2f mins\n", entry.Minutes())
			}
		}

	case "entry", "e":
		entry := project.CreateEntryWithDuration(message, duration, billable)
		err = data.Save(entry)

	case "report", "t":
		if month > 0 && month < 13 {
			data.PrintStatus(month)
		} else {
			err = fmt.Errorf("Month %d is not valid", month)
		}

	case "remove", "rm":
		var remove string
		fmt.Printf(
			"Remove all data for project \"%s\" ([No]/yes)? ",
			project.GetName(),
		)

		if fmt.Scanln(&remove); remove == "yes" {
			err = data.Destroy(project)
		}

	case "log":
		for i, entry := range project.Entries() {
			if i > 10 {
				break
			}
			fmt.Println(entry.HoursMins(), entry.GetContent())
		}
	}

	if err != nil {
		log.Fatalln(err)
	}
}

func parseFlags() {
	if len(os.Args) < 2 || os.Args[1] == "-help" || os.Args[1] == "-h" {
		initflags("command")
		usage()
		os.Exit(0)
	}
	cmd = os.Args[1]
	initflags(cmd)
	if len(os.Args) < 2 {
		cmdflags.Parse([]string{})
	} else {
		cmdflags.Parse(os.Args[2:])
	}

	if help {
		usage()
	}
}

func initflags(cmd string) {
	cmdflags = flag.NewFlagSet(cmd, flag.ExitOnError)
	cmdflags.StringVar(&name, "p", "Default", "Project Name")
	cmdflags.StringVar(&message, "m", "", "Log message")
	cmdflags.IntVar(&month, "r", int(time.Now().Month()), "Report Month")
	cmdflags.BoolVar(&billable, "bill", true, "Billable")
	cmdflags.DurationVar(&duration, "d", time.Hour, "Entry Duration")
	cmdflags.BoolVar(&help, "h", false, "This Help")
}

func usage() {
	fmt.Fprintf(os.Stderr,
		"Usage: samay [command] [options]\n\nCommands:\n%v\n\nOptions:\n",
		[]string{"start", "stop", "entry", "report", "remove"})
	cmdflags.Parse([]string{})
	cmdflags.PrintDefaults()
}
