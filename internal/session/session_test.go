package session

import (
	"os"
	"path/filepath"
	"testing"
)

type mockRepository struct {
	hasSession  bool
	allSessions []*Session
}

func (m *mockRepository) HasSession(name string) bool {
	return m.hasSession
}

func (m *mockRepository) AllSessions() []*Session {
	return m.allSessions
}

func TestFindExistingSession(t *testing.T) {
	t.Run("existing session found", func(t *testing.T) {
		checker := &mockRepository{hasSession: true}
		svc := NewService(checker, nil, nil)

		sess, err := svc.findExistingSession("test")
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
		checker := &mockRepository{hasSession: false}
		svc := NewService(checker, nil, nil)

		sess, err := svc.findExistingSession("test")
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
			svc := NewService(&mockRepository{}, tt.pre, nil)
			sess, err := svc.findPreDefinedSession(tt.lookup)
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
			svc := NewService(&mockRepository{}, nil, tt.smart)
			sess, err := svc.findSmartSessionDirectorySession(tt.lookup)
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
		checker    Repository
		pre        []PreDefinedSession
		smart      []SmartDirectory
		lookup     string
		wantExists bool
		wantDir    string
	}{
		{
			name:       "existing session",
			checker:    &mockRepository{hasSession: true},
			lookup:     "exists",
			wantExists: true,
		},
		{
			name:    "predefined session",
			checker: &mockRepository{hasSession: false},
			pre: []PreDefinedSession{
				{Name: "predefined", Dir: projectDir},
			},
			lookup:  "predefined",
			wantDir: projectDir,
		},
		{
			name:    "predefined session that is already running and matches on alias",
			checker: &mockRepository{hasSession: true},
			pre: []PreDefinedSession{
				{Name: "running", Dir: projectDir, Aliases: []string{"run"}},
			},
			lookup:     "run",
			wantExists: true,
		},
		{
			name:    "smart directory",
			checker: &mockRepository{hasSession: false},
			smart:   []SmartDirectory{{Dir: tmp}},
			lookup:  "myproject",
			wantDir: projectDir,
		},
		{
			name:    "fallback to cwd",
			checker: &mockRepository{hasSession: false},
			lookup:  "fallback",
			wantDir: os.Getenv("PWD"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.checker, tt.pre, tt.smart)
			sess, err := svc.Find(tt.lookup)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
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
