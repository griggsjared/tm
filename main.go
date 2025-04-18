package main

import (
	"fmt"
	"os"
)

func main() {
	os.Exit(run())
}

func run() int {
	config, err := LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return 1
	}

	cmdRunner := NewTmuxCommandRunner()
	tmuxRunner := NewTmuxRunner(cmdRunner, config.TmuxPath)
	sessionFinder := NewSessionFinder(tmuxRunner, config.PreDefinedSessions, config.SmartDirectories)

	NewApp(
		tmuxRunner,
		sessionFinder,
		config.Debug,
	).Run()

	return 0
}
