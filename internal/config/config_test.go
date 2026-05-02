package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/griggsjared/tm/internal/session"
)

func TestNew(t *testing.T) {
	pds := []session.PreDefinedSession{
		{Dir: "~/test", Name: "test", Aliases: []string{"t"}},
	}

	sd := []session.SmartDirectory{
		{Dir: "~/projects"},
	}

	cfg := New(true, "/usr/bin/tmux", "/usr/bin/fzf", pds, sd)

	if !cfg.Debug {
		t.Error("expected Debug to be true")
	}
	if cfg.TmuxPath != "/usr/bin/tmux" {
		t.Errorf("expected TmuxPath /usr/bin/tmux, got %s", cfg.TmuxPath)
	}
	if cfg.FzfPath != "/usr/bin/fzf" {
		t.Errorf("expected FzfPath /usr/bin/fzf, got %s", cfg.FzfPath)
	}
	if len(cfg.PreDefinedSessions) != 1 || cfg.PreDefinedSessions[0].Name != "test" {
		t.Errorf("expected PreDefinedSessions, got %v", cfg.PreDefinedSessions)
	}
	if len(cfg.SmartDirectories) != 1 || cfg.SmartDirectories[0].Dir != "~/projects" {
		t.Errorf("expected SmartDirectories, got %v", cfg.SmartDirectories)
	}
}

func TestDefaultConfigPath(t *testing.T) {
	home := os.Getenv("HOME")
	if home == "" {
		t.Skip("HOME not set, skipping test")
	}

	path, err := defaultConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(home, ".config", "tm", "config.yaml")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	tests := []struct {
		name       string
		envVars    map[string]string
		wantDebug  bool
		wantTmux   string
		wantFzf    string
		wantConfig string
	}{
		{
			name:       "default values",
			envVars:    map[string]string{},
			wantDebug:  false,
			wantTmux:   "",
			wantFzf:    "",
			wantConfig: "",
		},
		{
			name: "all env vars set",
			envVars: map[string]string{
				"TM_DEBUG":       "true",
				"TM_TMUX_PATH":   "/usr/bin/tmux",
				"TM_FZF_PATH":    "/usr/bin/fzf",
				"TM_CONFIG_PATH": "/custom/config.yaml",
			},
			wantDebug:  true,
			wantTmux:   "/usr/bin/tmux",
			wantFzf:    "/usr/bin/fzf",
			wantConfig: "/custom/config.yaml",
		},
		{
			name: "only fzf path set",
			envVars: map[string]string{
				"TM_FZF_PATH": "/opt/fzf/bin/fzf",
			},
			wantDebug:  false,
			wantTmux:   "",
			wantFzf:    "/opt/fzf/bin/fzf",
			wantConfig: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range []string{"TM_DEBUG", "TM_TMUX_PATH", "TM_FZF_PATH", "TM_CONFIG_PATH"} {
				t.Setenv(k, "")
			}
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			cfg, err := loadConfigFromEnv()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.Debug != tt.wantDebug {
				t.Errorf("Debug = %v, want %v", cfg.Debug, tt.wantDebug)
			}
			if cfg.TmuxPath != tt.wantTmux {
				t.Errorf("TmuxPath = %s, want %s", cfg.TmuxPath, tt.wantTmux)
			}
			if cfg.FzfPath != tt.wantFzf {
				t.Errorf("FzfPath = %s, want %s", cfg.FzfPath, tt.wantFzf)
			}
			if cfg.ConfigPath != tt.wantConfig {
				t.Errorf("ConfigPath = %s, want %s", cfg.ConfigPath, tt.wantConfig)
			}
		})
	}
}

func TestLoadConfigFromConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		content   string
		wantErr   bool
		wantPds   int
		wantSmart int
	}{
		{
			name:      "empty file",
			content:   "",
			wantErr:   false,
			wantPds:   0,
			wantSmart: 0,
		},
		{
			name: "valid config",
			content: `
sessions:
  - dir: ~/projects/app1
    name: app1
    aliases:
      - a1
  - dir: ~/projects/app2
    name: app2
smart_directories:
  - ~/projects
  - ~/work
`,
			wantErr:   false,
			wantPds:   2,
			wantSmart: 2,
		},
		{
			name: "invalid yaml",
			content: `
sessions: [
  invalid yaml here
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tmpDir, tt.name+".yaml")
			err := os.WriteFile(configPath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			cfg, err := loadConfigFromConfigFile(configPath, configPath)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(cfg.PreDefinedSessions) != tt.wantPds {
				t.Errorf("expected %d predefined sessions, got %d", tt.wantPds, len(cfg.PreDefinedSessions))
			}
			if len(cfg.SmartDirectories) != tt.wantSmart {
				t.Errorf("expected %d smart directories, got %d", tt.wantSmart, len(cfg.SmartDirectories))
			}
		})
	}

	t.Run("missing custom path", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "does-not-exist.yaml")
		_, err := loadConfigFromConfigFile(path, filepath.Join(tmpDir, "default.yaml"))
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "config file does not exist") {
			t.Errorf("expected 'config file does not exist' in error, got: %v", err)
		}
	})

	t.Run("missing default path creates empty config", func(t *testing.T) {
		tmpDir := t.TempDir()
		defaultPath := filepath.Join(tmpDir, "config.yaml")
		cfg, err := loadConfigFromConfigFile(defaultPath, defaultPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cfg.PreDefinedSessions) != 0 || len(cfg.SmartDirectories) != 0 {
			t.Errorf("expected empty config, got %+v", cfg)
		}
		if _, err := os.Stat(defaultPath); err != nil {
			t.Errorf("expected default config file to be created: %v", err)
		}
	})

	t.Run("permission denied", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("running as root, permission checks are ineffective")
		}
		tmpDir := t.TempDir()
		subDir := filepath.Join(tmpDir, "noperm")
		if err := os.Mkdir(subDir, 0755); err != nil {
			t.Fatal(err)
		}
		configPath := filepath.Join(subDir, "config.yaml")
		if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(subDir, 0000); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(subDir, 0755)

		_, err := loadConfigFromConfigFile(configPath, filepath.Join(tmpDir, "default.yaml"))
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "config file is inaccessible") {
			t.Errorf("expected 'config file is inaccessible' in error, got: %v", err)
		}
	})
}

func TestLoad(t *testing.T) {
	t.Run("tmux path validation leaves empty", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		t.Setenv("TM_TMUX_PATH", "/nonexistent/tmux")
		t.Setenv("TM_CONFIG_PATH", configPath)

		cfg, err := Load()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg.TmuxPath != "" {
			t.Errorf("expected empty TmuxPath for nonexistent path, got %s", cfg.TmuxPath)
		}
	})

	t.Run("fzf path validation leaves empty", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		t.Setenv("TM_FZF_PATH", "/nonexistent/fzf")
		t.Setenv("TM_CONFIG_PATH", configPath)

		cfg, err := Load()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg.FzfPath != "" {
			t.Errorf("expected empty FzfPath for nonexistent path, got %s", cfg.FzfPath)
		}
	})
}
