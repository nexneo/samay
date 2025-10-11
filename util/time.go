package util

import (
	"fmt"
	"strings"
	"time"
)

// HoursMins represents a duration broken into hours and minutes.
type HoursMins struct {
	Hours int
	Mins  int
}

// HoursMinsFromDuration converts a duration to HoursMins with truncated minutes.
func HoursMinsFromDuration(d time.Duration) HoursMins {
	hours := int(d.Hours())
	mins := int(d.Minutes()) - (hours * 60)
	return HoursMins{Hours: hours, Mins: mins}
}

// HmFromD converts a duration into an hours/minutes representation for output.
func HmFromD(d time.Duration) HoursMins {
	return HoursMinsFromDuration(d)
}

// String renders the HoursMins as "H:MM" with left padding similar to prior behavior.
func (hm HoursMins) String() string {
	return strings.Trim(fmt.Sprintf("%3d:%02d", hm.Hours, hm.Mins), " ")
}
