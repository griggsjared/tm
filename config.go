package main

import (
	"os"
	"os/exec"
	"strings"
)

type config struct {
	debug    bool
	tmuxPath string
	pds      []PreDefinedSession
	sds      []SmartSessionDirectories
}

func newConfig(debug bool, tmuxPath string, pds []PreDefinedSession, sds []SmartSessionDirectories) *config {
	return &config{
		debug:    debug,
		tmuxPath: tmuxPath,
		pds:      pds,
		sds:      sds,
	}
}

func loadConfig() (*config, error) {

	config, err := loadConfigFromEnv()
	if err != nil {
		return nil, err
	}

	if config.tmuxPath == "" {
		tmuxPath, err := exec.LookPath("tmux")
		if err != nil {
			return nil, err
		}
		config.tmuxPath = tmuxPath
	}

	return config, nil
}

// loadConfigFromEnv loads the config from the environment variables
// TM_DEBUG=true # or false
// TM_PREDEFINED_SESSIONS="name1:dir1,name2:dir2"
// TM_SMART_DIRECTORIES="dir1,dir2"
// TMUX_PATH="/path/to/tmux"
func loadConfigFromEnv() (*config, error) {
	debug := false

	// Check if debug is set in environment
	if debugEnv := os.Getenv("TM_DEBUG"); debugEnv != "" {
		debug = debugEnv == "true" || debugEnv == "1"
	}

	// Initialize empty slices
	pds := []PreDefinedSession{}
	sds := []SmartSessionDirectories{}

	// Parse predefined sessions (name:dir pairs)
	if predefinedStr := os.Getenv("TM_PREDEFINED_SESSIONS"); predefinedStr != "" {
		pairs := strings.Split(predefinedStr, ",")
		for _, pair := range pairs {
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) == 2 {
				pds = append(pds, PreDefinedSession{
					name: strings.TrimSpace(parts[0]),
					dir:  strings.TrimSpace(parts[1]),
				})
			}
		}
	}

	// Parse smart session directories (just directories)
	if smartDirsStr := os.Getenv("TM_SMART_DIRECTORIES"); smartDirsStr != "" {
		dirs := strings.Split(smartDirsStr, ",")
		for _, dir := range dirs {
			if trimmedDir := strings.TrimSpace(dir); trimmedDir != "" {
				sds = append(sds, SmartSessionDirectories{
					dir: trimmedDir,
				})
			}
		}
	}

	tmuxPath := os.Getenv("TMUX_PATH")

	return newConfig(debug, tmuxPath, pds, sds), nil
}
