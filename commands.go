package main

import (
	"fmt"
	"github.com/nexneo/samay/data"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	commands = map[string]func(*data.Project) error{
		"start": startTimer,
		"s":     startTimer,

		"stop": stopTimer,
		"p":    stopTimer,

		"entry": addEntry,
		"e":     addEntry,

		"remove": removeProject,
		"del":    deleteEntry,
		"report": report,
		"log":    logProject,
	}
)

func startTimer(project *data.Project) (err error) {
	data.Save(project)
	timer := data.CreateTimer(project)
	err = data.Save(timer)
	return
}

func stopTimer(project *data.Project) (err error) {
	if yes, timer := project.OnClock(); yes {
		entry := project.CreateEntry(getContent(), billable)

		if err = timer.Stop(entry); err == nil {
			fmt.Printf("%.2f mins\n", entry.Minutes())
		}
	}
	return
}

func addEntry(project *data.Project) (err error) {
	data.Save(project)
	entry := project.CreateEntryWithDuration(
		getContent(), duration, billable,
	)
	err = data.Save(entry)
	return
}

func report(project *data.Project) (err error) {
	if month > 0 && month < 13 {
		data.PrintStatus(month)
	} else {
		err = fmt.Errorf("Month %d is not valid", month)
	}
	return
}

func removeProject(project *data.Project) (err error) {
	var remove string
	fmt.Printf(
		"Remove all data for project \"%s\" ([No]/yes)? ",
		project.GetName(),
	)

	if fmt.Scanln(&remove); remove == "yes" {
		err = data.Destroy(project)
	}
	return
}

func deleteEntry(project *data.Project) (err error) {
	for i, entry := range project.Entries() {
		if i > idx {
			break
		}
		if i == idx {
			err = data.Destroy(entry)
			break
		}
	}
	return
}

func logProject(project *data.Project) error {
	entries := project.Entries()
	fmt.Println("\n #  Hours  Description\n -  -----  -----------")
	printHeader := func(ty *time.Time) {
		fmt.Printf("           - %d/%.2d -\n", ty.Month(), ty.Day())
	}
	var day int
	for i, entry := range entries {
		if i > 30 {
			break
		}

		ty, err := entry.EndedTime()
		if err == nil && day != ty.Day() {
			printHeader(ty)
			day = ty.Day()
		}

		fmt.Printf("%2d %s  %.54s", i, entry.HoursMins(), entry.GetContent())
		if len(entry.GetContent()) > 54 {
			fmt.Print("...")
		}
		fmt.Println("")
	}
	fmt.Println()
	return nil
}

// Set timelog entry content and tags using external editor.
// Only if it didn't set via -m flag
func getContent() string {
	if content == "" {
		content, _ = openEditor()
	}
	return content
}

// Open external editor for text input from user
func openEditor() (string, error) {
	file, err := ioutil.TempFile(os.TempDir(), "subl")
	args := strings.Split(os.Getenv("EDITOR"), " ")
	args = append(args, file.Name())
	cmd := exec.Command(args[0], args[1:]...)
	if err = cmd.Start(); err != nil {
		return "", err
	}
	cmd.Wait()
	data, err := ioutil.ReadFile(file.Name())
	return string(data), err
}
