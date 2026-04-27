package main

import (
	"fmt"
	"os"
	"runtime/debug"

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
	cmd := parseCommand(os.Args[1:])

	cfg, err := config.Load()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return 1
	}

	tmuxClient := tmux.NewClient(tmux.NewRunner(), cfg.TmuxPath)
	fzfClient := fzf.NewClient(fzf.NewRunner(), cfg.FzfPath)
	sessionFinder := session.NewFinder(tmuxClient, cfg.PreDefinedSessions, cfg.SmartDirectories)

	return app.New(tmuxClient, fzfClient, sessionFinder, cfg.Debug, getVersion()).Run(cmd)
}

func parseCommand(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return ""
}

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
