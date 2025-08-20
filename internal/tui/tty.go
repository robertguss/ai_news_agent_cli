package tui

import (
	"os"

	"golang.org/x/term"
)

func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func ShouldUseTUI() bool {
	if os.Getenv("CI") != "" {
		return false
	}
	
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	
	return IsTTY()
}
