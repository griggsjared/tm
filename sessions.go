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

// SessionRepository is a interface that defines a function that checks if a session exists
type SessionRepository interface {
	HasSession(name string) bool // HasSession checks if a session with the given name exists
}

// SessionService is a struct that defines a session finder
type SessionService struct {
	sessionRepository  SessionRepository
	preDefinedSessions []PreDefinedSession
	smartDirectories   []SmartDirectory
}

// NewSessionService is a constructor for the SessionFinder struct
func NewSessionService(sr SessionRepository, rds []PreDefinedSession, sd []SmartDirectory) *SessionService {
	return &SessionService{
		sessionRepository:  sr,
		preDefinedSessions: rds,
		smartDirectories:   sd,
	}
}

// Find is a function that finds a session by name
func (ss *SessionService) Find(name string) (*Session, error) {
	session, err := ss.findExistingSession(name)
	if err != nil {
		return nil, err
	}
	if session != nil {
		return session, nil
	}

	if len(ss.preDefinedSessions) > 0 {
		session, err = ss.findPreDefinedSession(name)
		if err != nil {
			return nil, err
		}
		if session != nil {
			return session, nil
		}
	}

	if len(ss.smartDirectories) > 0 {
		session, err = ss.findSmartSessionDirectories(name)
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

// findExistingSession is a function that checks if a session name matches on of any provided smart session directories
func (ss *SessionService) findExistingSession(name string) (*Session, error) {
	if ss.sessionRepository.HasSession(name) {
		return NewSession(name, "", true), nil
	}
	return nil, nil
}

// findPreDefinedSession is a function that checks if a session name matches on of any provided pre defined sessions,
// if the a name is not found, we check for any matching aliases
func (ss *SessionService) findPreDefinedSession(name string) (*Session, error) {
	for _, pd := range ss.preDefinedSessions {
		found := slices.Contains(nameToMatch(pd), name)
		if !found {
			continue
		}

		// We may have matched on an alias for al already running session
		// if that session exists we can just return that session instead
		// of starting a new one,
		existing, err := ss.findExistingSession(pd.name)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return existing, nil
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

// findSmartSessionDirectories is a function that checks if a session name matches on of any provided smart session directories
func (ss *SessionService) findSmartSessionDirectories(name string) (*Session, error) {
	for _, sd := range ss.smartDirectories {
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

// getAllPreDefinedSessions is a function that returns all pre defined sessions, each al
func (ss *SessionService) getAllPreDefinedSessions() []*Session {
	var sessions []*Session
	for _, pd := range ss.preDefinedSessions {
		dir, err := expandHomeDir(pd.dir)
		if err != nil {
			fmt.Println("Error expanding home directory:", err)
			continue
		}
		sessions = append(sessions, NewSession(pd.name, dir, false))
		if len(pd.aliases) > 0 {
			for _, alias := range pd.aliases {
				sessions = append(sessions, NewSession(alias, dir, false))
			}
		}
	}
	return sessions
}

// nameToMatch is a function that returns a slice of strings that contains the name and aliases of a pre defined session
func nameToMatch(pds PreDefinedSession) []string {
	var names []string

	names = append(names, pds.name)
	names = append(names, pds.aliases...)

	return names
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

