package version

import "fmt"

// Number represents the semantic version of the CLI.
const Number = "1.2.0"

// Name is the human-friendly identifier used when printing the version string.
const Name = "samay"

// String returns the formatted CLI version string.
func String() string {
	return fmt.Sprintf("%s %s", Name, Number)
}
