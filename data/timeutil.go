package data

import (
	"fmt"
	"strings"
	"time"
)

type hoursMins struct {
	hours int
	mins  int
}

func hoursMinsFromDuration(d time.Duration) hoursMins {
	hours := int(d.Hours())
	mins := int(d.Minutes()) - (hours * 60)
	return hoursMins{hours: hours, mins: mins}
}

// HmFromD converts a duration into an hours/minutes representation for output.
func HmFromD(d time.Duration) hoursMins {
	return hoursMinsFromDuration(d)
}

func (hm hoursMins) String() string {
	return strings.Trim(fmt.Sprintf("%3d:%02d", hm.hours, hm.mins), " ")
}
