package main

import (
	"os"
	"os/exec"
	"syscall"
)

type CommandRunner interface {
	Run(path string, args []string, sc bool) error
}

type TmuxCommandRunner struct{}

func NewTmuxCommandRunner() *TmuxCommandRunner {
	return &TmuxCommandRunner{}
}

func (t *TmuxCommandRunner) Run(path string, args []string, sc bool) error {
	switch sc {
	case true:
		return syscall.Exec(path, args, os.Environ())
	default:
		return exec.Command(path, args...).Run()
	}
}

type TmuxRunner struct {
	runner CommandRunner
	path   string // path to tmux
}

func NewTmuxRunner(runner CommandRunner, path string) *TmuxRunner {
	return &TmuxRunner{
		runner: runner,
		path:   path,
	}
}

func (t *TmuxRunner) HasSession(name string) bool {
	return t.runner.Run(t.path, []string{"has-session", "-t", name}, false) == nil
}

func (t *TmuxRunner) NewSession(s *Session) error {
	err := os.Chdir(s.dir)
	if err != nil {
		return err
	}

	return t.runner.Run(t.path, []string{"tmux", "new-session", "-s", s.name, "-c", s.dir}, true)
}

func (t *TmuxRunner) AttachSession(s *Session) error {
	return t.runner.Run(t.path, []string{"tmux", "attach-session", "-t", s.name}, true)
}
