package util

import (
	"testing"
)

func TestSHA1(t *testing.T) {
	c := "Project named something"
	if SHA1(c) != "e1129565cc4f7d01c0d4936c65ec4b30d3e00faa" {
		t.Fail()
	}
	c = ""
	if SHA1(c) != "da39a3ee5e6b4b0d3255bfef95601890afd80709" {
		t.Fail()
	}
}
