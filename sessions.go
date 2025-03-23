package main

import (
	"fmt"
	"os"
)

// Session is a struct that defines tmux session that we
// want to use to either start or join a new tmux session
type Session struct {
	name   string
	dir    string
	exists bool
}

// New session is a constructor for the session struct
func NewSession(name string, dir string, exists bool) *Session {
	return &Session{
		name:   name,
		dir:    dir,
		exists: exists,
	}
}

// PreDefinedSession is a struct that defines a session that we want to have a custom dir to session an
type PreDefinedSession struct {
	dir  string
	name string
}

// matchPreDefinedSession is a function that checks if a session name matches on of any provided pre defined sessions
func matchPreDefinedSession(name string, pds []PreDefinedSession) (*Session, error) {
	for _, pd := range pds {
		if pd.name != name {
			continue
		}

		dir, err := expandHomeDir(pd.dir)
		if err != nil {
			return nil, err
		}

		if dirExists(dir) {
			return NewSession(name, dir, false), nil
		}
	}
	return nil, nil
}

// SmartDirectories is a struct of a directory that we can search through to find a sub directory matching a session name
type SmartSessionDirectories struct {
	dir string
}

// SmartSessionDirectories is a function that checks if a session name matches on of any provided smart session directories
func matchSmartSessionDirectories(name string, sds []SmartSessionDirectories) (*Session, error) {
	for _, sd := range sds {
		dir, err := expandHomeDir(fmt.Sprintf("%s/%s", sd.dir, name))
		if err != nil {
			return nil, err
		}

		if dirExists(dir) {
			return NewSession(name, dir, false), nil
		}
	}
	return nil, nil
}

// SessionChecker is a function that checks if a session exists
type SessionChecker func(name string) bool

// matchExistingSession is a function that checks if a session name matches on of any provided smart session directories
func matchExistingSession(name string, check SessionChecker) (*Session, error) {
	if check(name) {
		return NewSession(name, "", true), nil
	}
	return nil, nil
}

type SessionFinder struct {
	tmuxHasSession SessionChecker
	config         *config
}

func NewSessionFinder(tmuxHasSession SessionChecker, config *config) *SessionFinder {
	return &SessionFinder{
		tmuxHasSession: tmuxHasSession,
		config:         config,
	}
}

func (sf *SessionFinder) Find(name string) (*Session, error) {
	session, err := matchExistingSession(name, sf.tmuxHasSession)
	if err != nil {
		return nil, err
	}
	if session != nil {
		return session, nil
	}

	if len(sf.config.pds) > 0 {
		session, err = matchPreDefinedSession(name, sf.config.pds)
		if err != nil {
			return nil, err
		}
		if session != nil {
			return session, nil
		}
	}

	if len(sf.config.sds) > 0 {
		session, err = matchSmartSessionDirectories(name, sf.config.sds)
		if err != nil {
			return nil, err
		}
		if session != nil {
			return session, nil
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current working directory:", err)
		return nil, err
	}
	return NewSession(name, cwd, false), nil
}
