package data

import (
	"fmt"
	"github.com/nexneo/samay/util"
	"time"
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

func (project *Project) OnClock() (bool, *Timer) {
	timer := GetTimer(project)
	return timer.GetStarted() > 0, timer
}

// fetch status of project in goroutine and return on channel
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

func HmFromD(d time.Duration) hoursMins {
	return hoursMinsFromDuration(d)
}

// consistent string presentation of hoursMins
func (hm hoursMins) String() string {
	return fmt.Sprintf("%3d:%02d", hm.hours, hm.mins)
}

// simple function that return repeated Char in string.
func repeatChar(c string, length int) (ret string) {
	for i := 0; i < length; i++ {
		ret += c
	}
	return
}

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

func PrintStatus(month int) {
	statues := make(chan *ProjectStatus)
	var counter, longest, size int
	now := time.Now()
	scoppedMonth = month
	for _, project := range DB.Projects() {
		counter++
		project.Status(statues)
		if size = len(project.GetName()); size > longest {
			longest = size
		}
	}

	// if no projects nothing to do
	if counter == 0 {
		return
	}

	if (longest % 2) != 0 {
		longest += 1
	}

	if scoppedMonth > int(now.Month()) {
		scoppedYear = now.Year() - 1
	} else {
		scoppedYear = now.Year()
	}

	printStatusHeader(longest)

	for i := 0; i < counter; i++ {
		status := <-statues
		name := status.Project.GetName()
		padding := longest - len(name) + 2
		fmt.Printf(
			"| %s%s| %3d:%02d | ",
			name,
			repeatChar(" ", padding),
			status.hours, status.mins,
		)
		if status.OnClock {
			fmt.Printf("%s |", status.timer)
		} else {
			fmt.Print("       |")
		}
		fmt.Println()
	}

	printRuler(longest)
}

func printStatusHeader(longest int) {
	fmt.Println("\nReport for", time.Month(scoppedMonth), scoppedYear)

	padding := repeatChar(" ", (longest-4)/2)
	printRuler(longest)
	fmt.Printf("|%sProject%s|  Hours |  Clock |\n", padding, padding)
	printRuler(longest)
}

func printRuler(longest int) {
	column := 1 + 6 + 1 + 2
	ruler := repeatChar("-", longest+3+column+column)
	fmt.Println(ruler)
}
