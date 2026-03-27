package tmux

import (
	"errors"
	"testing"

	"github.com/griggsjared/tm/internal/session"
)

type TestRunner struct {
	output       []byte
	error        error
	providedPath string
	providedArgs []string
}

func (t *TestRunner) Run(path string, args []string) ([]byte, error) {
	t.providedPath = path
	t.providedArgs = args
	return t.output, t.error
}

func (t *TestRunner) Exec(path string, args []string) error {
	t.providedPath = path
	t.providedArgs = args
	return t.error
}

func TestNewRunner(t *testing.T) {
	runner := NewRunner()
	if runner == nil {
		t.Fatalf("Expected a non-nil TmuxCommandRunner")
	}
}

func TestRepository_HasSession(t *testing.T) {
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
			cr := &TestRunner{error: tt.wantErr}
			repo := NewRepository(cr, "/usr/bin/tmux")

			var exists bool
			if tt.name == "session exists" {
				exists = repo.HasSession("test-session")
			} else {
				exists = repo.HasSession("non-existent")
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

func TestRepository_NewSession(t *testing.T) {
	tests := []struct {
		name     string
		sess     *session.Session
		wantArgs []string
		wantPath string
		wantErr  error
	}{
		{
			name: "successful session creation",
			sess: &session.Session{
				Name: "new-test-session",
				Dir:  "/tmp",
			},
			wantArgs: []string{"tmux", "new-session", "-s", "new-test-session", "-c", "/tmp"},
			wantPath: "/usr/bin/tmux",
			wantErr:  nil,
		},
		{
			name: "failed session creation",
			sess: &session.Session{
				Name: "failed-session",
				Dir:  "/tmp",
			},
			wantArgs: []string{"tmux", "new-session", "-s", "failed-session", "-c", "/tmp"},
			wantPath: "/usr/bin/tmux",
			wantErr:  errors.New("failed to create session"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &TestRunner{error: tt.wantErr}
			repo := NewRepository(cr, "/usr/bin/tmux")

			err := repo.NewSession(tt.sess)

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

func TestRepository_AttachSession(t *testing.T) {
	tests := []struct {
		name     string
		sess     *session.Session
		wantArgs []string
		wantPath string
		wantErr  error
	}{
		{
			name: "successful attach session",
			sess: &session.Session{
				Name: "existing-session",
				Dir:  "/tmp",
			},
			wantArgs: []string{"tmux", "attach-session", "-t", "existing-session"},
			wantPath: "/usr/bin/tmux",
			wantErr:  nil,
		},
		{
			name: "failed attach session",
			sess: &session.Session{
				Name: "non-existing-session",
				Dir:  "/tmp",
			},
			wantArgs: []string{"tmux", "attach-session", "-t", "non-existing-session"},
			wantPath: "/usr/bin/tmux",
			wantErr:  errors.New("session not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &TestRunner{error: tt.wantErr}
			repo := NewRepository(cr, "/usr/bin/tmux")

			err := repo.AttachSession(tt.sess)

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

func TestRepository_AllSessions(t *testing.T) {
	tests := []struct {
		name       string
		wantArgs   []string
		wantPath   string
		wantOutput []*session.Session
		crOutput   []byte
	}{
		{
			name:       "list no sessions",
			wantArgs:   []string{"list-sessions", "-F", "#{session_name}:#{session_path}"},
			wantPath:   "/usr/bin/tmux",
			wantOutput: []*session.Session{},
			crOutput:   []byte(""),
		},
		{
			name:     "list multiple sessions",
			wantArgs: []string{"list-sessions", "-F", "#{session_name}:#{session_path}"},
			wantPath: "/usr/bin/tmux",
			wantOutput: []*session.Session{
				{Name: "session1", Dir: "/path/to/session1", Exists: true},
				{Name: "session2", Dir: "/path/to/session2", Exists: true},
			},
			crOutput: []byte("session1:/path/to/session1\nsession2:/path/to/session2\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &TestRunner{output: tt.crOutput}
			repo := NewRepository(cr, "/usr/bin/tmux")

			sessions := repo.AllSessions()

			if len(sessions) != len(tt.wantOutput) {
				t.Fatalf("Expected %d sessions, got %d", len(tt.wantOutput), len(sessions))
			}

			if len(sessions) > 0 {
				for i, sess := range sessions {
					if sess.Name != tt.wantOutput[i].Name || sess.Dir != tt.wantOutput[i].Dir {
						t.Fatalf("Expected session %d to be %s:%s, got %s:%s", i, tt.wantOutput[i].Name, tt.wantOutput[i].Dir, sess.Name, sess.Dir)
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
