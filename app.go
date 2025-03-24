package main

import (
	"fmt"
	"os"
)

type App struct {
	config        *Config
	tmuxRunner    *TmuxRunner
	sessionFinder *SessionFinder
}

func NewApp(tmuxRunner *TmuxRunner, sessionFinder *SessionFinder, config *Config) *App {
	return &App{
		config:        config,
		tmuxRunner:    tmuxRunner,
		sessionFinder: sessionFinder,
	}
}

func (a *App) debugMsg(msg string) {
	if a.config.debug {
		fmt.Println(msg)
	}
}

func (a *App) Run() {

	//we only want the first argument from the cli,
	//if there are less than 1 error, of there are more than 1 jsut ignore the rest
	if len(os.Args) < 2 {
		fmt.Println("Please provide a session name")
		return
	}

	//the input will be a session name
	session, err := a.sessionFinder.Find(os.Args[1])
	if err != nil {
		fmt.Println("Error finding session:", err)
	}

	a.debugMsg(fmt.Sprintf("Session: %s, dir: %s, exists: %t", session.name, session.dir, session.exists))

	// if our founds sessions do not exist yet create one
	// first we set the os dir to the session dir
	// then we create a new session
	if session.exists == false {
		a.debugMsg(fmt.Sprintf("Creating new session: %s and setting cwd to %s", session.name, session.dir))
		a.tmuxRunner.NewSession(session)
	}

	// finally join the session
	a.debugMsg(fmt.Sprintf("Attaching to session: %s", session.name))
	a.tmuxRunner.AttachSession(session)
}
