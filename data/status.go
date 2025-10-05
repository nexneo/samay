package data

import (
	"fmt"
	"strings"
	"time"

	"github.com/nexneo/samay/util"
)

var (
	scoppedMonth int
	scoppedYear  int
)

type hoursMins struct {
	hours, mins int
}

type ProjectStatus struct {
	Project *Project
	timer   hoursMins
	OnClock bool
	hoursMins
}

// OnClock reports whether the project currently has a running timer and returns it when present.
func (project *Project) OnClock() (bool, *Timer) {
	timer := GetTimer(project)
	return timer.GetStarted() > 0, timer
}

// Status streams a snapshot of the project's current hours and timer state over the provided channel.
// The work is performed asynchronously so callers can aggregate multiple projects concurrently.
func (project *Project) Status(statues chan *ProjectStatus) {
	go func() {
		status := new(ProjectStatus)
		status.Project = project
		hm := project.hoursMins()
		status.hours = hm.hours
		status.mins = hm.mins
		if yes, timer := project.OnClock(); yes {
			status.OnClock = yes
			status.timer = timer.hoursMins()
		}
		statues <- status
	}()
}

func (project *Project) hoursMins() (ret hoursMins) {
	var d int64
	var t *time.Time
	for _, entry := range project.Entries() {
		t, _ = entry.StartedTime()

		if t.Year() == scoppedYear &&
			int(t.Month()) == scoppedMonth {

			d += entry.GetDuration()
		}

	}
	return hoursMinsFromDuration(time.Duration(d))
}

func (t *Timer) hoursMins() hoursMins {
	return hoursMinsFromDuration(t.Duration())
}

// hours and minutes from duration
func hoursMinsFromDuration(d time.Duration) (ret hoursMins) {
	ret.hours = int(d.Hours())
	ret.mins = int(d.Minutes()) - (ret.hours * 60)
	return
}

// HmFromD converts a duration into an hours/minutes pair for display purposes.
func HmFromD(d time.Duration) hoursMins {
	return hoursMinsFromDuration(d)
}

// String returns a trimmed HH:MM representation suitable for aligned terminal output.
func (hm hoursMins) String() string {
	return strings.Trim(fmt.Sprintf("%3d:%02d", hm.hours, hm.mins), " ")
}

// PrintProjectLog prints the 30 most recent entries for the project grouped by day and subtotaled.
func PrintProjectLog(project *Project) {
	printHeader := func(ty *time.Time) {
		headerStr := "           "
		if time.Now().Day() == ty.Day() {
			headerStr = fmt.Sprintf("%s%s \n", headerStr, "Today")
		} else {
			headerStr = fmt.Sprintf("%s%.2d/%.2d \n", headerStr, ty.Month(), ty.Day())
		}

		fmt.Print(util.Color("37", headerStr))
	}

	printTotal := func(total int64) {
		if total != 0 {
			hm := HmFromD(time.Duration(total))
			fmt.Printf("%18s\n", util.Color("32", hm.String()))
		}
	}

	var day int
	var total int64
	fmt.Println(util.Color("33", "\n #  Hours  Date | Description"))

	for i, entry := range project.Entries() {
		if i > 30 {
			break
		}
		ty, err := entry.EndedTime()
		if err == nil && day != ty.Day() {
			printTotal(total)
			printHeader(ty)
			day = ty.Day()
			total = 0
		}
		total = total + *entry.Duration
		fmt.Printf("%2d %s  %.54s", i, entry.HoursMins(), entry.GetContent())
		if len(entry.GetContent()) > 54 {
			fmt.Print("...")
		}
		fmt.Println("")
	}

	printTotal(total)
	fmt.Println()
}

// PrintProjectStatus renders an overview of each project's tracked hours for the given month.
// It derives the year boundary automatically so callers can pass future months in the same year.
func PrintProjectStatus(month int) {
	statues := make(chan *ProjectStatus)
	var counter int
	now := time.Now()
	scoppedMonth = month
	for _, project := range DB.Projects() {
		counter++
		project.Status(statues)
	}

	// if no projects nothing to do
	if counter == 0 {
		return
	}

	if scoppedMonth > int(now.Month()) {
		scoppedYear = now.Year() - 1
	} else {
		scoppedYear = now.Year()
	}

	fmt.Print(util.Color("33", "\n Hours   Clock"),
		" | ",
		util.Color("34", fmt.Sprintln("Projects -", time.Month(scoppedMonth), scoppedYear)),
		"\n")

	for i := 0; i < counter; i++ {
		status := <-statues
		p := status.Project
		name := p.GetName()
		hm := p.hoursMins()

		fmt.Print(util.Color("36", fmt.Sprintf("%s ", hm)))
		if status.OnClock {
			fmt.Print(util.Color("35", fmt.Sprintf(" %s ", status.timer)))
		} else {
			fmt.Print("        ")
		}
		fmt.Print(util.Color("34", fmt.Sprintf("  %s\n", name)))
	}
	fmt.Println("")

}
