package data

import (
	"fmt"
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
