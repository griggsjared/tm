package app

import (
	"errors"
	"testing"

	"github.com/griggsjared/tm/internal/session"
)

type mockTmuxRunner struct {
	newSessionCalled    bool
	attachSessionCalled bool
	newSessionError     error
	attachSessionError  error
	lastSession         *session.Session
}

func (m *mockTmuxRunner) NewSession(s *session.Session) error {
	m.newSessionCalled = true
	m.lastSession = s
	return m.newSessionError
}

func (m *mockTmuxRunner) AttachSession(s *session.Session) error {
	m.attachSessionCalled = true
	m.lastSession = s
	return m.attachSessionError
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
	selectResult  string
	selectOk      bool
	selectError   error
	providedItems []string
	providedQuery string
}

func (m *mockFzfRunner) IsAvailable() bool {
	return m.available
}

func (m *mockFzfRunner) Select(items []string, query string) (string, bool, error) {
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
		newSessionErr    error
		attachSessionErr error
		wantNewCalled    bool
		wantAttachCalled bool
		wantErr          bool
	}{
		{
			name:             "attach to existing session",
			session:          &session.Session{Name: "test", Exists: true},
			wantNewCalled:    false,
			wantAttachCalled: true,
			wantErr:          false,
		},
		{
			name:             "create and attach to new session",
			session:          &session.Session{Name: "test", Exists: false},
			wantNewCalled:    true,
			wantAttachCalled: true,
			wantErr:          false,
		},
		{
			name:             "error creating session",
			session:          &session.Session{Name: "test", Exists: false},
			newSessionErr:    errors.New("create failed"),
			wantNewCalled:    true,
			wantAttachCalled: false,
			wantErr:          true,
		},
		{
			name:             "error attaching to session",
			session:          &session.Session{Name: "test", Exists: true},
			attachSessionErr: errors.New("attach failed"),
			wantNewCalled:    false,
			wantAttachCalled: true,
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmuxMock := &mockTmuxRunner{
				newSessionError:    tt.newSessionErr,
				attachSessionError: tt.attachSessionErr,
			}
			fzfMock := &mockFzfRunner{}
			sessionMock := &mockSessionFinder{}

			app := New(tmuxMock, sessionMock, fzfMock, false)
			err := app.attachToSession(tt.session)

			if (err != nil) != tt.wantErr {
				t.Errorf("attachToSession() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tmuxMock.newSessionCalled != tt.wantNewCalled {
				t.Errorf("NewSession called = %v, want %v", tmuxMock.newSessionCalled, tt.wantNewCalled)
			}
			if tmuxMock.attachSessionCalled != tt.wantAttachCalled {
				t.Errorf("AttachSession called = %v, want %v", tmuxMock.attachSessionCalled, tt.wantAttachCalled)
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
		fzfResult        string
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
			fzfResult:        "session1\t/path/1",
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
			fzfResult:    "",
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
