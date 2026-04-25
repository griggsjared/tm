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
	calledWith   string // "output" or "exec"
}

func (t *TestRunner) Output(path string, args []string) ([]byte, error) {
	t.providedPath = path
	t.providedArgs = args
	t.calledWith = "output"
	return t.output, t.error
}

func (t *TestRunner) Exec(path string, args []string) error {
	t.providedPath = path
	t.providedArgs = args
	t.calledWith = "exec"
	return t.error
}

func TestNewRunner(t *testing.T) {
	runner := NewRunner()
	if runner == nil {
		t.Fatalf("Expected a non-nil TmuxCommandRunner")
	}
}

func TestClient_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantAvail bool
	}{
		{
			name:      "with path",
			path:      "/usr/bin/tmux",
			wantAvail: true,
		},
		{
			name:      "empty path",
			path:      "",
			wantAvail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &TestRunner{}
			client := NewClient(cr, tt.path)

			got := client.IsAvailable()
			if got != tt.wantAvail {
				t.Fatalf("IsAvailable() = %v, want %v", got, tt.wantAvail)
			}
		})
	}
}

func TestClient_Path(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantPath string
	}{
		{
			name:     "with path",
			path:     "/usr/bin/tmux",
			wantPath: "/usr/bin/tmux",
		},
		{
			name:     "empty path",
			path:     "",
			wantPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &TestRunner{}
			client := NewClient(cr, tt.path)

			got := client.Path()
			if got != tt.wantPath {
				t.Fatalf("Path() = %v, want %v", got, tt.wantPath)
			}
		})
	}
}

func TestClient_HasSession(t *testing.T) {
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
			wantArgs:   []string{"has-session", "-t", "=test-session"},
			wantPath:   "/usr/bin/tmux",
			wantErr:    nil,
		},
		{
			name:       "session doesn't exist",
			wantExists: false,
			wantArgs:   []string{"has-session", "-t", "=non-existent"},
			wantPath:   "/usr/bin/tmux",
			wantErr:    errors.New("session not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &TestRunner{error: tt.wantErr}
			client := NewClient(cr, "/usr/bin/tmux")

			var exists bool
			if tt.name == "session exists" {
				exists = client.HasSession("test-session")
			} else {
				exists = client.HasSession("non-existent")
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

func TestClient_NewSession(t *testing.T) {
	tests := []struct {
		name           string
		sess           *session.Session
		detached       bool
		wantArgs       []string
		wantPath       string
		wantCalledWith string
		wantErr        error
	}{
		{
			name:           "non-detached: successful session creation",
			sess:           &session.Session{Name: "new-test-session", Dir: "/tmp"},
			detached:       false,
			wantArgs:       []string{"tmux", "new-session", "-s", "new-test-session", "-c", "/tmp"},
			wantPath:       "/usr/bin/tmux",
			wantCalledWith: "exec",
			wantErr:        nil,
		},
		{
			name:           "non-detached: failed session creation",
			sess:           &session.Session{Name: "failed-session", Dir: "/tmp"},
			detached:       false,
			wantArgs:       []string{"tmux", "new-session", "-s", "failed-session", "-c", "/tmp"},
			wantPath:       "/usr/bin/tmux",
			wantCalledWith: "exec",
			wantErr:        errors.New("failed to create session"),
		},
		{
			name:           "detached: successful session creation",
			sess:           &session.Session{Name: "new-test-session", Dir: "/tmp"},
			detached:       true,
			wantArgs:       []string{"new-session", "-d", "-s", "new-test-session", "-c", "/tmp"},
			wantPath:       "/usr/bin/tmux",
			wantCalledWith: "output",
			wantErr:        nil,
		},
		{
			name:           "detached: failed session creation",
			sess:           &session.Session{Name: "failed-session", Dir: "/tmp"},
			detached:       true,
			wantArgs:       []string{"new-session", "-d", "-s", "failed-session", "-c", "/tmp"},
			wantPath:       "/usr/bin/tmux",
			wantCalledWith: "output",
			wantErr:        errors.New("failed to create session"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &TestRunner{error: tt.wantErr}
			client := NewClient(cr, "/usr/bin/tmux")

			err := client.NewSession(tt.sess, tt.detached)

			if err != tt.wantErr {
				t.Fatalf("NewSession error = %v, wantErr %v", err, tt.wantErr)
			}
			if cr.calledWith != tt.wantCalledWith {
				t.Fatalf("Expected runner method %q, got %q", tt.wantCalledWith, cr.calledWith)
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

func TestClient_AttachSession(t *testing.T) {
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
			client := NewClient(cr, "/usr/bin/tmux")

			err := client.AttachSession(tt.sess)

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

func TestClient_SwitchSession(t *testing.T) {
	tests := []struct {
		name     string
		sess     *session.Session
		wantArgs []string
		wantPath string
		wantErr  error
	}{
		{
			name: "successful switch session",
			sess: &session.Session{
				Name: "existing-session",
				Dir:  "/tmp",
			},
			wantArgs: []string{"tmux", "switch-client", "-t", "existing-session"},
			wantPath: "/usr/bin/tmux",
			wantErr:  nil,
		},
		{
			name: "failed switch session",
			sess: &session.Session{
				Name: "non-existing-session",
				Dir:  "/tmp",
			},
			wantArgs: []string{"tmux", "switch-client", "-t", "non-existing-session"},
			wantPath: "/usr/bin/tmux",
			wantErr:  errors.New("session not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &TestRunner{error: tt.wantErr}
			client := NewClient(cr, "/usr/bin/tmux")

			err := client.SwitchSession(tt.sess)

			if err != tt.wantErr {
				t.Fatalf("SwitchSession error = %v, wantErr %v", err, tt.wantErr)
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

func TestClient_CurrentSession(t *testing.T) {
	tests := []struct {
		name       string
		crOutput   []byte
		crError    error
		wantResult string
		wantArgs   []string
		wantPath   string
	}{
		{
			name:       "successful current session",
			crOutput:   []byte("my-session\n"),
			wantResult: "my-session",
			wantArgs:   []string{"display-message", "-p", "#S"},
			wantPath:   "/usr/bin/tmux",
		},
		{
			name:       "error returns empty string",
			crError:    errors.New("tmux error"),
			wantResult: "",
			wantArgs:   []string{"display-message", "-p", "#S"},
			wantPath:   "/usr/bin/tmux",
		},
		{
			name:       "trims whitespace",
			crOutput:   []byte("  session-name  \n"),
			wantResult: "session-name",
			wantArgs:   []string{"display-message", "-p", "#S"},
			wantPath:   "/usr/bin/tmux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &TestRunner{output: tt.crOutput, error: tt.crError}
			client := NewClient(cr, "/usr/bin/tmux")

			got := client.CurrentSession()
			if got != tt.wantResult {
				t.Fatalf("CurrentSession() = %q, want %q", got, tt.wantResult)
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

func TestClient_AllSessions(t *testing.T) {
	tests := []struct {
		name       string
		wantArgs   []string
		wantPath   string
		wantOutput []*session.Session
		crOutput   []byte
	}{
		{
			name:       "list no sessions",
			wantArgs:   []string{"list-sessions", "-F", listSessionsFormat},
			wantPath:   "/usr/bin/tmux",
			wantOutput: []*session.Session{},
			crOutput:   []byte(""),
		},
		{
			name:     "list multiple sessions",
			wantArgs: []string{"list-sessions", "-F", listSessionsFormat},
			wantPath: "/usr/bin/tmux",
			wantOutput: []*session.Session{
				{Name: "session1", Dir: "/path/to/session1", Exists: true, LastAttached: 1000},
				{Name: "session2", Dir: "/path/to/session2", Exists: true, LastAttached: 2000},
			},
			crOutput: []byte("session1\t/path/to/session1\t1000\nsession2\t/path/to/session2\t2000\n"),
		},
		{
			name:     "session never attached has zero LastAttached",
			wantArgs: []string{"list-sessions", "-F", listSessionsFormat},
			wantPath: "/usr/bin/tmux",
			wantOutput: []*session.Session{
				{Name: "session1", Dir: "/path/to/session1", Exists: true, LastAttached: 0},
			},
			crOutput: []byte("session1\t/path/to/session1\t0\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &TestRunner{output: tt.crOutput}
			client := NewClient(cr, "/usr/bin/tmux")

			sessions := client.AllSessions()

			if len(sessions) != len(tt.wantOutput) {
				t.Fatalf("Expected %d sessions, got %d", len(tt.wantOutput), len(sessions))
			}

			if len(sessions) > 0 {
				for i, sess := range sessions {
					if sess.Name != tt.wantOutput[i].Name || sess.Dir != tt.wantOutput[i].Dir || sess.LastAttached != tt.wantOutput[i].LastAttached {
						t.Fatalf("Expected session %d to be %s:%s:%d, got %s:%s:%d", i, tt.wantOutput[i].Name, tt.wantOutput[i].Dir, tt.wantOutput[i].LastAttached, sess.Name, sess.Dir, sess.LastAttached)
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
