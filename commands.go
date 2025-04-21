package main

import (
	"fmt"

	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/nexneo/samay/data"
	"github.com/nexneo/samay/web"
)

var (
	commands = map[string]func(*data.Project) error{
		"start": startTimer,
		"s":     startTimer,
		"stop":  stopTimer,
		"p":     stopTimer,

		"entry": addEntry,
		"e":     addEntry,
		"rm":    deleteEntryOrProject,
		"show":  showEntryOrProject,

		"mv":  moveProjectOrEntry,
		"log": logProject,

		"report": report,
		"web":    startweb,
	}
)

func startTimer(project *data.Project) (err error) {
	return project.StartTimer()
}

func stopTimer(project *data.Project) (err error) {
	return project.StopTimer(getContent(), billable)
}

func addEntry(project *data.Project) (err error) {
	data.Save(project)
	entry := project.CreateEntryWithDuration(
		getContent(), duration, billable,
	)
	err = data.Save(entry)
	return
}

func showEntryOrProject(project *data.Project) (err error) {
	if theIdx < 0 {
		return showProject(project)
	}
	return showEntry(project)
}

func showProject(project *data.Project) (err error) {
	fmt.Printf("       id : %s\n", project.GetShaFromName())
	fmt.Printf("     name : %s\n", project.GetName())
	fmt.Printf("  entries : %d\n", len(project.Entries()))
	fmt.Printf(" location : %s\n", project.Location())
	return nil
}

func showEntry(project *data.Project) (err error) {
	var started, ended *time.Time
	for i, entry := range project.Entries() {
		if i == theIdx {
			started, err = entry.StartedTime()
			if err != nil {
				return err
			}
			ended, err = entry.EndedTime()
			if err != nil {
				return err
			}
			fmt.Printf("       id : %s\n", entry.GetId())
			fmt.Printf(" contents : %s\n", entry.GetContent())
			fmt.Printf(" duration : %s\n", strings.Trim(entry.HoursMins().String(), " "))
			fmt.Printf("  started : %s\n", started)
			fmt.Printf("    ended : %s\n", ended)
			fmt.Printf("     tags : %v\n", entry.GetTags())
			fmt.Printf(" billable : %t\n", entry.GetBillable())
			break
		}
	}
	return err
}

func deleteEntryOrProject(project *data.Project) (err error) {
	if theIdx < 0 {
		return deleteProject(project)
	}
	return deleteEntry(project)
}

func deleteEntry(project *data.Project) (err error) {
	for i, entry := range project.Entries() {
		if i == theIdx {
			return data.Destroy(entry)
		}
	}
	return
}

func deleteProject(project *data.Project) (err error) {
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

func moveProjectOrEntry(project *data.Project) (err error) {
	newProject := data.CreateProject(newName)
	if err = data.Save(newProject); err != nil {
		return err
	}

	if theIdx < 0 {
		return moveProject(project, newProject)
	}

	return moveEntry(project, newProject)
}

func moveProject(project, newProject *data.Project) error {
	for _, entry := range project.Entries() {
		entry.Project = newProject
		if err := data.Save(entry); err != nil {
			return err
		}
	}

	fmt.Printf("All entries copied to project \"%s\" \n...\n", newProject.GetName())
	return deleteProject(project)
}

func moveEntry(project, newProject *data.Project) error {
	for i, entry := range project.Entries() {
		if i == theIdx {
			entry.Project = newProject
			return data.Save(entry)
		}
	}
	return nil
}

func report(project *data.Project) (err error) {
	if month > 0 && month < 13 {
		data.PrintProjectStatus(month)
	} else {
		err = fmt.Errorf("month %d is not valid", month)
	}
	return
}

func startweb(project *data.Project) error {
	return web.StartServer(httpPort)
}

func logProject(project *data.Project) error {
	data.PrintProjectLog(project)
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
	file, err := os.CreateTemp(os.TempDir(), "samay-buffer")
	if err != nil {
		return "", err
	}
	cmd := exec.Command("subl", "-w", file.Name())
	if err = cmd.Start(); err != nil {
		return "", err
	}
	cmd.Wait()
	data, err := os.ReadFile(file.Name())
	return string(data), err
}
