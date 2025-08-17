package main

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// CommandRunner is an interface that defines a function that runs a CommandRunner that is used by the TmuxRunner.
type CommandRunner interface {
	Run(path string, args []string) ([]byte, error)
	Exec(path string, args []string) error
}

// TmuxCommandRunner is a struct that defines a concrete implementation of the CommandRunner interface.
type TmuxCommandRunner struct{}

// NewTmuxCommandRunner is a constructor for the TmuxCommandRunner struct.
func NewTmuxCommandRunner() *TmuxCommandRunner {
	return &TmuxCommandRunner{}
}

// Run is a function that runs a command with the provided path, and a slice of arguments and return output and error
// path is the path to the command to run: e.g. /usr/bin/tmux
// args is a slice of strings that are the arguments to the command: e.g. []string{"has-session", "-t", "test-session"}
func (t *TmuxCommandRunner) Run(path string, args []string) ([]byte, error) {
	return exec.Command(path, args...).Output()
}

// Exec is a function that executes a command with the provided path and arguments.
// This is used to replace the current process with the new command.
func (t *TmuxCommandRunner) Exec(path string, args []string) error {
	return syscall.Exec(path, args, os.Environ())
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
	if _, err := t.runner.Run(t.path, []string{"has-session", "-t", name}); err != nil {
		return false
	}
	return true
}

// NewSession is a function that creates a new tmux session.
func (t *TmuxRunner) NewSession(s *Session) error {
	return t.runner.Exec(t.path, []string{"tmux", "new-session", "-s", s.name, "-c", s.dir})
}

// AttachSession is a function that attaches to an existing tmux session.
func (t *TmuxRunner) AttachSession(s *Session) error {
	return t.runner.Exec(t.path, []string{"tmux", "attach-session", "-t", s.name})
}

// ListSessions is a function that lists all tmux active sessions.
func (t *TmuxRunner) ListSessions() []*Session {
	var sessions []*Session
	output, err := t.runner.Run(t.path, []string{"list-sessions", "-F", "#{session_name}:#{session_path}"})
	if err != nil {
		return sessions
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) == 0 {
		return sessions
	}

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		sessions = append(sessions, NewSession(parts[0], parts[1], true))
	}
	return sessions
}
