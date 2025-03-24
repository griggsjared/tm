package main

import (
	"fmt"
	"os"
)

func main() {
	run()
}

func run() {
	config, err := LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	cmdRunner := NewTmuxCommandRunner()
	tmuxRunner := NewTmuxRunner(cmdRunner, config.tmuxPath)
	sessionFinder := NewSessionFinder(tmuxRunner, config.pds, config.sds)

	NewApp(
		tmuxRunner,
		sessionFinder,
		config,
	).Run()
}
