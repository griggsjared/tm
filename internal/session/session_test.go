package session

import (
	"os"
	"path/filepath"
	"testing"
)

type mockTmuxRepository struct {
	hasSession  bool
	allSessions []*Session
}

func (m *mockTmuxRepository) HasSession(name string) bool {
	return m.hasSession
}

func (m *mockTmuxRepository) AllSessions() []*Session {
	return m.allSessions
}

func TestFindExistingSession(t *testing.T) {
	t.Run("existing session found", func(t *testing.T) {
		checker := &mockTmuxRepository{hasSession: true}
		finder := NewFinder(checker, nil, nil)

		sess, err := finder.findExistingSession("test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sess == nil {
			t.Fatal("expected session, got nil")
		}
		if sess.Name != "test" {
			t.Errorf("expected name test, got %s", sess.Name)
		}
		if !sess.Exists {
			t.Error("expected session to exist")
		}
	})

	t.Run("no existing session", func(t *testing.T) {
		checker := &mockTmuxRepository{hasSession: false}
		finder := NewFinder(checker, nil, nil)

		sess, err := finder.findExistingSession("test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sess != nil {
			t.Errorf("expected nil, got %v", sess)
		}
	})
}

func TestFindPreDefinedSession(t *testing.T) {
	tmp := t.TempDir()
	validDir := filepath.Join(tmp, "valid")
	os.Mkdir(validDir, 0755)

	tests := []struct {
		name     string
		pre      []PreDefinedSession
		lookup   string
		wantName string
		wantDir  string
		wantNil  bool
	}{
		{
			name: "match by name",
			pre: []PreDefinedSession{
				{Name: "myapp", Dir: validDir},
			},
			lookup:   "myapp",
			wantName: "myapp",
			wantDir:  validDir,
		},
		{
			name: "match by alias",
			pre: []PreDefinedSession{
				{Name: "myapp", Dir: validDir, Aliases: []string{"ma"}},
			},
			lookup:   "ma",
			wantName: "myapp",
			wantDir:  validDir,
		},
		{
			name: "no match",
			pre: []PreDefinedSession{
				{Name: "myapp", Dir: validDir},
			},
			lookup:  "other",
			wantNil: true,
		},
		{
			name: "directory does not exist",
			pre: []PreDefinedSession{
				{Name: "myapp", Dir: "/nonexistent"},
			},
			lookup:  "myapp",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := NewFinder(&mockTmuxRepository{}, tt.pre, nil)
			sess, err := finder.findPreDefinedSession(tt.lookup)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if sess != nil {
					t.Errorf("expected nil, got %v", sess)
				}
				return
			}
			if sess == nil {
				t.Fatal("expected session, got nil")
			}
			if sess.Name != tt.wantName {
				t.Errorf("expected name %s, got %s", tt.wantName, sess.Name)
			}
			if sess.Dir != tt.wantDir {
				t.Errorf("expected dir %s, got %s", tt.wantDir, sess.Dir)
			}
		})
	}
}

func TestNameToMatch(t *testing.T) {
	pds := PreDefinedSession{
		Name:    "main",
		Aliases: []string{"m", "ma"},
	}
	got := nameToMatch(pds)
	want := []string{"main", "m", "ma"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("at index %d: expected %s, got %s", i, want[i], got[i])
		}
	}
}

func TestFindSmartSessionDirectories(t *testing.T) {
	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "myproject")
	os.Mkdir(projectDir, 0755)

	tests := []struct {
		name    string
		smart   []SmartDirectory
		lookup  string
		wantDir string
		wantNil bool
	}{
		{
			name:    "project found",
			smart:   []SmartDirectory{{Dir: tmp}},
			lookup:  "myproject",
			wantDir: projectDir,
		},
		{
			name:    "project not found",
			smart:   []SmartDirectory{{Dir: tmp}},
			lookup:  "nonexistent",
			wantNil: true,
		},
		{
			name:    "multiple smart dirs",
			smart:   []SmartDirectory{{Dir: "/nonexistent"}, {Dir: tmp}},
			lookup:  "myproject",
			wantDir: projectDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := NewFinder(&mockTmuxRepository{}, nil, tt.smart)
			sess, err := finder.findSmartSessionDirectorySession(tt.lookup)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if sess != nil {
					t.Errorf("expected nil, got %v", sess)
				}
				return
			}
			if sess == nil {
				t.Fatal("expected session, got nil")
			}
			if sess.Name != tt.lookup {
				t.Errorf("expected name %s, got %s", tt.lookup, sess.Name)
			}
			if sess.Dir != tt.wantDir {
				t.Errorf("expected dir %s, got %s", tt.wantDir, sess.Dir)
			}
		})
	}
}

func TestServiceFind(t *testing.T) {
	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "myproject")
	os.Mkdir(projectDir, 0755)

	tests := []struct {
		name       string
		checker    TmuxRepository
		pre        []PreDefinedSession
		smart      []SmartDirectory
		lookup     string
		wantExists bool
		wantDir    string
		wantNil    bool
	}{
		{
			name:       "existing session",
			checker:    &mockTmuxRepository{hasSession: true},
			lookup:     "exists",
			wantExists: true,
		},
		{
			name:    "predefined session",
			checker: &mockTmuxRepository{hasSession: false},
			pre: []PreDefinedSession{
				{Name: "predefined", Dir: projectDir},
			},
			lookup:  "predefined",
			wantDir: projectDir,
		},
		{
			name:    "predefined session that is already running and matches on alias",
			checker: &mockTmuxRepository{hasSession: true},
			pre: []PreDefinedSession{
				{Name: "running", Dir: projectDir, Aliases: []string{"run"}},
			},
			lookup:     "run",
			wantExists: true,
		},
		{
			name:    "smart directory",
			checker: &mockTmuxRepository{hasSession: false},
			smart:   []SmartDirectory{{Dir: tmp}},
			lookup:  "myproject",
			wantDir: projectDir,
		},
		{
			name:    "no match returns nil",
			checker: &mockTmuxRepository{hasSession: false},
			lookup:  "nonexistent",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := NewFinder(tt.checker, tt.pre, tt.smart)
			sess, err := finder.Find(tt.lookup)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if sess != nil {
					t.Errorf("expected nil, got %v", sess)
				}
				return
			}
			if sess.Exists != tt.wantExists {
				t.Errorf("expected exists=%v, got %v", tt.wantExists, sess.Exists)
			}
			if tt.wantDir != "" && sess.Dir != tt.wantDir {
				t.Errorf("expected dir %s, got %s", tt.wantDir, sess.Dir)
			}
		})
	}
}

func TestListExcluding(t *testing.T) {
	tmp := t.TempDir()
	predefinedDir := filepath.Join(tmp, "predefined")
	os.Mkdir(predefinedDir, 0755)

	existing := []*Session{
		{Name: "session-a", Dir: "/tmp/a", Exists: true, LastAttached: 3000},
		{Name: "session-b", Dir: "/tmp/b", Exists: true, LastAttached: 2000},
		{Name: "session-c", Dir: "/tmp/c", Exists: true, LastAttached: 1000},
	}

	pre := []PreDefinedSession{
		{Name: "predefined", Dir: predefinedDir, Aliases: []string{"pd", "pre"}},
		{Name: "other", Dir: predefinedDir},
	}

	finder := NewFinder(&mockTmuxRepository{allSessions: existing}, pre, nil)

	tests := []struct {
		name         string
		exclude      string
		wantNames    []string
		wantNotFound []string
	}{
		{
			name:      "empty exclude returns full list",
			exclude:   "",
			wantNames: []string{"session-a", "session-b", "session-c", "other", "pd", "pre", "predefined"},
		},
		{
			name:         "exclude non-predefined removes only that name",
			exclude:      "session-b",
			wantNames:    []string{"session-a", "session-c", "other", "pd", "pre", "predefined"},
			wantNotFound: []string{"session-b"},
		},
		{
			name:         "exclude predefined canonical removes canonical and aliases",
			exclude:      "predefined",
			wantNames:    []string{"session-a", "session-b", "session-c", "other"},
			wantNotFound: []string{"predefined", "pd", "pre"},
		},
		{
			name:         "exclude predefined alias removes canonical and aliases",
			exclude:      "pd",
			wantNames:    []string{"session-a", "session-b", "session-c", "other"},
			wantNotFound: []string{"predefined", "pd", "pre"},
		},
		{
			name:         "only matching predefined session is excluded",
			exclude:      "other",
			wantNames:    []string{"session-a", "session-b", "session-c", "pd", "pre", "predefined"},
			wantNotFound: []string{"other"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := finder.ListExcluding(false, tt.exclude)

			if len(got) != len(tt.wantNames) {
				t.Fatalf("expected %d sessions, got %d", len(tt.wantNames), len(got))
			}

			for i, want := range tt.wantNames {
				if got[i].Name != want {
					t.Errorf("position %d: expected %s, got %s", i, want, got[i].Name)
				}
			}

			for _, notWant := range tt.wantNotFound {
				for _, s := range got {
					if s.Name == notWant {
						t.Errorf("expected %s to be excluded, but it was found", notWant)
					}
				}
			}
		})
	}
}

func TestList_SortedByLastAttached(t *testing.T) {
	sessions := []*Session{
		{Name: "oldest", Dir: "/tmp/oldest", Exists: true, LastAttached: 1000},
		{Name: "newest", Dir: "/tmp/newest", Exists: true, LastAttached: 3000},
		{Name: "never", Dir: "/tmp/never", Exists: true, LastAttached: 0},
		{Name: "middle", Dir: "/tmp/middle", Exists: true, LastAttached: 2000},
	}

	finder := NewFinder(&mockTmuxRepository{allSessions: sessions}, nil, nil)
	got := finder.List(true)

	wantOrder := []string{"newest", "middle", "oldest", "never"}
	if len(got) != len(wantOrder) {
		t.Fatalf("expected %d sessions, got %d", len(wantOrder), len(got))
	}
	for i, name := range wantOrder {
		if got[i].Name != name {
			t.Errorf("position %d: expected %s, got %s", i, name, got[i].Name)
		}
	}
}

func TestList_MixedSessionsOrdering(t *testing.T) {
	preTmp := t.TempDir()
	predefinedDirM := filepath.Join(preTmp, "predefined-m")
	os.Mkdir(predefinedDirM, 0755)
	predefinedDirA := filepath.Join(preTmp, "predefined-a")
	os.Mkdir(predefinedDirA, 0755)

	smartTmp := t.TempDir()
	os.Mkdir(filepath.Join(smartTmp, "smart-b"), 0755)
	os.Mkdir(filepath.Join(smartTmp, "smart-z"), 0755)
	os.Mkdir(filepath.Join(smartTmp, "smart-c"), 0755)

	existing := []*Session{
		{Name: "existing-oldest", Dir: "/tmp/existing-oldest", Exists: true, LastAttached: 1000},
		{Name: "existing-newest", Dir: "/tmp/existing-newest", Exists: true, LastAttached: 3000},
	}

	pre := []PreDefinedSession{
		{Name: "predefined-m", Dir: predefinedDirM},
		{Name: "predefined-a", Dir: predefinedDirA},
	}

	smart := []SmartDirectory{{Dir: smartTmp}}

	finder := NewFinder(&mockTmuxRepository{allSessions: existing}, pre, smart)
	got := finder.List(false)

	wantOrder := []string{"existing-newest", "existing-oldest", "predefined-a", "predefined-m", "smart-b", "smart-c", "smart-z"}
	if len(got) != len(wantOrder) {
		t.Fatalf("expected %d sessions, got %d", len(wantOrder), len(got))
	}
	for i, name := range wantOrder {
		if got[i].Name != name {
			t.Errorf("position %d: expected %s, got %s", i, name, got[i].Name)
		}
	}
}

func TestList_PredefinedSmartNameCollision(t *testing.T) {
	preTmp := t.TempDir()
	predefinedDir := filepath.Join(preTmp, "collision")
	os.Mkdir(predefinedDir, 0755)

	smartTmp := t.TempDir()
	os.Mkdir(filepath.Join(smartTmp, "collision"), 0755)

	pre := []PreDefinedSession{
		{Name: "collision", Dir: predefinedDir},
	}

	smart := []SmartDirectory{{Dir: smartTmp}}

	finder := NewFinder(&mockTmuxRepository{}, pre, smart)
	got := finder.List(false)

	// Both predefined and smart sessions with the same name appear since
	// deduplication only checks against existing tmux sessions.
	if len(got) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(got))
	}
	for _, s := range got {
		if s.Name != "collision" {
			t.Errorf("expected name collision, got %s", s.Name)
		}
	}
}

func TestDirExists(t *testing.T) {
	tmp := t.TempDir()
	exists := dirExists(tmp)
	if !exists {
		t.Error("expected temp dir to exist")
	}
	exists = dirExists("/nonexistent/path")
	if exists {
		t.Error("expected nonexistent path to not exist")
	}
}

func TestExpandHomeDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		input string
		want  string
	}{
		{"~/test", home + "/test"},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := expandHomeDir(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("expected %s, got %s", tt.want, got)
			}
		})
	}
}
