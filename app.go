package main

import (
	"fmt"
	"os"
)

type app struct {
	config        *config
	tmuxService   *TmuxService
	sessionFinder *SessionFinder
}

func newApp(c *config, tmuxPath string) *app {
	ts := NewTmuxService(tmuxPath, &TmuxCommandRunner{})
	sd := NewSessionFinder(ts.HasSession, c)

	return &app{
		config:        c,
		tmuxService:   ts,
		sessionFinder: sd,
	}
}

func (a *app) debugMsg(msg string) {
	if a.config.debug {
		fmt.Println(msg)
	}
}

func (a *app) run() {

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
		a.tmuxService.NewSession(session)
	}

	// finally join the session
	a.debugMsg(fmt.Sprintf("Attaching to session: %s", session.name))
	a.tmuxService.AttachSession(session)
}
