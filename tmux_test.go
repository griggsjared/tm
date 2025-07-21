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
	tests := []struct {
		name       string
		cmdResult  error
		wantExists bool
		wantArgs   []string
		wantPath   string
		wantSc     bool
	}{
		{
			name:       "session exists",
			cmdResult:  nil,
			wantExists: true,
			wantArgs:   []string{"has-session", "-t", "test-session"},
			wantPath:   "/usr/bin/tmux",
			wantSc:     false,
		},
		{
			name:       "session doesn't exist",
			cmdResult:  errors.New("session not found"),
			wantExists: false,
			wantArgs:   []string{"has-session", "-t", "non-existent"},
			wantPath:   "/usr/bin/tmux",
			wantSc:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &TestCommandRunner{result: tt.cmdResult}
			runner := NewTmuxRunner(cr, "/usr/bin/tmux")

			var exists bool
			if tt.name == "session exists" {
				exists = runner.HasSession("test-session")
			} else {
				exists = runner.HasSession("non-existent")
			}

			if exists != tt.wantExists {
				t.Fatalf("HasSession should return %v, got %v", tt.wantExists, exists)
			}

			if cr.providedPath != tt.wantPath {
				t.Fatalf("Expected path %s, got %s", tt.wantPath, cr.providedPath)
			}

			for i, wantArg := range tt.wantArgs {
				if len(cr.providedArgs) <= i || cr.providedArgs[i] != wantArg {
					t.Fatalf("Expected arg[%d] %s, got %v", i, wantArg, cr.providedArgs)
				}
			}

			if cr.providedSc != tt.wantSc {
				t.Fatalf("Expected syscall.Exec %v, got %v", tt.wantSc, cr.providedSc)
			}
		})
	}
}

func TestTmuxRunner_NewSession(t *testing.T) {
	tests := []struct {
		name      string
		session   *Session
		cmdResult error
		wantArgs  []string
		wantPath  string
		wantSc    bool
		wantErr   bool
	}{
		{
			name: "successful session creation",
			session: &Session{
				name: "new-test-session",
				dir:  "/tmp",
			},
			cmdResult: nil,
			wantArgs:  []string{"tmux", "new-session", "-s", "new-test-session", "-c", "/tmp"},
			wantPath:  "/usr/bin/tmux",
			wantSc:    true,
			wantErr:   false,
		},
		{
			name: "failed session creation",
			session: &Session{
				name: "failed-session",
				dir:  "/tmp",
			},
			cmdResult: errors.New("failed to create session"),
			wantArgs:  []string{"tmux", "new-session", "-s", "failed-session", "-c", "/tmp"},
			wantPath:  "/usr/bin/tmux",
			wantSc:    true,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &TestCommandRunner{result: tt.cmdResult}
			runner := NewTmuxRunner(cr, "/usr/bin/tmux")

			err := runner.NewSession(tt.session)

			if (err != nil) != tt.wantErr {
				t.Fatalf("NewSession error = %v, wantErr %v", err, tt.wantErr)
			}

			if cr.providedPath != tt.wantPath {
				t.Fatalf("Expected path %s, got %s", tt.wantPath, cr.providedPath)
			}

			for i, wantArg := range tt.wantArgs {
				if len(cr.providedArgs) <= i || cr.providedArgs[i] != wantArg {
					t.Fatalf("Expected arg[%d] %s, got %v", i, wantArg, cr.providedArgs)
				}
			}

			if cr.providedSc != tt.wantSc {
				t.Fatalf("Expected syscall.Exec %v, got %v", tt.wantSc, cr.providedSc)
			}
		})
	}
}

func TestTmuxRunner_AttachSession(t *testing.T) {
	tests := []struct {
		name      string
		session   *Session
		cmdResult error
		wantArgs  []string
		wantPath  string
		wantSc    bool
		wantErr   bool
	}{
		{
			name: "successful attach session",
			session: &Session{
				name: "existing-session",
				dir:  "/tmp",
			},
			cmdResult: nil,
			wantArgs:  []string{"tmux", "attach-session", "-t", "existing-session"},
			wantPath:  "/usr/bin/tmux",
			wantSc:    true,
			wantErr:   false,
		},
		{
			name: "failed attach session",
			session: &Session{
				name: "non-existing-session",
				dir:  "/tmp",
			},
			cmdResult: errors.New("session not found"),
			wantArgs:  []string{"tmux", "attach-session", "-t", "non-existing-session"},
			wantPath:  "/usr/bin/tmux",
			wantSc:    true,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &TestCommandRunner{result: tt.cmdResult}
			runner := NewTmuxRunner(cr, "/usr/bin/tmux")

			err := runner.AttachSession(tt.session)

			if (err != nil) != tt.wantErr {
				t.Fatalf("AttachSession error = %v, wantErr %v", err, tt.wantErr)
			}

			if cr.providedPath != tt.wantPath {
				t.Fatalf("Expected path %s, got %s", tt.wantPath, cr.providedPath)
			}

			for i, wantArg := range tt.wantArgs {
				if len(cr.providedArgs) <= i || cr.providedArgs[i] != wantArg {
					t.Fatalf("Expected arg[%d] %s, got %v", i, wantArg, cr.providedArgs)
				}
			}

			if cr.providedSc != tt.wantSc {
				t.Fatalf("Expected syscall.Exec %v, got %v", tt.wantSc, cr.providedSc)
			}
		})
	}
}
