package util

import (
	"testing"
	"time"
)

func TestHoursMinsFromDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want HoursMins
	}{
		{
			name: "zero duration",
			d:    0,
			want: HoursMins{Hours: 0, Mins: 0},
		},
		{
			name: "minutes only",
			d:    59 * time.Minute,
			want: HoursMins{Hours: 0, Mins: 59},
		},
		{
			name: "hours and minutes",
			d:    3*time.Hour + 15*time.Minute,
			want: HoursMins{Hours: 3, Mins: 15},
		},
		{
			name: "truncate partial minutes",
			d:    1*time.Hour + 45*time.Second,
			want: HoursMins{Hours: 1, Mins: 0},
		},
		{
			name: "large duration",
			d:    28*time.Hour + 125*time.Minute,
			want: HoursMins{Hours: 30, Mins: 5},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := HoursMinsFromDuration(tc.d)
			if got != tc.want {
				t.Fatalf("HoursMinsFromDuration(%v) = %+v, want %+v", tc.d, got, tc.want)
			}
			if indirect := HmFromD(tc.d); indirect != tc.want {
				t.Fatalf("HmFromD(%v) = %+v, want %+v", tc.d, indirect, tc.want)
			}
		})
	}
}

func TestHoursMinsString(t *testing.T) {
	tests := []struct {
		hm   HoursMins
		want string
	}{
		{hm: HoursMins{Hours: 0, Mins: 0}, want: "0:00"},
		{hm: HoursMins{Hours: 0, Mins: 5}, want: "0:05"},
		{hm: HoursMins{Hours: 2, Mins: 7}, want: "2:07"},
		{hm: HoursMins{Hours: 12, Mins: 45}, want: "12:45"},
	}

	for _, tc := range tests {
		if got := tc.hm.String(); got != tc.want {
			t.Fatalf("HoursMins{Hours:%d, Mins:%d}.String() = %q, want %q", tc.hm.Hours, tc.hm.Mins, got, tc.want)
		}
	}
}
