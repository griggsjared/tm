package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
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
	dir     string
	name    string
	aliases []string
}

// SmartDirectory is a struct of a directory that we can search through to find a sub directory matching a session name
type SmartDirectory struct {
	dir string
}

// SessionChecker is a interface that defines a function that checks if a session exists
type SessionChecker interface {
	HasSession(name string) bool
}

// SessionFinder is a struct that defines a session finder
type SessionFinder struct {
	sessionChecker     SessionChecker
	preDefinedSessions []PreDefinedSession
	smartDirectories   []SmartDirectory
}

// NewSessionFinder is a constructor for the SessionFinder struct
func NewSessionFinder(tmuxHasSession SessionChecker, preDefinedSessions []PreDefinedSession, smartDirectories []SmartDirectory) *SessionFinder {
	return &SessionFinder{
		sessionChecker:     tmuxHasSession,
		preDefinedSessions: preDefinedSessions,
		smartDirectories:   smartDirectories,
	}
}

// Find is a function that finds a session by name
func (sf *SessionFinder) Find(name string) (*Session, error) {
	session, err := sf.findExistingSession(name)
	if err != nil {
		return nil, err
	}
	if session != nil {
		return session, nil
	}

	if len(sf.preDefinedSessions) > 0 {
		session, err = sf.findPreDefinedSession(name)
		if err != nil {
			return nil, err
		}
		if session != nil {
			return session, nil
		}
	}

	if len(sf.smartDirectories) > 0 {
		session, err = sf.findSmartSessionDirectories(name)
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

// findPreDefinedSession is a function that checks if a session name matches on of any provided pre defined sessions,
// if the a name is not found, we check for any matching aliases
func (sf *SessionFinder) findPreDefinedSession(name string) (*Session, error) {
	for _, pd := range sf.preDefinedSessions {
		found := slices.Contains(nameToMatch(pd), name)
		if !found {
			continue
		}

		dir, err := expandHomeDir(pd.dir)
		if err != nil {
			return nil, err
		}

		if !dirExists(dir) {
			continue
		}

		return NewSession(pd.name, dir, false), nil
	}
	return nil, nil
}

// nameToMatch is a function that returns a slice of strings that contains the name and aliases of a pre defined session
func nameToMatch(pds PreDefinedSession) []string {
	var names []string

	names = append(names, pds.name)
	names = append(names, pds.aliases...)

	return names
}

// findSmartSessionDirectories is a function that checks if a session name matches on of any provided smart session directories
func (sf *SessionFinder) findSmartSessionDirectories(name string) (*Session, error) {
	for _, sd := range sf.smartDirectories {
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

// findExistingSession is a function that checks if a session name matches on of any provided smart session directories
func (sf *SessionFinder) findExistingSession(name string) (*Session, error) {
	if sf.sessionChecker.HasSession(name) {
		return NewSession(name, "", true), nil
	}
	return nil, nil
}

// dirExists is a function that checks if a directory exists
func dirExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// expandHomeDir is a function that expands a path that starts with ~ to the home directory
func expandHomeDir(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return strings.Replace(path, "~", homeDir, 1), nil
	}
	return path, nil
}
