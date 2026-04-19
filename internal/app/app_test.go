package app

import (
	"errors"
	"testing"

	"github.com/griggsjared/tm/internal/session"
)

type mockTmuxRunner struct {
	newSessionCalled    bool
	newSessionDetached  bool
	attachSessionCalled bool
	switchSessionCalled bool
	newSessionError     error
	attachSessionError  error
	switchSessionError  error
	lastSession         *session.Session
}

func (m *mockTmuxRunner) NewSession(s *session.Session, detached bool) error {
	m.newSessionCalled = true
	m.newSessionDetached = detached
	m.lastSession = s
	return m.newSessionError
}

func (m *mockTmuxRunner) AttachSession(s *session.Session) error {
	m.attachSessionCalled = true
	m.lastSession = s
	return m.attachSessionError
}

func (m *mockTmuxRunner) SwitchSession(s *session.Session) error {
	m.switchSessionCalled = true
	m.lastSession = s
	return m.switchSessionError
}

type mockSessionFinder struct {
	findCalled bool
	listCalled bool
	findResult *session.Session
	findError  error
	listResult []*session.Session
}

func (m *mockSessionFinder) Find(name string) (*session.Session, error) {
	m.findCalled = true
	return m.findResult, m.findError
}

func (m *mockSessionFinder) List(onlyActive bool) []*session.Session {
	m.listCalled = true
	return m.listResult
}

type mockFzfRunner struct {
	available     bool
	selectResult  int
	selectOk      bool
	selectError   error
	providedItems []string
	providedQuery string
}

func (m *mockFzfRunner) IsAvailable() bool {
	return m.available
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
			if tt.insideTmux {
				t.Setenv("TMUX", "/tmp/tmux-1234/default,0,0")
			} else {
				t.Setenv("TMUX", "")
			}

			tmuxMock := &mockTmuxRunner{
				newSessionError:    tt.newSessionErr,
				attachSessionError: tt.attachSessionErr,
				switchSessionError: tt.switchSessionErr,
			}
			app := New(tmuxMock, &mockSessionFinder{}, &mockFzfRunner{}, false)
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
			tmuxMock := &mockTmuxRunner{}
			fzfMock := &mockFzfRunner{
				available:    tt.fzfAvailable,
				selectResult: tt.fzfResult,
				selectOk:     tt.fzfOk,
				selectError:  tt.fzfError,
			}
			sessionMock := &mockSessionFinder{}

			app := New(tmuxMock, sessionMock, fzfMock, false)
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

var _ TmuxRunner = &mockTmuxRunner{}
var _ SessionFinder = &mockSessionFinder{}
var _ FzfRunner = &mockFzfRunner{}

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
