package main

import (
	"errors"
	"testing"
)

type TestCommandRunner struct {
	output       []byte
	error        error
	providedPath string
	providedArgs []string
}

func (t *TestCommandRunner) Run(path string, args []string) ([]byte, error) {
	t.providedPath = path
	t.providedArgs = args
	return t.output, t.error
}

func (t *TestCommandRunner) Exec(path string, args []string) error {
	t.providedPath = path
	t.providedArgs = args
	return t.error
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
		wantExists bool
		wantArgs   []string
		wantPath   string
		wantErr    error
	}{
		{
			name:       "session exists",
			wantExists: true,
			wantArgs:   []string{"has-session", "-t", "test-session"},
			wantPath:   "/usr/bin/tmux",
			wantErr:    nil,
		},
		{
			name:       "session doesn't exist",
			wantExists: false,
			wantArgs:   []string{"has-session", "-t", "non-existent"},
			wantPath:   "/usr/bin/tmux",
			wantErr:    errors.New("session not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &TestCommandRunner{error: tt.wantErr}
			runner := NewTmuxRepository(cr, "/usr/bin/tmux")

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
		})
	}
}

func TestTmuxRunner_NewSession(t *testing.T) {
	tests := []struct {
		name     string
		session  *Session
		wantArgs []string
		wantPath string
		wantErr  error
	}{
		{
			name: "successful session creation",
			session: &Session{
				name: "new-test-session",
				dir:  "/tmp",
			},
			wantArgs: []string{"tmux", "new-session", "-s", "new-test-session", "-c", "/tmp"},
			wantPath: "/usr/bin/tmux",
			wantErr:  nil,
		},
		{
			name: "failed session creation",
			session: &Session{
				name: "failed-session",
				dir:  "/tmp",
			},
			wantArgs: []string{"tmux", "new-session", "-s", "failed-session", "-c", "/tmp"},
			wantPath: "/usr/bin/tmux",
			wantErr:  errors.New("failed to create session"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &TestCommandRunner{error: tt.wantErr}
			runner := NewTmuxRepository(cr, "/usr/bin/tmux")

			err := runner.NewSession(tt.session)

			if err != tt.wantErr {
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
		})
	}
}

func TestTmuxRunner_AttachSession(t *testing.T) {
	tests := []struct {
		name     string
		session  *Session
		wantArgs []string
		wantPath string
		wantErr  error
	}{
		{
			name: "successful attach session",
			session: &Session{
				name: "existing-session",
				dir:  "/tmp",
			},
			wantArgs: []string{"tmux", "attach-session", "-t", "existing-session"},
			wantPath: "/usr/bin/tmux",
			wantErr:  nil,
		},
		{
			name: "failed attach session",
			session: &Session{
				name: "non-existing-session",
				dir:  "/tmp",
			},
			wantArgs: []string{"tmux", "attach-session", "-t", "non-existing-session"},
			wantPath: "/usr/bin/tmux",
			wantErr:  errors.New("session not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &TestCommandRunner{error: tt.wantErr}
			runner := NewTmuxRepository(cr, "/usr/bin/tmux")

			err := runner.AttachSession(tt.session)

			if err != tt.wantErr {
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
		})
	}
}

func TestTmuxRunner_AllSessions(t *testing.T) {

	tests := []struct {
		name       string
		wantArgs   []string
		wantPath   string
		wantOutput []*Session
		crOutput   []byte
	}{
		{
			name:       "list no sessions",
			wantArgs:   []string{"list-sessions", "-F", "#{session_name}:#{session_path}"},
			wantPath:   "/usr/bin/tmux",
			wantOutput: []*Session{},
			crOutput:   []byte(""),
		},
		{
			name:     "list multiple sessions",
			wantArgs: []string{"list-sessions", "-F", "#{session_name}:#{session_path}"},
			wantPath: "/usr/bin/tmux",
			wantOutput: []*Session{
				{name: "session1", dir: "/path/to/session1", exists: true},
				{name: "session2", dir: "/path/to/session2", exists: true},
			},
			crOutput: []byte("session1:/path/to/session1\nsession2:/path/to/session2\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &TestCommandRunner{output: tt.crOutput}
			runner := NewTmuxRepository(cr, "/usr/bin/tmux")

			sessions := runner.AllSessions()

			if len(sessions) != len(tt.wantOutput) {
				t.Fatalf("Expected %d sessions, got %d", len(tt.wantOutput), len(sessions))
			}

			if len(sessions) > 0 {
				for i, session := range sessions {
					if session.name != tt.wantOutput[i].name || session.dir != tt.wantOutput[i].dir {
						t.Fatalf("Expected session %d to be %s:%s, got %s:%s", i, tt.wantOutput[i].name, tt.wantOutput[i].dir, session.name, session.dir)
					}
				}
			}

			if cr.providedPath != tt.wantPath {
				t.Fatalf("Expected path %s, got %s", tt.wantPath, cr.providedPath)
			}

			for i, wantArg := range tt.wantArgs {
				if len(cr.providedArgs) <= i || cr.providedArgs[i] != wantArg {
					t.Fatalf("Expected arg[%d] %s, got %v", i, wantArg, cr.providedArgs)
				}
			}
		})
	}

}
