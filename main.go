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
	tmuxRunner := NewTmuxRunner(cmdRunner, config.tmuxPath)
	sessionFinder := NewSessionFinder(tmuxRunner, config.preDefinedSessions, config.smartDirectories)

	NewApp(
		tmuxRunner,
		sessionFinder,
		config.debug,
	).Run()

  return 0
}
