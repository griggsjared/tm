package tmux

import (
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/griggsjared/tm/internal/session"
)

// Runner defines an interface for running and executing tmux commands.
type Runner interface {
	Output(path string, args []string) ([]byte, error)
	Exec(path string, args []string) error
}

// TmuxRunner is a concrete implementation of Runner that uses the os/exec package to run tmux commands.
type TmuxRunner struct{}

// NewRunner creates a new instance of TmuxRunner.
func NewRunner() *TmuxRunner {
	return &TmuxRunner{}
}

// Output executes a tmux command and returns its output or an error if the command fails.
func (t *TmuxRunner) Output(path string, args []string) ([]byte, error) {
	return exec.Command(path, args...).Output()
}

// Exec executes a tmux command and replaces the current process with the new command.
func (t *TmuxRunner) Exec(path string, args []string) error {
	return syscall.Exec(path, args, os.Environ())
}

// Repository provides methods to interact with tmux sessions using a Runner.
type Repository struct {
	runner Runner
	path   string
}

// NewRepository creates a new Repository with the given Runner and tmux path.
func NewRepository(r Runner, path string) *Repository {
	return &Repository{
		runner: r,
		path:   path,
	}
}

// HasSession checks if a tmux session with the given name exists.
func (t *Repository) HasSession(name string) bool {
	if _, err := t.runner.Output(t.path, []string{"has-session", "-t", name}); err != nil {
		return false
	}
	return true
}

// NewSession creates a new tmux session with the given Session details.
func (t *Repository) NewSession(s *session.Session) error {
	return t.runner.Exec(t.path, []string{"tmux", "new-session", "-s", s.Name, "-c", s.Dir})
}

// AttachSession attaches to an existing tmux session with the given Session details.
func (t *Repository) AttachSession(s *session.Session) error {
	return t.runner.Exec(t.path, []string{"tmux", "attach-session", "-t", s.Name})
}

// AllSessions retrieves all tmux sessions and returns them as a slice of Session pointers.
func (t *Repository) AllSessions() []*session.Session {
	var sessions []*session.Session
	output, err := t.runner.Output(t.path, []string{"list-sessions", "-F", "#{session_name}:#{session_path}"})
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
		sessions = append(sessions, session.New(parts[0], parts[1], true))
	}
	return sessions
}
