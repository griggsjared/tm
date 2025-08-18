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
	tmuxRepo := NewTmuxRepository(cmdRunner, config.TmuxPath)
	sessionService := NewSessionService(tmuxRepo, config.PreDefinedSessions, config.SmartDirectories)

	NewApp(
		tmuxRepo,
		sessionService,
		config.Debug,
	).Run()

	return 0
}
