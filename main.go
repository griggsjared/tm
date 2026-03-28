package main

import (
	"fmt"
	"os"

	"github.com/griggsjared/tm/internal/app"
	"github.com/griggsjared/tm/internal/config"
	"github.com/griggsjared/tm/internal/fzf"
	"github.com/griggsjared/tm/internal/session"
	"github.com/griggsjared/tm/internal/tmux"
)

var version = "dev"

func main() {
	os.Exit(run())
}

func run() int {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("tm version", version)
		return 0
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return 1
	}

	tmuxRunner := tmux.NewRunner()
	tmuxClient := tmux.NewClient(tmuxRunner, cfg.TmuxPath)
	sessionFinder := session.NewFinder(tmuxClient, cfg.PreDefinedSessions, cfg.SmartDirectories)
	fzfRunner := fzf.NewRunner(cfg.FzfPath)

	app.New(
		tmuxClient,
		sessionFinder,
		fzfRunner,
		cfg.Debug,
	).Run()

	return 0
}
