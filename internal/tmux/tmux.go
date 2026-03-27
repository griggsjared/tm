package tmux

import (
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/griggsjared/tm/internal/session"
)

type Runner interface {
	Run(path string, args []string) ([]byte, error)
	Exec(path string, args []string) error
}

type TmuxRunner struct{}

func NewRunner() *TmuxRunner {
	return &TmuxRunner{}
}

func (t *TmuxRunner) Run(path string, args []string) ([]byte, error) {
	return exec.Command(path, args...).Output()
}

func (t *TmuxRunner) Exec(path string, args []string) error {
	return syscall.Exec(path, args, os.Environ())
}

type Repository struct {
	runner Runner
	path   string
}

func NewRepository(r Runner, path string) *Repository {
	return &Repository{
		runner: r,
		path:   path,
	}
}

func (t *Repository) HasSession(name string) bool {
	if _, err := t.runner.Run(t.path, []string{"has-session", "-t", name}); err != nil {
		return false
	}
	return true
}

func (t *Repository) NewSession(s *session.Session) error {
	return t.runner.Exec(t.path, []string{"tmux", "new-session", "-s", s.Name, "-c", s.Dir})
}

func (t *Repository) AttachSession(s *session.Session) error {
	return t.runner.Exec(t.path, []string{"tmux", "attach-session", "-t", s.Name})
}

func (t *Repository) AllSessions() []*session.Session {
	var sessions []*session.Session
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
		sessions = append(sessions, session.New(parts[0], parts[1], true))
	}
	return sessions
}
