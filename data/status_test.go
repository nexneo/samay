package data_test

import (
	"fmt"
	"testing"

	"github.com/nexneo/samay/data"
)

func TestListProjects(t *testing.T) {
	var project *data.Project
	for _, project = range data.DB.Projects() {
		if ok, _ := project.OnClock(); ok {
			fmt.Printf("Project: %s (ticking...)\n",
				project.GetName(),
			)
		}
	}
}
