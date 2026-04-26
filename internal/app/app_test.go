package app

import (
	"errors"
	"testing"

	"github.com/griggsjared/tm/internal/session"
)

type mockTmuxClient struct {
	available           bool
	insideTmux          bool
	path                string
	currentSession      string
	newSessionCalled    bool
	newSessionDetached  bool
	attachSessionCalled bool
	switchSessionCalled bool
	newSessionError     error
	attachSessionError  error
	switchSessionError  error
	lastSession         *session.Session
}

func (m *mockTmuxClient) IsAvailable() bool {
	return m.available
}

func (m *mockTmuxClient) InsideTmux() bool {
	return m.insideTmux
}

func (m *mockTmuxClient) Path() string {
	return m.path
}

func (m *mockTmuxClient) CurrentSession() string {
	return m.currentSession
}

func (m *mockTmuxClient) NewSession(s *session.Session, detached bool) error {
	m.newSessionCalled = true
	m.newSessionDetached = detached
	m.lastSession = s
	return m.newSessionError
}

func (m *mockTmuxClient) AttachSession(s *session.Session) error {
	m.attachSessionCalled = true
	m.lastSession = s
	return m.attachSessionError
}

func (m *mockTmuxClient) SwitchSession(s *session.Session) error {
	m.switchSessionCalled = true
	m.lastSession = s
	return m.switchSessionError
}

type mockSessionFinder struct {
	findCalled           bool
	listCalled           bool
	listExcludingCalled  bool
	listExcludingExclude string
	findResult           *session.Session
	findError            error
	listResult           []*session.Session
}

func (m *mockSessionFinder) Find(name string) (*session.Session, error) {
	m.findCalled = true
	return m.findResult, m.findError
}

func (m *mockSessionFinder) List(onlyActive bool) []*session.Session {
	m.listCalled = true
	return m.listResult
}

func (m *mockSessionFinder) ListExcluding(onlyExisting bool, exclude string) []*session.Session {
	m.listExcludingCalled = true
	m.listExcludingExclude = exclude
	return m.listResult
}

type mockFzfRunner struct {
	available     bool
	path          string
	selectResult  int
	selectOk      bool
	selectError   error
	providedItems []string
	providedQuery string
}

func (m *mockFzfRunner) IsAvailable() bool {
	return m.available
}

func (m *mockFzfRunner) Path() string {
	return m.path
}

func (m *mockFzfRunner) Select(items []string, query string) (int, bool, error) {
	m.providedItems = items
	m.providedQuery = query
	return m.selectResult, m.selectOk, m.selectError
}

func TestFilterSessions(t *testing.T) {
	tests := []struct {
		name      string
		sessions  []*session.Session
		query     string
		wantLen   int
		wantNames []string
	}{
		{
			name:      "empty query returns all",
			sessions:  []*session.Session{{Name: "a"}, {Name: "b"}, {Name: "c"}},
			query:     "",
			wantLen:   3,
			wantNames: []string{"a", "b", "c"},
		},
		{
			name:      "case insensitive match",
			sessions:  []*session.Session{{Name: "Alpha"}, {Name: "beta"}, {Name: "Gamma"}},
			query:     "al",
			wantLen:   1,
			wantNames: []string{"Alpha"},
		},
		{
			name:      "multiple matches",
			sessions:  []*session.Session{{Name: "app1"}, {Name: "app2"}, {Name: "web"}},
			query:     "app",
			wantLen:   2,
			wantNames: []string{"app1", "app2"},
		},
		{
			name:      "no matches",
			sessions:  []*session.Session{{Name: "a"}, {Name: "b"}},
			query:     "xyz",
			wantLen:   0,
			wantNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterSessions(tt.sessions, tt.query)
			if len(got) != tt.wantLen {
				t.Errorf("expected %d sessions, got %d", tt.wantLen, len(got))
			}
			for i, want := range tt.wantNames {
				if i >= len(got) || got[i].Name != want {
					t.Errorf("expected session %d to be %s, got %v", i, want, got)
					break
				}
			}
		})
	}
}

func TestAppAttachToSession(t *testing.T) {
	tests := []struct {
		name             string
		session          *session.Session
		insideTmux       bool
		newSessionErr    error
		attachSessionErr error
		switchSessionErr error
		wantNewCalled    bool
		wantNewDetached  bool
		wantAttachCalled bool
		wantSwitchCalled bool
		wantErr          bool
	}{
		{
			name:             "outside tmux: attach to existing session",
			session:          &session.Session{Name: "test", Exists: true},
			insideTmux:       false,
			wantNewCalled:    false,
			wantAttachCalled: true,
			wantSwitchCalled: false,
			wantErr:          false,
		},
		{
			name:             "outside tmux: create and attach to new session",
			session:          &session.Session{Name: "test", Exists: false},
			insideTmux:       false,
			wantNewCalled:    true,
			wantNewDetached:  false,
			wantAttachCalled: true,
			wantSwitchCalled: false,
			wantErr:          false,
		},
		{
			name:          "outside tmux: error creating session",
			session:       &session.Session{Name: "test", Exists: false},
			insideTmux:    false,
			newSessionErr: errors.New("create failed"),
			wantNewCalled: true,
			wantErr:       true,
		},
		{
			name:             "outside tmux: error attaching to session",
			session:          &session.Session{Name: "test", Exists: true},
			insideTmux:       false,
			attachSessionErr: errors.New("attach failed"),
			wantAttachCalled: true,
			wantErr:          true,
		},
		{
			name:             "inside tmux: switch to existing session",
			session:          &session.Session{Name: "test", Exists: true},
			insideTmux:       true,
			wantSwitchCalled: true,
			wantErr:          false,
		},
		{
			name:             "inside tmux: create and switch to new session",
			session:          &session.Session{Name: "test", Exists: false},
			insideTmux:       true,
			wantNewCalled:    true,
			wantNewDetached:  true,
			wantSwitchCalled: true,
			wantErr:          false,
		},
		{
			name:            "inside tmux: error creating session",
			session:         &session.Session{Name: "test", Exists: false},
			insideTmux:      true,
			newSessionErr:   errors.New("create failed"),
			wantNewCalled:   true,
			wantNewDetached: true,
			wantErr:         true,
		},
		{
			name:             "inside tmux: error switching session",
			session:          &session.Session{Name: "test", Exists: true},
			insideTmux:       true,
			switchSessionErr: errors.New("switch failed"),
			wantSwitchCalled: true,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmuxMock := &mockTmuxClient{
				available:          true,
				insideTmux:         tt.insideTmux,
				newSessionError:    tt.newSessionErr,
				attachSessionError: tt.attachSessionErr,
				switchSessionError: tt.switchSessionErr,
			}
			app := New(tmuxMock, &mockSessionFinder{}, &mockFzfRunner{}, false, "test")
			err := app.attachToSession(tt.session)

			if (err != nil) != tt.wantErr {
				t.Errorf("attachToSession() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tmuxMock.newSessionCalled != tt.wantNewCalled {
				t.Errorf("NewSession called = %v, want %v", tmuxMock.newSessionCalled, tt.wantNewCalled)
			}
			if tmuxMock.newSessionCalled && tmuxMock.newSessionDetached != tt.wantNewDetached {
				t.Errorf("NewSession detached = %v, want %v", tmuxMock.newSessionDetached, tt.wantNewDetached)
			}
			if tmuxMock.attachSessionCalled != tt.wantAttachCalled {
				t.Errorf("AttachSession called = %v, want %v", tmuxMock.attachSessionCalled, tt.wantAttachCalled)
			}
			if tmuxMock.switchSessionCalled != tt.wantSwitchCalled {
				t.Errorf("SwitchSession called = %v, want %v", tmuxMock.switchSessionCalled, tt.wantSwitchCalled)
			}
		})
	}
}
func TestAppSelectSession(t *testing.T) {
	sessions := []*session.Session{
		{Name: "session1", Dir: "/path/1"},
		{Name: "session2", Dir: "/path/2"},
	}

	tests := []struct {
		name             string
		fzfAvailable     bool
		fzfResult        int
		fzfOk            bool
		fzfError         error
		sessions         []*session.Session
		query            string
		wantNil          bool
		wantErr          bool
		wantSelectedName string
	}{
		{
			name:         "no sessions available",
			fzfAvailable: false,
			sessions:     []*session.Session{},
			query:        "",
			wantNil:      true,
			wantErr:      false,
		},
		{
			name:             "fzf selects session",
			fzfAvailable:     true,
			fzfResult:        0,
			fzfOk:            true,
			sessions:         sessions,
			query:            "",
			wantNil:          false,
			wantErr:          false,
			wantSelectedName: "session1",
		},
		{
			name:         "fzf cancelled",
			fzfAvailable: true,
			fzfResult:    0,
			fzfOk:        false,
			sessions:     sessions,
			query:        "",
			wantNil:      true,
			wantErr:      false,
		},
		{
			name:         "fzf not available prints list",
			fzfAvailable: false,
			sessions:     sessions,
			query:        "test",
			wantNil:      true,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmuxMock := &mockTmuxClient{available: true}
			fzfMock := &mockFzfRunner{
				available:    tt.fzfAvailable,
				selectResult: tt.fzfResult,
				selectOk:     tt.fzfOk,
				selectError:  tt.fzfError,
			}
			sessionMock := &mockSessionFinder{}

			app := New(tmuxMock, sessionMock, fzfMock, false, "test")
			got, err := app.selectSession(tt.sessions, tt.query)

			if (err != nil) != tt.wantErr {
				t.Errorf("selectSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantNil {
				if got != nil {
					t.Errorf("selectSession() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Error("selectSession() = nil, want non-nil")
				} else if got.Name != tt.wantSelectedName {
					t.Errorf("selectSession() = %s, want %s", got.Name, tt.wantSelectedName)
				}
			}
		})
	}
}

var _ TmuxClient = &mockTmuxClient{}
var _ SessionFinder = &mockSessionFinder{}
var _ FzfRunner = &mockFzfRunner{}

func TestApp_Run_TmuxUnavailable(t *testing.T) {
	tmuxMock := &mockTmuxClient{available: false}
	sessionMock := &mockSessionFinder{}
	fzfMock := &mockFzfRunner{available: false}

	app := New(tmuxMock, sessionMock, fzfMock, false, "test")

	// Run should return early without calling any session methods
	app.Run("")

	// Verify no session operations were attempted
	if sessionMock.findCalled {
		t.Error("Find should not be called when tmux unavailable")
	}
	if sessionMock.listCalled {
		t.Error("List should not be called when tmux unavailable")
	}
	if tmuxMock.newSessionCalled {
		t.Error("NewSession should not be called when tmux unavailable")
	}
}

func TestApp_Run_InsideTmux_NoArgs(t *testing.T) {
	tmuxMock := &mockTmuxClient{available: true, insideTmux: true, currentSession: "my-session"}
	sessionMock := &mockSessionFinder{listResult: []*session.Session{{Name: "other"}}}
	fzfMock := &mockFzfRunner{available: false}

	app := New(tmuxMock, sessionMock, fzfMock, false, "test")
	app.Run("")

	if !sessionMock.listExcludingCalled {
		t.Error("ListExcluding should be called when inside tmux with no args")
	}
	if sessionMock.listExcludingExclude != "my-session" {
		t.Errorf("ListExcluding exclude = %q, want %q", sessionMock.listExcludingExclude, "my-session")
	}
}

func TestApp_Run_OutsideTmux_NoArgs(t *testing.T) {
	tmuxMock := &mockTmuxClient{available: true}
	sessionMock := &mockSessionFinder{listResult: []*session.Session{{Name: "session1"}}}
	fzfMock := &mockFzfRunner{available: false}

	app := New(tmuxMock, sessionMock, fzfMock, false, "test")
	app.Run("")

	if !sessionMock.listExcludingCalled {
		t.Error("ListExcluding should be called when outside tmux with no args")
	}
	if sessionMock.listExcludingExclude != "" {
		t.Errorf("ListExcluding exclude = %q, want empty string", sessionMock.listExcludingExclude)
	}
}

func TestApp_Run_InsideTmux_ExactMatch(t *testing.T) {
	tmuxMock := &mockTmuxClient{available: true, insideTmux: true, currentSession: "my-session"}
	sessionMock := &mockSessionFinder{
		findResult: &session.Session{Name: "exact", Exists: true},
		listResult: []*session.Session{{Name: "exact", Exists: true}},
	}
	fzfMock := &mockFzfRunner{available: false}

	app := New(tmuxMock, sessionMock, fzfMock, false, "test")
	app.Run("exact")

	if !sessionMock.findCalled {
		t.Error("Find should be called for exact match")
	}
	if sessionMock.listExcludingCalled {
		t.Error("ListExcluding should NOT be called when exact match succeeds")
	}
	if !tmuxMock.switchSessionCalled {
		t.Error("SwitchSession should be called for exact match inside tmux")
	}
}

func TestApp_Run_BuiltinsUseFullList(t *testing.T) {
	tmuxMock := &mockTmuxClient{available: true, insideTmux: true, currentSession: "my-session"}
	sessionMock := &mockSessionFinder{listResult: []*session.Session{{Name: "my-session", Exists: true}}}
	fzfMock := &mockFzfRunner{available: false}

	app := New(tmuxMock, sessionMock, fzfMock, false, "test")
	app.Run("ls")

	if !sessionMock.listCalled {
		t.Error("List should be called for builtin commands")
	}
	if sessionMock.listExcludingCalled {
		t.Error("ListExcluding should NOT be called for builtin commands")
	}
}

func TestApp_Run_InsideTmux_PartialMatch(t *testing.T) {
	tmuxMock := &mockTmuxClient{available: true, insideTmux: true, currentSession: "my-session"}
	sessionMock := &mockSessionFinder{
		listResult: []*session.Session{{Name: "other"}},
	}
	fzfMock := &mockFzfRunner{available: false}

	app := New(tmuxMock, sessionMock, fzfMock, false, "test")
	app.Run("oth")

	if !sessionMock.listExcludingCalled {
		t.Error("ListExcluding should be called when inside tmux with partial match")
	}
	if sessionMock.listExcludingExclude != "my-session" {
		t.Errorf("ListExcluding exclude = %q, want %q", sessionMock.listExcludingExclude, "my-session")
	}
}

func TestApp_Run_Version(t *testing.T) {
	tmuxMock := &mockTmuxClient{available: false}
	sessionMock := &mockSessionFinder{}
	fzfMock := &mockFzfRunner{available: false}

	app := New(tmuxMock, sessionMock, fzfMock, false, "1.2.3")
	got := app.Run("version")

	if got != 0 {
		t.Errorf("Run(\"version\") = %d, want 0", got)
	}
}

func TestApp_Run_Status(t *testing.T) {
	tests := []struct {
		name      string
		tmuxAvail bool
		tmuxPath  string
		fzfAvail  bool
		fzfPath   string
		wantExit  int
	}{
		{
			name:      "both available",
			tmuxAvail: true,
			tmuxPath:  "/usr/bin/tmux",
			fzfAvail:  true,
			fzfPath:   "/usr/bin/fzf",
			wantExit:  0,
		},
		{
			name:      "tmux available, fzf missing",
			tmuxAvail: true,
			tmuxPath:  "/usr/bin/tmux",
			fzfAvail:  false,
			fzfPath:   "",
			wantExit:  0,
		},
		{
			name:      "tmux missing, fzf available",
			tmuxAvail: false,
			tmuxPath:  "",
			fzfAvail:  true,
			fzfPath:   "/usr/bin/fzf",
			wantExit:  1,
		},
		{
			name:      "both missing",
			tmuxAvail: false,
			tmuxPath:  "",
			fzfAvail:  false,
			fzfPath:   "",
			wantExit:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmuxMock := &mockTmuxClient{available: tt.tmuxAvail, path: tt.tmuxPath}
			fzfMock := &mockFzfRunner{available: tt.fzfAvail, path: tt.fzfPath}
			app := New(tmuxMock, &mockSessionFinder{}, fzfMock, false, "test")

			got := app.Run("status")
			if got != tt.wantExit {
				t.Errorf("Run(\"status\") = %v, want %v", got, tt.wantExit)
			}
		})
	}
}

func TestApp_Run_TmuxUnavailable_ReturnsOne(t *testing.T) {
	tmuxMock := &mockTmuxClient{available: false}
	sessionMock := &mockSessionFinder{}
	fzfMock := &mockFzfRunner{available: false}

	app := New(tmuxMock, sessionMock, fzfMock, false, "test")
	got := app.Run("")

	if got != 1 {
		t.Errorf("Run(\"\") = %d, want 1", got)
	}
}

func TestFormatSessionLine(t *testing.T) {
	tests := []struct {
		name     string
		session  *session.Session
		expected string
	}{
		{
			name:     "with directory, not existing",
			session:  &session.Session{Name: "myapp", Dir: "/home/user/myapp", Exists: false},
			expected: "myapp [/home/user/myapp]",
		},
		{
			name:     "with directory, existing",
			session:  &session.Session{Name: "myapp", Dir: "/home/user/myapp", Exists: true},
			expected: "myapp [/home/user/myapp] *",
		},
		{
			name:     "no directory, not existing",
			session:  &session.Session{Name: "myapp", Dir: "", Exists: false},
			expected: "myapp []",
		},
		{
			name:     "no directory, existing",
			session:  &session.Session{Name: "myapp", Dir: "", Exists: true},
			expected: "myapp [] *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSessionLine(tt.session)
			if got != tt.expected {
				t.Errorf("formatSessionLine() = %q, want %q", got, tt.expected)
			}
		})
	}
}
