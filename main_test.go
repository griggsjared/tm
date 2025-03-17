package main

import (
	"errors"
	"testing"
)

type MockCommandRunner struct {
	shouldError bool
}

func (m MockCommandRunner) Run(name string, args ...string) error {
	if m.shouldError {
		return errors.New("command not found")
	}
	return nil
}

func TestTmuxInstalled(t *testing.T) {
	// Test case: tmux is installed
	successRunner := MockCommandRunner{shouldError: false}
	successService := NewCommandService(successRunner)

	if !successService.tmuxInstalled() {
		t.Error("Expected tmux to be reported as installed")
	}

	// Test case: tmux is not installed
	failRunner := MockCommandRunner{shouldError: true}
	failService := NewCommandService(failRunner)

	if failService.tmuxInstalled() {
		t.Error("Expected tmux to be reported as not installed")
	}
}
