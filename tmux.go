package main

import (
	"os"
	"os/exec"
	"syscall"
)

// CommandRunner is an interface that defines a function that runs a CommandRunner that is used by the TmuxRunner.
type CommandRunner interface {
	Run(path string, args []string, sc bool) error
}

// TmuxCommandRunner is a struct that defines a concrete implementation of the CommandRunner interface.
type TmuxCommandRunner struct{}

// NewTmuxCommandRunner is a constructor for the TmuxCommandRunner struct.
func NewTmuxCommandRunner() *TmuxCommandRunner {
	return &TmuxCommandRunner{}
}

// Run is a function that runs a command with the provided path, args, and sc.
// path is the path to the command to run: e.g. /usr/bin/tmux
// args is a slice of strings that are the arguments to the command: e.g. []string{"has-session", "-t", "test-session"}
// sc is a boolean that determines whether to use syscall.Exec or exec.Command to run the command. I dont know what else to call it.
func (t *TmuxCommandRunner) Run(path string, args []string, sc bool) error {
	switch sc {
	case true:
		return syscall.Exec(path, args, os.Environ())
	default:
		return exec.Command(path, args...).Run()
	}
}

// TmuxRunner is a struct that handles running tmux commands.
type TmuxRunner struct {
	runner CommandRunner
	path   string // path to tmux
}

// MewTmuxRunner is a constructor for the TmuxRunner struct.
func NewTmuxRunner(runner CommandRunner, path string) *TmuxRunner {
	return &TmuxRunner{
		runner: runner,
		path:   path,
	}
}

// HasSession is a function that checks if a session exists.
func (t *TmuxRunner) HasSession(name string) bool {
	return t.runner.Run(t.path, []string{"has-session", "-t", name}, false) == nil
}

// NewSession is a function that creates a new tmux session.
func (t *TmuxRunner) NewSession(s *Session) error {
	return t.runner.Run(t.path, []string{"tmux", "new-session", "-s", s.name, "-c", s.dir}, true)
}

// AttachSession is a function that attaches to an existing tmux session.
func (t *TmuxRunner) AttachSession(s *Session) error {
	return t.runner.Run(t.path, []string{"tmux", "attach-session", "-t", s.name}, true)
}
