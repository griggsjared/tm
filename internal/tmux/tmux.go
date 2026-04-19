package tmux

import (
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/griggsjared/tm/internal/session"
)

type Runner interface {
	Output(path string, args []string) ([]byte, error)
	Exec(path string, args []string) error
}

type TmuxRunner struct{}

func NewRunner() *TmuxRunner {
	return &TmuxRunner{}
}

func (t *TmuxRunner) Output(path string, args []string) ([]byte, error) {
	return exec.Command(path, args...).Output()
}

func (t *TmuxRunner) Exec(path string, args []string) error {
	return syscall.Exec(path, args, os.Environ())
}

type Client struct {
	runner Runner
	path   string
}

func NewClient(r Runner, path string) *Client {
	return &Client{
		runner: r,
		path:   path,
	}
}

func (c *Client) IsAvailable() bool {
	return c.path != ""
}

func (c *Client) HasSession(name string) bool {
	if _, err := c.runner.Output(c.path, []string{"has-session", "-t", "=" + name}); err != nil {
		return false
	}
	return true
}

func (c *Client) NewSession(s *session.Session, detached bool) error {
	if detached {
		_, err := c.runner.Output(c.path, []string{"new-session", "-d", "-s", s.Name, "-c", s.Dir})
		return err
	}
	return c.runner.Exec(c.path, []string{"tmux", "new-session", "-s", s.Name, "-c", s.Dir})
}

func (c *Client) AttachSession(s *session.Session) error {
	return c.runner.Exec(c.path, []string{"tmux", "attach-session", "-t", s.Name})
}

func (c *Client) SwitchSession(s *session.Session) error {
	return c.runner.Exec(c.path, []string{"tmux", "switch-client", "-t", s.Name})
}

func (c *Client) AllSessions() []*session.Session {
	var sessions []*session.Session
	output, err := c.runner.Output(c.path, []string{"list-sessions", "-F", "#{session_name}:#{session_path}"})
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
