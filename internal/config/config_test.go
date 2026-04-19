package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	pds := []struct {
		Dir     string
		Name    string
		Aliases []string
	}{
		{Dir: "~/test", Name: "test", Aliases: []string{"t"}},
	}

	sd := []struct {
		Dir string
	}{
		{Dir: "~/projects"},
	}

	cfg := New(true, "/usr/bin/tmux", nil, nil)

	if !cfg.Debug {
		t.Error("expected Debug to be true")
	}
	if cfg.TmuxPath != "/usr/bin/tmux" {
		t.Errorf("expected TmuxPath /usr/bin/tmux, got %s", cfg.TmuxPath)
	}
	if cfg.FzfPath != "" {
		t.Errorf("expected FzfPath empty, got %s", cfg.FzfPath)
	}

	_ = pds
	_ = sd
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
			// Clean up env vars after test
			oldVars := make(map[string]string)
			for k, v := range tt.envVars {
				oldVars[k] = os.Getenv(k)
				os.Setenv(k, v)
			}
			defer func() {
				for k, v := range oldVars {
					if v == "" {
						os.Unsetenv(k)
					} else {
						os.Setenv(k, v)
					}
				}
			}()

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
}

func TestLoad(t *testing.T) {
	t.Run("tmux path validation leaves empty", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		os.WriteFile(configPath, []byte(""), 0644)

		// Set invalid tmux path
		oldTmux := os.Getenv("TM_TMUX_PATH")
		os.Setenv("TM_TMUX_PATH", "/nonexistent/tmux")
		defer os.Setenv("TM_TMUX_PATH", oldTmux)

		oldConfig := os.Getenv("TM_CONFIG_PATH")
		os.Setenv("TM_CONFIG_PATH", configPath)
		defer os.Setenv("TM_CONFIG_PATH", oldConfig)

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
		os.WriteFile(configPath, []byte(""), 0644)

		// Set invalid fzf path
		oldFzf := os.Getenv("TM_FZF_PATH")
		os.Setenv("TM_FZF_PATH", "/nonexistent/fzf")
		defer os.Setenv("TM_FZF_PATH", oldFzf)

		oldConfig := os.Getenv("TM_CONFIG_PATH")
		os.Setenv("TM_CONFIG_PATH", configPath)
		defer os.Setenv("TM_CONFIG_PATH", oldConfig)

		cfg, err := Load()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if cfg.FzfPath != "" {
			t.Errorf("expected empty FzfPath for nonexistent path, got %s", cfg.FzfPath)
		}
	})
}
