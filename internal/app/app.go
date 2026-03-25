package app

import (
	"fmt"
	"os"

	"github.com/griggsjared/tm/internal/session"
)

type TmuxRunner interface {
	NewSession(s *session.Session) error
	AttachSession(s *session.Session) error
}

type SessionService interface {
	Find(name string) (*session.Session, error)
	List(onlyActive bool) []*session.Session
}

type App struct {
	debug          bool
	tmuxRunner     TmuxRunner
	sessionService SessionService
}

func New(tr TmuxRunner, ss SessionService, debug bool) *App {
	return &App{
		debug:          debug,
		tmuxRunner:     tr,
		sessionService: ss,
	}
}

func (a *App) Run() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a session name")
		return
	}

	input := os.Args[1]

	if input == "ls" || input == "list" || input == "ls-all" || input == "list-all" {
		all := input == "ls-all" || input == "list-all"
		for _, s := range a.sessionService.List(!all) {
			line := fmt.Sprintf("%s [%s]", s.Name, s.Dir)
			if s.Exists {
				line += "*"
			}
			fmt.Println(line)
		}
		return
	}

	s, err := a.sessionService.Find(input)
	if err != nil {
		fmt.Println("Error finding session:", err)
	}

	a.debugMsg(fmt.Sprintf("Session: %s, dir: %s, exists: %t", s.Name, s.Dir, s.Exists))

	if !s.Exists {
		a.debugMsg(fmt.Sprintf("Creating new session: %s and setting cwd to %s", s.Name, s.Dir))
		err = a.tmuxRunner.NewSession(s)
		if err != nil {
			fmt.Println("Error creating session:", err)
		}
	}

	a.debugMsg(fmt.Sprintf("Attaching to session: %s", s.Name))
	err = a.tmuxRunner.AttachSession(s)
	if err != nil {
		fmt.Println("Error attaching to session:", err)
	}
}

func (a *App) debugMsg(msg string) {
	if a.debug {
		fmt.Println(msg)
	}
}
