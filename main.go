package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/griggsjared/tm/internal/app"
	"github.com/griggsjared/tm/internal/config"
	"github.com/griggsjared/tm/internal/fzf"
	"github.com/griggsjared/tm/internal/session"
	"github.com/griggsjared/tm/internal/status"
	"github.com/griggsjared/tm/internal/tmux"
)

var version = "dev"

func getVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return "dev"
}

func main() {
	os.Exit(run())
}

func run() int {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println("tm version", getVersion())
		return 0
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return 1
	}

	tmuxRunner := tmux.NewRunner()
	tmuxClient := tmux.NewClient(tmuxRunner, cfg.TmuxPath)
	fzfRunner := fzf.NewRunner(cfg.FzfPath)

	if len(os.Args) > 1 && os.Args[1] == "status" {
		return status.New(getVersion(), tmuxClient, fzfRunner).Run()
	}

	sessionFinder := session.NewFinder(tmuxClient, cfg.PreDefinedSessions, cfg.SmartDirectories)

	app.New(
		tmuxClient,
		sessionFinder,
		fzfRunner,
		cfg.Debug,
	).Run()

	return 0
}
