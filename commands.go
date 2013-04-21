package main

import (
	"fmt"
	"github.com/nexneo/samay/data"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
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

		"mv":  moveProject,
		"log": logProject,

		"report": report,
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

func showEntryOrProject(project *data.Project) (err error) {
	if idx < 0 {
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
		if i > idx {
			break
		}
		if i == idx {
			started, err = entry.StartedTime()
			ended, err = entry.EndedTime()
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
	if idx < 0 {
		return deleteProject(project)
	}
	return deleteEntry(project)
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

func moveProject(project *data.Project) (err error) {
	newProject := data.CreateProject(newName)
	if err = data.Save(newProject); err != nil {
		return err
	}
	for _, entry := range project.Entries() {
		entry.Project = newProject
		err = data.Save(entry)
		if err != nil {
			return err
		}
	}
	fmt.Printf("All entries copied to project \"%s\" \n...\n", newProject.GetName())
	deleteProject(project)
	return nil
}

func report(project *data.Project) (err error) {
	if month > 0 && month < 13 {
		data.PrintStatus(month)
	} else {
		err = fmt.Errorf("Month %d is not valid", month)
	}
	return
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
