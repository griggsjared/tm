package fzf

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

type TestRunner struct {
	output         []byte
	exitCode       int
	error          error
	providedPath   string
	providedArgs   []string
	providedStdin  io.Reader
	providedStderr io.Writer
}

func (t *TestRunner) Output(path string, args []string) ([]byte, error) {
	t.providedPath = path
	t.providedArgs = args
	return t.output, t.error
}

func (t *TestRunner) Run(path string, args []string, stdin io.Reader, stderr io.Writer) ([]byte, int, error) {
	t.providedPath = path
	t.providedArgs = args
	t.providedStdin = stdin
	t.providedStderr = stderr
	return t.output, t.exitCode, t.error
}

func TestNewRunner(t *testing.T) {
	runner := NewRunner()
	if runner == nil {
		t.Fatalf("Expected a non-nil FzfRunner")
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
			path:      "/usr/local/bin/fzf",
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
			tr := &TestRunner{}
			client := NewClient(tr, tt.path)

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
			path:     "/usr/local/bin/fzf",
			wantPath: "/usr/local/bin/fzf",
		},
		{
			name:     "empty path",
			path:     "",
			wantPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TestRunner{}
			client := NewClient(tr, tt.path)

			got := client.Path()
			if got != tt.wantPath {
				t.Fatalf("Path() = %v, want %v", got, tt.wantPath)
			}
		})
	}
}

func TestClient_Version(t *testing.T) {
	tests := []struct {
		name       string
		trOutput   []byte
		trError    error
		wantResult string
		wantArgs   []string
		wantPath   string
	}{
		{
			name:       "successful version parse",
			trOutput:   []byte("0.46.1 (add1f66)\n"),
			wantResult: "0.46.1",
			wantArgs:   []string{"--version"},
			wantPath:   "/usr/local/bin/fzf",
		},
		{
			name:       "error returns empty string",
			trError:    errors.New("fzf error"),
			wantResult: "",
			wantArgs:   []string{"--version"},
			wantPath:   "/usr/local/bin/fzf",
		},
		{
			name:       "empty output returns empty string",
			trOutput:   []byte(""),
			wantResult: "",
			wantArgs:   []string{"--version"},
			wantPath:   "/usr/local/bin/fzf",
		},
		{
			name:       "single word output",
			trOutput:   []byte("0.46.1\n"),
			wantResult: "0.46.1",
			wantArgs:   []string{"--version"},
			wantPath:   "/usr/local/bin/fzf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TestRunner{output: tt.trOutput, error: tt.trError}
			client := NewClient(tr, "/usr/local/bin/fzf")

			got := client.Version()
			if got != tt.wantResult {
				t.Fatalf("Version() = %q, want %q", got, tt.wantResult)
			}

			if tr.providedPath != tt.wantPath {
				t.Fatalf("Expected path %s, got %s", tt.wantPath, tr.providedPath)
			}

			for i, wantArg := range tt.wantArgs {
				if len(tr.providedArgs) <= i || tr.providedArgs[i] != wantArg {
					t.Fatalf("Expected arg[%d] %s, got %v", i, wantArg, tr.providedArgs)
				}
			}
		})
	}
}

func TestClient_Select(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		items          []string
		query          string
		trOutput       []byte
		trExitCode     int
		trError        error
		wantIdx        int
		wantOk         bool
		wantErr        bool
		errContains    string
		wantArgs       []string
		wantPath       string
		wantStdinLines []string
	}{
		{
			name:        "not available returns error",
			path:        "",
			items:       []string{"item"},
			query:       "test",
			wantErr:     true,
			errContains: "not available",
		},
		{
			name:           "successful selection",
			path:           "/usr/local/bin/fzf",
			items:          []string{"alpha", "beta", "gamma"},
			query:          "bet",
			trOutput:       []byte("2\tbeta\n"),
			wantIdx:        1,
			wantOk:         true,
			wantArgs:       []string{"--select-1", "--exit-0", "--with-nth=2..", "--query=bet"},
			wantPath:       "/usr/local/bin/fzf",
			wantStdinLines: []string{"1\talpha", "2\tbeta", "3\tgamma"},
		},
		{
			name:       "exit code 1 returns not ok",
			path:       "/usr/local/bin/fzf",
			items:      []string{"item"},
			query:      "test",
			trExitCode: 1,
			trError:    errors.New("exit status 1"),
			wantIdx:    0,
			wantOk:     false,
			wantErr:    false,
			wantArgs:   []string{"--select-1", "--exit-0", "--with-nth=2..", "--query=test"},
			wantPath:   "/usr/local/bin/fzf",
		},
		{
			name:       "exit code 130 returns not ok",
			path:       "/usr/local/bin/fzf",
			items:      []string{"item"},
			query:      "test",
			trExitCode: 130,
			trError:    errors.New("exit status 130"),
			wantIdx:    0,
			wantOk:     false,
			wantErr:    false,
			wantArgs:   []string{"--select-1", "--exit-0", "--with-nth=2..", "--query=test"},
			wantPath:   "/usr/local/bin/fzf",
		},
		{
			name:        "other exit code returns error",
			path:        "/usr/local/bin/fzf",
			items:       []string{"item"},
			query:       "test",
			trExitCode:  2,
			trError:     errors.New("exit status 2"),
			wantErr:     true,
			errContains: "fzf error",
			wantArgs:    []string{"--select-1", "--exit-0", "--with-nth=2..", "--query=test"},
			wantPath:    "/usr/local/bin/fzf",
		},
		{
			name:     "empty output returns not ok",
			path:     "/usr/local/bin/fzf",
			items:    []string{"item"},
			query:    "test",
			trOutput: []byte(""),
			wantIdx:  0,
			wantOk:   false,
			wantErr:  false,
			wantArgs: []string{"--select-1", "--exit-0", "--with-nth=2..", "--query=test"},
			wantPath: "/usr/local/bin/fzf",
		},
		{
			name:        "non-numeric index returns error",
			path:        "/usr/local/bin/fzf",
			items:       []string{"item"},
			query:       "test",
			trOutput:    []byte("abc\titem\n"),
			wantErr:     true,
			errContains: "failed to parse selection index",
			wantArgs:    []string{"--select-1", "--exit-0", "--with-nth=2..", "--query=test"},
			wantPath:    "/usr/local/bin/fzf",
		},
		{
			name:           "zero items still runs",
			path:           "/usr/local/bin/fzf",
			items:          []string{},
			query:          "",
			trOutput:       []byte(""),
			wantIdx:        0,
			wantOk:         false,
			wantErr:        false,
			wantArgs:       []string{"--select-1", "--exit-0", "--with-nth=2..", "--query="},
			wantPath:       "/usr/local/bin/fzf",
			wantStdinLines: []string{},
		},
		{
			name:     "whitespace-only output returns not ok",
			path:     "/usr/local/bin/fzf",
			items:    []string{"item"},
			query:    "test",
			trOutput: []byte("   \n"),
			wantIdx:  0,
			wantOk:   false,
			wantErr:  false,
			wantArgs: []string{"--select-1", "--exit-0", "--with-nth=2..", "--query=test"},
			wantPath: "/usr/local/bin/fzf",
		},
		{
			name:     "output without tab separator parses index",
			path:     "/usr/local/bin/fzf",
			items:    []string{"item"},
			query:    "test",
			trOutput: []byte("42\n"),
			wantIdx:  41,
			wantOk:   true,
			wantErr:  false,
			wantArgs: []string{"--select-1", "--exit-0", "--with-nth=2..", "--query=test"},
			wantPath: "/usr/local/bin/fzf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &TestRunner{output: tt.trOutput, exitCode: tt.trExitCode, error: tt.trError}
			client := NewClient(tr, tt.path)

			idx, ok, err := client.Select(tt.items, tt.query)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if idx != tt.wantIdx {
				t.Fatalf("Select() idx = %d, want %d", idx, tt.wantIdx)
			}
			if ok != tt.wantOk {
				t.Fatalf("Select() ok = %v, want %v", ok, tt.wantOk)
			}

			if tr.providedPath != tt.wantPath {
				t.Fatalf("Expected path %s, got %s", tt.wantPath, tr.providedPath)
			}

			for i, wantArg := range tt.wantArgs {
				if len(tr.providedArgs) <= i || tr.providedArgs[i] != wantArg {
					t.Fatalf("Expected arg[%d] %s, got %v", i, wantArg, tr.providedArgs)
				}
			}

			if tt.wantStdinLines != nil {
				if tr.providedStdin == nil {
					t.Fatalf("Expected stdin to be provided")
				}
				stdinBytes, _ := io.ReadAll(tr.providedStdin)
				lines := strings.Split(strings.TrimSuffix(string(stdinBytes), "\n"), "\n")
				if len(stdinBytes) == 0 {
					lines = []string{}
				}
				if len(lines) != len(tt.wantStdinLines) {
					t.Fatalf("Expected %d stdin lines, got %d: %q", len(tt.wantStdinLines), len(lines), string(stdinBytes))
				}
				for i, wantLine := range tt.wantStdinLines {
					if lines[i] != wantLine {
						t.Fatalf("Expected stdin line %d to be %q, got %q", i, wantLine, lines[i])
					}
				}
			}
		})
	}
}

func TestFzfRunner_Run(t *testing.T) {
	runner := NewRunner()

	// Test that Run correctly pipes stdin and captures stdout with exit code
	output, exitCode, err := runner.Run("cat", []string{}, bytes.NewReader([]byte("hello\n")), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if string(output) != "hello\n" {
		t.Fatalf("expected output %q, got %q", "hello\n", string(output))
	}
}

func TestFzfRunner_Run_NonZeroExit(t *testing.T) {
	runner := NewRunner()

	// Use a command that exits with code 1 but also writes to stdout
	output, exitCode, err := runner.Run("sh", []string{"-c", "echo hello; exit 1"}, nil, nil)
	if err == nil {
		t.Fatal("expected error for non-zero exit")
	}
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if string(output) != "hello\n" {
		t.Fatalf("expected output %q, got %q", "hello\n", string(output))
	}
}

func TestFzfRunner_Run_CommandNotFound(t *testing.T) {
	runner := NewRunner()

	output, exitCode, err := runner.Run("/nonexistent/binary", []string{}, nil, nil)
	if err == nil {
		t.Fatal("expected error for command not found")
	}
	if exitCode != -1 {
		t.Fatalf("expected exit code -1, got %d", exitCode)
	}
	if len(output) != 0 {
		t.Fatalf("expected empty output, got %q", string(output))
	}
}

func TestFzfRunner_Run_StderrPassthrough(t *testing.T) {
	runner := NewRunner()

	var stderr bytes.Buffer
	output, exitCode, err := runner.Run("sh", []string{"-c", "echo hello; echo error >&2; exit 0"}, nil, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", exitCode)
	}
	if string(output) != "hello\n" {
		t.Fatalf("expected stdout %q, got %q", "hello\n", string(output))
	}
	if !strings.Contains(stderr.String(), "error") {
		t.Fatalf("expected stderr to contain %q, got %q", "error", stderr.String())
	}
}
