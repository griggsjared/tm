package main

import (
	"os"
	"path/filepath"
	"testing"
)

type mockChecker struct {
	hasSession bool
}

func (m *mockChecker) HasSession(name string) bool {
	return m.hasSession
}

func TestFindExistingSession(t *testing.T) {
	t.Run("existing session found", func(t *testing.T) {
		checker := &mockChecker{hasSession: true}
		finder := NewSessionFinder(checker, nil, nil)

		session, err := finder.findExistingSession("test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if session == nil {
			t.Fatal("expected session, got nil")
		}
		if session.name != "test" {
			t.Errorf("expected name test, got %s", session.name)
		}
		if !session.exists {
			t.Error("expected session to exist")
		}
	})

	t.Run("no existing session", func(t *testing.T) {
		checker := &mockChecker{hasSession: false}
		finder := NewSessionFinder(checker, nil, nil)

		session, err := finder.findExistingSession("test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if session != nil {
			t.Errorf("expected nil, got %v", session)
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
				{name: "myapp", dir: validDir},
			},
			lookup:   "myapp",
			wantName: "myapp",
			wantDir:  validDir,
		},
		{
			name: "match by alias",
			pre: []PreDefinedSession{
				{name: "myapp", dir: validDir, aliases: []string{"ma"}},
			},
			lookup:   "ma",
			wantName: "myapp",
			wantDir:  validDir,
		},
		{
			name: "no match",
			pre: []PreDefinedSession{
				{name: "myapp", dir: validDir},
			},
			lookup:  "other",
			wantNil: true,
		},
		{
			name: "directory does not exist",
			pre: []PreDefinedSession{
				{name: "myapp", dir: "/nonexistent"},
			},
			lookup:  "myapp",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := NewSessionFinder(&mockChecker{}, tt.pre, nil)
			session, err := finder.findPreDefinedSession(tt.lookup)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if session != nil {
					t.Errorf("expected nil, got %v", session)
				}
				return
			}
			if session == nil {
				t.Fatal("expected session, got nil")
			}
			if session.name != tt.wantName {
				t.Errorf("expected name %s, got %s", tt.wantName, session.name)
			}
			if session.dir != tt.wantDir {
				t.Errorf("expected dir %s, got %s", tt.wantDir, session.dir)
			}
		})
	}
}

func TestNameToMatch(t *testing.T) {
	pds := PreDefinedSession{
		name:    "main",
		aliases: []string{"m", "ma"},
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
			smart:   []SmartDirectory{{dir: tmp}},
			lookup:  "myproject",
			wantDir: projectDir,
		},
		{
			name:    "project not found",
			smart:   []SmartDirectory{{dir: tmp}},
			lookup:  "nonexistent",
			wantNil: true,
		},
		{
			name:    "multiple smart dirs",
			smart:   []SmartDirectory{{dir: "/nonexistent"}, {dir: tmp}},
			lookup:  "myproject",
			wantDir: projectDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := NewSessionFinder(&mockChecker{}, nil, tt.smart)
			session, err := finder.findSmartSessionDirectories(tt.lookup)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if session != nil {
					t.Errorf("expected nil, got %v", session)
				}
				return
			}
			if session == nil {
				t.Fatal("expected session, got nil")
			}
			if session.name != tt.lookup {
				t.Errorf("expected name %s, got %s", tt.lookup, session.name)
			}
			if session.dir != tt.wantDir {
				t.Errorf("expected dir %s, got %s", tt.wantDir, session.dir)
			}
		})
	}
}

func TestSessionFinderFind(t *testing.T) {
	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "myproject")
	os.Mkdir(projectDir, 0755)

	tests := []struct {
		name       string
		checker    SessionChecker
		pre        []PreDefinedSession
		smart      []SmartDirectory
		lookup     string
		wantExists bool
		wantDir    string
	}{
		{
			name:       "existing session",
			checker:    &mockChecker{hasSession: true},
			lookup:     "exists",
			wantExists: true,
		},
		{
			name:    "predefined session",
			checker: &mockChecker{hasSession: false},
			pre: []PreDefinedSession{
				{name: "predefined", dir: projectDir},
			},
			lookup:  "predefined",
			wantDir: projectDir,
		},
		{
			name:    "smart directory",
			checker: &mockChecker{hasSession: false},
			smart:   []SmartDirectory{{dir: tmp}},
			lookup:  "myproject",
			wantDir: projectDir,
		},
		{
			name:    "fallback to cwd",
			checker: &mockChecker{hasSession: false},
			lookup:  "fallback",
			wantDir: os.Getenv("PWD"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := NewSessionFinder(tt.checker, tt.pre, tt.smart)
			session, err := finder.Find(tt.lookup)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if session.exists != tt.wantExists {
				t.Errorf("expected exists=%v, got %v", tt.wantExists, session.exists)
			}
			if tt.wantDir != "" && session.dir != tt.wantDir {
				t.Errorf("expected dir %s, got %s", tt.wantDir, session.dir)
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
