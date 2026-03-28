package fzf

import (
	"os/exec"
	"strings"
	"testing"
)

func TestNewRunner(t *testing.T) {
	t.Run("with explicit path", func(t *testing.T) {
		runner := NewRunner("/usr/local/bin/fzf", nil)
		if runner.path != "/usr/local/bin/fzf" {
			t.Errorf("expected path /usr/local/bin/fzf, got %s", runner.path)
		}
	})

	t.Run("with empty path looks up fzf", func(t *testing.T) {
		runner := NewRunner("", nil)
		// If fzf is in PATH, path should be set
		// If not, path should be empty
		if runner.path != "" {
			// Verify it's a valid path
			if _, err := exec.LookPath("fzf"); err != nil {
				t.Errorf("expected empty path when fzf not in PATH, got %s", runner.path)
			}
		}
	})

	t.Run("with custom opts", func(t *testing.T) {
		customOpts := []string{"--height=50%", "--border"}
		runner := NewRunner("/usr/local/bin/fzf", customOpts)
		if len(runner.opts) != 2 {
			t.Errorf("expected 2 opts, got %d", len(runner.opts))
		}
		if runner.opts[0] != "--height=50%" || runner.opts[1] != "--border" {
			t.Errorf("expected custom opts, got %v", runner.opts)
		}
	})

	t.Run("with nil opts uses defaults", func(t *testing.T) {
		runner := NewRunner("/usr/local/bin/fzf", nil)
		if len(runner.opts) != 3 {
			t.Errorf("expected 3 default opts, got %d", len(runner.opts))
		}
		if runner.opts[0] != "--height=20%" || runner.opts[1] != "--ansi" || runner.opts[2] != "--reverse" {
			t.Errorf("expected default opts, got %v", runner.opts)
		}
	})
}

func TestRunner_IsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "available with path",
			path:     "/usr/local/bin/fzf",
			expected: true,
		},
		{
			name:     "not available with empty path",
			path:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &Runner{path: tt.path}
			if got := runner.IsAvailable(); got != tt.expected {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRunner_Select_NotAvailable(t *testing.T) {
	runner := &Runner{path: ""}
	_, _, err := runner.Select([]string{"item"}, "query")
	if err == nil {
		t.Error("expected error when fzf not available")
	}
}

func TestRunner_Select_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func() *Runner
		items       []string
		query       string
		wantErr     bool
		wantOk      bool
		errContains string
	}{
		{
			name: "not available returns error",
			setupMock: func() *Runner {
				return &Runner{path: ""}
			},
			items:       []string{"item"},
			query:       "test",
			wantErr:     true,
			errContains: "not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := tt.setupMock()
			result, ok, err := runner.Select(tt.items, tt.query)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}

			if ok != tt.wantOk {
				t.Errorf("expected ok=%v, got %v", tt.wantOk, ok)
			}

			if !tt.wantErr && !tt.wantOk && result != 0 {
				t.Errorf("expected empty result when not ok, got %d", result)
			}
		})
	}
}
