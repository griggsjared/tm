package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sethvargo/go-envconfig"
	yaml "gopkg.in/yaml.v3"
)

// Config is a struct that defines the configuration for the app
type Config struct {
	Debug              bool
	TmuxPath           string
	PreDefinedSessions []PreDefinedSession
	SmartDirectories   []SmartDirectory
}

// NewConfig is a constructor for the Config struct
func NewConfig(debug bool, tmuxPath string, preDefinedSessions []PreDefinedSession, smartDirectories []SmartDirectory) *Config {
	return &Config{
		Debug:              debug,
		TmuxPath:           tmuxPath,
		PreDefinedSessions: preDefinedSessions,
		SmartDirectories:   smartDirectories,
	}
}

// LoadConfig loads the final configuration from various sources
func LoadConfig() (*Config, error) {

	config := NewConfig(false, "", nil, nil)

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

	config.PreDefinedSessions = make([]PreDefinedSession, len(fileConfig.Pds))
	for i, pd := range fileConfig.Pds {
		config.PreDefinedSessions[i] = PreDefinedSession{
			dir:     pd.Dir,
			name:    pd.Name,
			aliases: pd.Aliases,
		}
	}

	config.SmartDirectories = make([]SmartDirectory, len(fileConfig.Sds))
	for i, sd := range fileConfig.Sds {
		config.SmartDirectories[i] = SmartDirectory{
			dir: sd,
		}
	}

	config.Debug = envConfig.Debug
	config.TmuxPath = envConfig.TmuxPath

	if config.TmuxPath != "" {
		if _, err := os.Stat(config.TmuxPath); err != nil {
			return nil, fmt.Errorf("tmux path does not exist: %w", err)
		}
	} else {
		tmuxPath, err := exec.LookPath("tmux")
		if err != nil {
			return nil, err
		}
		config.TmuxPath = tmuxPath
	}

	return config, nil
}

// envConfig is a struct that defines the environment configuration
type envConfig struct {
	Debug      bool   `env:"TM_DEBUG"`
	TmuxPath   string `env:"TM_TMUX_PATH"`
	ConfigPath string `env:"TM_CONFIG_PATH"`
}

// loadConfigFromEnv loads the config from the environment variables
func loadConfigFromEnv() (*envConfig, error) {
	ctx := context.Background()
	var c envConfig
	if err := envconfig.Process(ctx, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

// fileConfig is a struct that defines the file configuration
type fileConfig struct {
	Pds []struct {
		Dir     string   `yaml:"dir"`
		Name    string   `yaml:"name"`
		Aliases []string `yaml:"aliases"`
	} `yaml:"sessions"`
	Sds []string `yaml:"smart_directories"`
}

// loadConfigFromConfigFile loads the config from the config file
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

// defaultConfigPath returns the default config path
func defaultConfigPath() (string, error) {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return "", fmt.Errorf("could not find home directory")
	}
	return filepath.Join(homeDir, ".config", "tm", "config.yaml"), nil
}
