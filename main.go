package main

import (
	"fmt"
	"os/exec"
)

func main() {
	service := NewCommandService(RealCommandRunner{})
	if !service.tmuxInstalled() {
		fmt.Println("tmux is not installed")
		return
	}

	fmt.Println("tmux is installed and we are good to go")
}

type CommandRunner interface {
	Run(name string, args ...string) error
}

type RealCommandRunner struct{}

func (r RealCommandRunner) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

type CommandService struct {
	runner CommandRunner
}

func NewCommandService(runner CommandRunner) *CommandService {
	return &CommandService{runner: runner}
}

func (s CommandService) tmuxInstalled() bool {
	err := s.runner.Run("which", "tmux")
	return err == nil
}
