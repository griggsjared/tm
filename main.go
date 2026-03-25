package main

import (
	"fmt"
	"os"

	"github.com/griggsjared/tm/internal/app"
	"github.com/griggsjared/tm/internal/config"
	"github.com/griggsjared/tm/internal/session"
	"github.com/griggsjared/tm/internal/tmux"
)

func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return 1
	}

	cmdRunner := tmux.NewCommandRunner()
	tmuxRepo := tmux.NewRepository(cmdRunner, cfg.TmuxPath)
	sessionService := session.NewService(tmuxRepo, cfg.PreDefinedSessions, cfg.SmartDirectories)

	app.New(
		tmuxRepo,
		sessionService,
		cfg.Debug,
	).Run()

	return 0
}
