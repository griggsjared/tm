package main

import (
	"fmt"
	"os"
)

// AppTmuxRunner is an interface that defines a function that runs a CommandRunner that is used by the TmuxRunner.
type AppTmuxRunner interface {
	NewSession(s *Session) error
	AttachSession(s *Session) error
}

// AppSessionFinder is an interface that defines a function that finds a session by name
type AppSessionFinder interface {
	Find(name string) (*Session, error)
}

// App is a struct that defines the application and its dependencies
type App struct {
	debug         bool
	tmuxRunner    AppTmuxRunner
	sessionFinder AppSessionFinder
}

// NewApp is a constructor for the App struct
func NewApp(tmuxRunner AppTmuxRunner, sessionFinder AppSessionFinder, debug bool) *App {
	return &App{
		debug:         debug,
		tmuxRunner:    tmuxRunner,
		sessionFinder: sessionFinder,
	}
}

// Run is a function that bootstraps and runs the application
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

	// if our found session does not exist yet create one
	// first we set the os dir to the session dir
	// then we create a new session
	if !session.exists {
		a.debugMsg(fmt.Sprintf("Creating new session: %s and setting cwd to %s", session.name, session.dir))
		err = a.tmuxRunner.NewSession(session)
		if err != nil {
			fmt.Println("Error creating session:", err)
		}
	}

	// finally join the session
	a.debugMsg(fmt.Sprintf("Attaching to session: %s", session.name))
	err = a.tmuxRunner.AttachSession(session)
	if err != nil {
		fmt.Println("Error attaching to session:", err)
	}
}

// debugMsg is a function that prints a message if the debug flag is set
func (a *App) debugMsg(msg string) {
	if a.debug {
		fmt.Println(msg)
	}
}
