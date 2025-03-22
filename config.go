package main

import (
	"fmt"
	"os"
	"strings"
)

type config struct {
	debug bool
	pds   []PreDefinedSession
	sds   []SmartSessionDirectories
}

func newConfig(debug bool, pds []PreDefinedSession, sds []SmartSessionDirectories) *config {
	return &config{
		debug: debug,
		pds:   pds,
		sds:   sds,
	}
}

func (c *config) debugMsg(msg string) {
	if c.debug {
		fmt.Println(msg)
	}
}

//loadConfigFromEnv loads the config from the environment variables
// TM_DEBUG=true # or false
// TM_PREDEFINED_SESSIONS="name1:dir1,name2:dir2"
// TM_SMART_DIRECTORIES="dir1,dir2"
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

	return newConfig(debug, pds, sds), nil
}

