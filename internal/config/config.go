package config

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sethvargo/go-envconfig"
	yaml "gopkg.in/yaml.v3"

	"github.com/griggsjared/tm/internal/session"
)

type Config struct {
	Debug              bool
	TmuxPath           string
	FzfPath            string
	PreDefinedSessions []session.PreDefinedSession
	SmartDirectories   []session.SmartDirectory
}

func New(debug bool, tmuxPath string, preDefinedSessions []session.PreDefinedSession, smartDirectories []session.SmartDirectory) *Config {
	return &Config{
		Debug:              debug,
		TmuxPath:           tmuxPath,
		PreDefinedSessions: preDefinedSessions,
		SmartDirectories:   smartDirectories,
	}
}

func Load() (*Config, error) {
	config := New(false, "", nil, nil)

	envConfig, err := loadConfigFromEnv()
	if err != nil {
		return nil, err
	}

	defaultConfigPath, err := defaultConfigPath()
	if err != nil {
		return nil, err
	}

	configPath := envConfig.ConfigPath
	if configPath == "" {
		configPath = defaultConfigPath
	}

	fileConfig, err := loadConfigFromConfigFile(configPath, defaultConfigPath)
	if err != nil {
		return nil, err
	}

	config.PreDefinedSessions = make([]session.PreDefinedSession, len(fileConfig.PreDefinedSessions))
	for i, pd := range fileConfig.PreDefinedSessions {
		config.PreDefinedSessions[i] = session.PreDefinedSession{
			Dir:     pd.Dir,
			Name:    pd.Name,
			Aliases: pd.Aliases,
		}
	}

	config.SmartDirectories = make([]session.SmartDirectory, len(fileConfig.SmartDirectories))
	for i, sd := range fileConfig.SmartDirectories {
		config.SmartDirectories[i] = session.SmartDirectory{
			Dir: sd,
		}
	}

	config.Debug = envConfig.Debug
	config.TmuxPath = envConfig.TmuxPath

	if config.TmuxPath != "" {
		if _, err := os.Stat(config.TmuxPath); err == nil {
			// path exists, keep it
		} else {
			// path doesn't exist, leave empty (caller will check)
			config.TmuxPath = ""
		}
	} else {
		if path, err := exec.LookPath("tmux"); err == nil {
			config.TmuxPath = path
		}
		// not found, leave empty (caller will check)
	}

	config.FzfPath = envConfig.FzfPath

	if config.FzfPath != "" {
		if _, err := os.Stat(config.FzfPath); err != nil {
			// path doesn't exist, leave empty (optional tool)
			config.FzfPath = ""
		}
	}

	return config, nil
}

type envConfig struct {
	Debug      bool   `env:"TM_DEBUG"`
	TmuxPath   string `env:"TM_TMUX_PATH"`
	FzfPath    string `env:"TM_FZF_PATH"`
	ConfigPath string `env:"TM_CONFIG_PATH"`
}

func loadConfigFromEnv() (*envConfig, error) {
	ctx := context.Background()
	var c envConfig
	if err := envconfig.Process(ctx, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

type fileConfig struct {
	PreDefinedSessions []struct {
		Dir     string   `yaml:"dir"`
		Name    string   `yaml:"name"`
		Aliases []string `yaml:"aliases"`
	} `yaml:"sessions"`
	SmartDirectories []string `yaml:"smart_directories"`
}

func loadConfigFromConfigFile(path string, dPath string) (*fileConfig, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) && path == dPath {
			if err := os.MkdirAll(filepath.Dir(dPath), 0755); err != nil {
				return nil, err
			}
			if err := os.WriteFile(dPath, []byte(""), 0644); err != nil {
				return nil, err
			}
			return &fileConfig{}, nil
		}
		return nil, fmt.Errorf("config file does not exist: %w", err)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var c fileConfig
	if err := yaml.Unmarshal(content, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func defaultConfigPath() (string, error) {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return "", fmt.Errorf("could not find home directory")
	}
	return filepath.Join(homeDir, ".config", "tm", "config.yaml"), nil
}
