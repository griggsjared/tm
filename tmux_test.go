package main

import (
	"errors"
	"testing"
)

type TestCommandRunner struct {
	result       error
	providedPath string
	providedArgs []string
	providedSc   bool
}

func (t *TestCommandRunner) Run(path string, args []string, sc bool) error {
	t.providedPath = path
	t.providedArgs = args
	t.providedSc = sc
	return t.result
}

func TestNewTmuxCommandRunner(t *testing.T) {
	runner := NewTmuxCommandRunner()
	if runner == nil {
		t.Fatalf("Expected a non-nil TmuxCommandRunner")
	}
}

func TestTmuxRunner_HasSession(t *testing.T) {
	t.Run("session exists", func(t *testing.T) {
		cr := &TestCommandRunner{result: nil} // nil error means command succeeded
		runner := NewTmuxRunner(cr, "/usr/bin/tmux")

		exists := runner.HasSession("test-session")

		if exists != true {
			t.Fatalf("HasSession should return true when command succeeds")
		}

		if cr.providedPath != "/usr/bin/tmux" {
			t.Fatalf("Expected /usr/bin/tmux, got %s", cr.providedPath)
		}

		if cr.providedArgs[0] != "has-session" {
			t.Fatalf("Expected has-session, got %s", cr.providedArgs[0])
		}

		if cr.providedArgs[1] != "-t" {
			t.Fatalf("Expected -t, got %s", cr.providedArgs[1])
		}

		if cr.providedArgs[2] != "test-session" {
			t.Fatalf("Expected test-session, got %s", cr.providedArgs[2])
		}

		if cr.providedSc != false {
			t.Fatalf("Should not use syscall.Exec for has-session")
		}
	})

	t.Run("session doesn't exist", func(t *testing.T) {
		cr := &TestCommandRunner{result: errors.New("session not found")}
		runner := NewTmuxRunner(cr, "/usr/bin/tmux")

		exists := runner.HasSession("non-existent")

		if exists != false {
			t.Fatalf("HasSession should return false when command fails")
		}
	})
}

func TestTmuxRunnzaer_NewSession(t *testing.T) {
	t.Run("successful session creation", func(t *testing.T) {
		cr := &TestCommandRunner{result: nil}
		runner := NewTmuxRunner(cr, "/usr/bin/tmux")
		session := &Session{
			name: "new-test-session",
			dir:  "/tmp",
		}

		err := runner.NewSession(session)

		if err != nil {
			t.Fatalf("NewSession should not return an error when command succeeds")
		}

		if cr.providedPath != "/usr/bin/tmux" {
			t.Fatalf("Expected /usr/bin/tmux, got %s", cr.providedPath)
		}

		if cr.providedArgs[0] != "tmux" {
			t.Fatalf("Expected tmux, got %s", cr.providedArgs[0])
		}

		if cr.providedArgs[1] != "new-session" {
			t.Fatalf("Expected new-session, got %s", cr.providedArgs[1])
		}

		if cr.providedArgs[2] != "-s" {
			t.Fatalf("Expected -s, got %s", cr.providedArgs[2])
		}

		if cr.providedArgs[3] != "new-test-session" {
			t.Fatalf("Expected new-test-session, got %s", cr.providedArgs[3])
		}

		if cr.providedArgs[4] != "-c" {
			t.Fatalf("Expected -c, got %s", cr.providedArgs[4])
		}

		if cr.providedArgs[5] != "/tmp" {
			t.Fatalf("Expected /tmp, got %s", cr.providedArgs[5])
		}

		if cr.providedSc != true {
			t.Fatalf("Should use syscall.Exec for new-session")
		}
	})
}

func TestTmuxRunner_AttachSession(t *testing.T) {
	cr := &TestCommandRunner{result: nil}
	runner := NewTmuxRunner(cr, "/usr/bin/tmux")
	session := &Session{
		name: "existing-session",
		dir:  "/tmp", // Not used in attach
	}

	err := runner.AttachSession(session)

	if err != nil {
		t.Fatalf("AttachSession should not return an error when command succeeds")
	}

	if cr.providedPath != "/usr/bin/tmux" {
		t.Fatalf("Expected /usr/bin/tmux, got %s", cr.providedPath)
	}

	if cr.providedArgs[0] != "tmux" {
		t.Fatalf("Expected tmux, got %s", cr.providedArgs[0])
	}

	if cr.providedArgs[1] != "attach-session" {
		t.Fatalf("Expected attach-session, got %s", cr.providedArgs[1])
	}

	if cr.providedArgs[2] != "-t" {
		t.Fatalf("Expected -t, got %s", cr.providedArgs[2])
	}

	if cr.providedArgs[3] != "existing-session" {
		t.Fatalf("Expected existing-session, got %s", cr.providedArgs[3])
	}

	if cr.providedSc != true {
		t.Fatalf("Should use syscall.Exec for attach-session")
	}
}
