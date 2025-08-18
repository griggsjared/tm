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

// AppSessionService is an interface that defines a function that finds a session by name
type AppSessionService interface {
	Find(name string) (*Session, error)
	List(onlyActive bool) []*Session
}

// App is a struct that defines the application and its dependencies
type App struct {
	debug          bool
	tmuxRunner     AppTmuxRunner
	sessionService AppSessionService
}

// NewApp is a constructor for the App struct
func NewApp(tmuxRunner AppTmuxRunner, sessionService AppSessionService, debug bool) *App {
	return &App{
		debug:          debug,
		tmuxRunner:     tmuxRunner,
		sessionService: sessionService,
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

	input := os.Args[1]

	if input == "ls" || input == "list" || input == "ls-all" || input == "list-all" {
		all := input == "ls-all" || input == "list-all"
		for _, session := range a.sessionService.List(!all) {
			line := fmt.Sprintf("%s (%s)", session.name, session.dir)
			if session.exists {
				line += "*"
			}
			fmt.Println(line)
		}
		return
	}


	//the input will be a session name
	session, err := a.sessionService.Find(input)
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
