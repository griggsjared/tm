package doctor

import (
	"fmt"
	"strings"
)

// TmuxClient provides tmux availability and path information
type TmuxClient interface {
	IsAvailable() bool
	Path() string
}

// FzfRunner provides fzf availability and path information
type FzfRunner interface {
	IsAvailable() bool
	Path() string
}

// Doctor checks the health of tm dependencies
type Doctor struct {
	version    string
	tmuxClient TmuxClient
	fzfRunner  FzfRunner
}

// New creates a new Doctor instance
func New(version string, tmuxClient TmuxClient, fzfRunner FzfRunner) *Doctor {
	return &Doctor{
		version:    version,
		tmuxClient: tmuxClient,
		fzfRunner:  fzfRunner,
	}
}

const dotWidth = 12

func printStatusLine(name, status string) {
	dots := strings.Repeat(".", dotWidth-len(name))
	fmt.Printf("%s%s %s\n", name, dots, status)
}

// Run performs the health check and returns the exit code
func (d *Doctor) Run() int {
	exitCode := 0

	printStatusLine("tm", fmt.Sprintf("ok (%s)", d.version))

	if d.tmuxClient.IsAvailable() {
		printStatusLine("tmux", fmt.Sprintf("ok (%s)", d.tmuxClient.Path()))
	} else {
		printStatusLine("tmux", "missing")
		exitCode = 1
	}

	if d.fzfRunner.IsAvailable() {
		printStatusLine("fzf", fmt.Sprintf("ok (%s) (optional)", d.fzfRunner.Path()))
	} else {
		printStatusLine("fzf", "missing (optional)")
	}

	return exitCode
}
