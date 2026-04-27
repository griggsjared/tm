package session

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"strings"
)

type Session struct {
	Name         string
	Dir          string
	Exists       bool
	LastAttached int64
	Aliases      []string
}

func New(name string, dir string, exists bool, lastAttached int64) *Session {
	return &Session{
		Name:         name,
		Dir:          dir,
		Exists:       exists,
		LastAttached: lastAttached,
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

func (f *Finder) ListExcluding(onlyExisting bool, exclude string) []*Session {
	all := f.List(onlyExisting)
	if exclude == "" {
		return all
	}

	exclusionSet := make(map[string]struct{})
	exclusionSet[exclude] = struct{}{}

	for _, pd := range f.preDefinedSessions {
		names := nameToMatch(pd)
		if slices.Contains(names, exclude) {
			for _, n := range names {
				exclusionSet[n] = struct{}{}
			}
		}
	}

	var filtered []*Session
	for _, s := range all {
		if _, ok := exclusionSet[s.Name]; !ok {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func (f *Finder) List(onlyExisting bool) []*Session {
	var sessions []*Session
	sessions = f.repository.AllSessions()

	for _, s := range sessions {
		for _, pd := range f.preDefinedSessions {
			if s.Name == pd.Name {
				s.Aliases = pd.Aliases
				break
			}
		}
	}

	slices.SortFunc(sessions, func(a, b *Session) int {
		return cmp.Compare(b.LastAttached, a.LastAttached)
	})

	if !onlyExisting {
		var nonExisting []*Session

		for _, pd := range f.getAllPreDefinedSessions() {
			found := slices.ContainsFunc(sessions, func(s *Session) bool {
				return s.Name == pd.Name
			})

			if !found {
				nonExisting = append(nonExisting, pd)
			}
		}

		for _, sd := range f.getAllSmartSessionDirectorySessions() {
			found := slices.ContainsFunc(sessions, func(s *Session) bool {
				return s.Name == sd.Name
			})

			if !found {
				nonExisting = append(nonExisting, sd)
			}
		}

		slices.SortFunc(nonExisting, func(a, b *Session) int {
			return cmp.Compare(a.Name, b.Name)
		})

		sessions = append(sessions, nonExisting...)
	}
	return sessions
}

func (f *Finder) findExistingSession(name string) (*Session, error) {
	if f.repository.HasSession(name) {
		return New(name, "", true, 0), nil
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
			existing.Aliases = pd.Aliases
			return existing, nil
		}

		dir, err := expandHomeDir(pd.Dir)
		if err != nil {
			return nil, err
		}

		if !dirExists(dir) {
			continue
		}

		s := New(pd.Name, dir, false, 0)
		s.Aliases = pd.Aliases
		return s, nil
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
			return New(name, dir, false, 0), nil
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
		s := New(pd.Name, dir, false, 0)
		s.Aliases = pd.Aliases
		sessions = append(sessions, s)
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

				sessions = append(sessions, New(file.Name(), fmt.Sprintf("%s/%s", dir, file.Name()), false, 0))
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
