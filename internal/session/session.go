package session

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

type Session struct {
	Name   string
	Dir    string
	Exists bool
}

func New(name string, dir string, exists bool) *Session {
	return &Session{
		Name:   name,
		Dir:    dir,
		Exists: exists,
	}
}

type PreDefinedSession struct {
	Dir     string
	Name    string
	Aliases []string
}

type SmartDirectory struct {
	Dir string
}

type TmuxRepository interface {
	HasSession(name string) bool
	AllSessions() []*Session
}

type Finder struct {
	repository         TmuxRepository
	preDefinedSessions []PreDefinedSession
	smartDirectories   []SmartDirectory
}

func NewFinder(r TmuxRepository, pds []PreDefinedSession, sd []SmartDirectory) *Finder {
	return &Finder{
		repository:         r,
		preDefinedSessions: pds,
		smartDirectories:   sd,
	}
}

func (f *Finder) Find(name string) (*Session, error) {
	session, err := f.findExistingSession(name)
	if err != nil {
		return nil, err
	}
	if session != nil {
		return session, nil
	}

	if len(f.preDefinedSessions) > 0 {
		session, err = f.findPreDefinedSession(name)
		if err != nil {
			return nil, err
		}
		if session != nil {
			return session, nil
		}
	}

	if len(f.smartDirectories) > 0 {
		session, err = f.findSmartSessionDirectorySession(name)
		if err != nil {
			return nil, err
		}
		if session != nil {
			return session, nil
		}
	}

	return nil, nil
}

func (f *Finder) List(onlyExisting bool) []*Session {
	var sessions []*Session
	sessions = f.repository.AllSessions()

	if !onlyExisting {
		for _, pd := range f.getAllPreDefinedSessions() {
			found := slices.ContainsFunc(sessions, func(s *Session) bool {
				return s.Name == pd.Name
			})

			if !found {
				sessions = append(sessions, pd)
			}
		}

		for _, sd := range f.getAllSmartSessionDirectorySessions() {
			found := slices.ContainsFunc(sessions, func(s *Session) bool {
				return s.Name == sd.Name
			})

			if !found {
				sessions = append(sessions, sd)
			}
		}
	}
	return sessions
}

func (f *Finder) findExistingSession(name string) (*Session, error) {
	if f.repository.HasSession(name) {
		return New(name, "", true), nil
	}
	return nil, nil
}

func (f *Finder) findPreDefinedSession(name string) (*Session, error) {
	for _, pd := range f.preDefinedSessions {
		found := slices.Contains(nameToMatch(pd), name)
		if !found {
			continue
		}

		existing, err := f.findExistingSession(pd.Name)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return existing, nil
		}

		dir, err := expandHomeDir(pd.Dir)
		if err != nil {
			return nil, err
		}

		if !dirExists(dir) {
			continue
		}

		return New(pd.Name, dir, false), nil
	}
	return nil, nil
}

func (f *Finder) findSmartSessionDirectorySession(name string) (*Session, error) {
	for _, sd := range f.smartDirectories {
		dir, err := expandHomeDir(fmt.Sprintf("%s/%s", sd.Dir, name))
		if err != nil {
			return nil, err
		}

		if dirExists(dir) {
			return New(name, dir, false), nil
		}
	}
	return nil, nil
}

func (f *Finder) getAllPreDefinedSessions() []*Session {
	var sessions []*Session
	for _, pd := range f.preDefinedSessions {
		dir, err := expandHomeDir(pd.Dir)
		if err != nil {
			continue
		}
		sessions = append(sessions, New(pd.Name, dir, false))
		if len(pd.Aliases) > 0 {
			for _, alias := range pd.Aliases {
				sessions = append(sessions, New(alias, dir, false))
			}
		}
	}
	return sessions
}

func (f *Finder) getAllSmartSessionDirectorySessions() []*Session {
	var sessions []*Session
	for _, sd := range f.smartDirectories {
		dir, err := expandHomeDir(sd.Dir)
		if err != nil {
			continue
		}

		if dirExists(dir) {
			files, err := os.ReadDir(dir)
			if err != nil {
				continue
			}

			for _, file := range files {
				if !file.IsDir() {
					continue
				}

				sessions = append(sessions, New(file.Name(), fmt.Sprintf("%s/%s", dir, file.Name()), false))
			}
		}
	}

	return sessions
}

func nameToMatch(pds PreDefinedSession) []string {
	var names []string

	names = append(names, pds.Name)
	names = append(names, pds.Aliases...)

	return names
}

func dirExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

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
