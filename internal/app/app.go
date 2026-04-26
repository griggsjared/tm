package app

import (
	"fmt"
	"strings"

	"github.com/griggsjared/tm/internal/session"
)

type TmuxRunner interface {
	IsAvailable() bool
	InsideTmux() bool
	Path() string
	NewSession(s *session.Session, detached bool) error
	AttachSession(s *session.Session) error
	SwitchSession(s *session.Session) error
	CurrentSession() string
}

type SessionFinder interface {
	Find(name string) (*session.Session, error)
	List(onlyActive bool) []*session.Session
	ListExcluding(onlyExisting bool, exclude string) []*session.Session
}

type FzfRunner interface {
	IsAvailable() bool
	Path() string
	Select(items []string, query string) (int, bool, error)
}

type App struct {
	version       string
	debug         bool
	tmuxRunner    TmuxRunner
	sessionFinder SessionFinder
	fzfRunner     FzfRunner
}

func New(tr TmuxRunner, ss SessionFinder, fr FzfRunner, debug bool, version string) *App {
	return &App{
		version:       version,
		debug:         debug,
		tmuxRunner:    tr,
		sessionFinder: ss,
		fzfRunner:     fr,
	}
}

func (a *App) Run(query string) int {
	if query == "version" {
		fmt.Println("tm version", a.version)
		return 0
	}

	if query == "status" {
		return a.runStatus()
	}

	if !a.tmuxRunner.IsAvailable() {
		fmt.Println("Error: tmux not found. Install tmux or set TM_TMUX_PATH.")
		return 1
	}

	currentSession := a.currentSession()

	if query == "" {
		a.runInteractive(currentSession)
		return 0
	}

	a.runWithQuery(query, currentSession)
	return 0
}

func (a *App) currentSession() string {
	if a.tmuxRunner.InsideTmux() {
		return a.tmuxRunner.CurrentSession()
	}
	return ""
}

func (a *App) runStatus() int {
	exitCode := 0

	printStatusLine("tm", fmt.Sprintf("ok (%s)", a.version))

	if a.tmuxRunner.IsAvailable() {
		printStatusLine("tmux", fmt.Sprintf("ok (%s)", a.tmuxRunner.Path()))
	} else {
		printStatusLine("tmux", "missing")
		exitCode = 1
	}

	if a.fzfRunner.IsAvailable() {
		printStatusLine("fzf", fmt.Sprintf("ok (%s) (optional)", a.fzfRunner.Path()))
	} else {
		printStatusLine("fzf", "missing (optional)")
	}

	return exitCode
}

func printStatusLine(name, status string) {
	dots := strings.Repeat(".", 12-len(name))
	fmt.Printf("%s%s %s\n", name, dots, status)
}

func (a *App) runInteractive(currentSession string) {
	sessions := a.sessionFinder.ListExcluding(false, currentSession)
	selected, err := a.selectSession(sessions, "")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if selected != nil {
		if err := a.attachToSession(selected); err != nil {
			fmt.Println(err)
		}
	}
}

func (a *App) runWithQuery(query, currentSession string) {
	if a.handleBuiltinCommand(query) {
		return
	}

	// Try exact match first
	session, err := a.sessionFinder.Find(query)
	if err != nil {
		fmt.Println("Error finding session:", err)
		return
	}
	if session != nil {
		if err := a.attachToSession(session); err != nil {
			fmt.Println(err)
		}
		return
	}

	// No exact match - filter by partial
	allSessions := a.sessionFinder.ListExcluding(false, currentSession)
	matches := filterSessions(allSessions, query)

	if len(matches) == 1 {
		// Single partial match - attach directly
		if err := a.attachToSession(matches[0]); err != nil {
			fmt.Println(err)
		}
		return
	}

	// 0 or >1 matches - need selection
	selected, err := a.selectSession(allSessions, query)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if selected != nil {
		if err := a.attachToSession(selected); err != nil {
			fmt.Println(err)
		}
	}
}

func (a *App) selectSession(sessions []*session.Session, query string) (*session.Session, error) {
	if len(sessions) == 0 {
		if query == "" {
			fmt.Println("No sessions available")
		} else {
			fmt.Printf("No sessions matching %q\n", query)
		}
		return nil, nil
	}

	if a.fzfRunner.IsAvailable() {
		items := make([]string, len(sessions))
		for i, s := range sessions {
			items[i] = formatSessionLine(s)
		}

		idx, ok, err := a.fzfRunner.Select(items, query)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, nil // user cancelled
		}

		return sessions[idx], nil
	}

	// No fzf - print list
	fmt.Println("Available sessions:")
	for _, s := range sessions {
		fmt.Println(formatSessionLine(s))
	}
	fmt.Printf("\nProvide a more specific name (query: %q)\n", query)
	return nil, nil
}

func (a *App) attachToSession(s *session.Session) error {
	if a.tmuxRunner.InsideTmux() {
		if !s.Exists {
			a.debugMsg(fmt.Sprintf("Creating new session: %s", s.Name))
			if err := a.tmuxRunner.NewSession(s, true); err != nil {
				return fmt.Errorf("error creating session: %w", err)
			}
		}
		a.debugMsg(fmt.Sprintf("Switching to session: %s", s.Name))
		if err := a.tmuxRunner.SwitchSession(s); err != nil {
			return fmt.Errorf("error switching session: %w", err)
		}
		return nil
	}

	if !s.Exists {
		a.debugMsg(fmt.Sprintf("Creating new session: %s", s.Name))
		if err := a.tmuxRunner.NewSession(s, false); err != nil {
			return fmt.Errorf("error creating session: %w", err)
		}
	}
	a.debugMsg(fmt.Sprintf("Attaching to session: %s", s.Name))
	if err := a.tmuxRunner.AttachSession(s); err != nil {
		return fmt.Errorf("error attaching to session: %w", err)
	}
	return nil
}

func formatSessionLine(s *session.Session) string {
	line := fmt.Sprintf("%s [%s]", s.Name, s.Dir)
	if s.Exists {
		line += " *"
	}
	return line
}

func (a *App) debugMsg(msg string) {
	if a.debug {
		fmt.Println(msg)
	}
}

func (a *App) handleBuiltinCommand(input string) bool {
	switch input {
	case "ls", "list":
		a.printSessionList(true)
		return true
	case "ls-all", "list-all":
		a.printSessionList(false)
		return true
	}
	return false
}

func (a *App) printSessionList(onlyExisting bool) {
	for _, s := range a.sessionFinder.List(onlyExisting) {
		fmt.Println(formatSessionLine(s))
	}
}

func filterSessions(sessions []*session.Session, query string) []*session.Session {
	if query == "" {
		return sessions
	}
	query = strings.ToLower(query)
	var matches []*session.Session
	for _, s := range sessions {
		if strings.Contains(strings.ToLower(s.Name), query) {
			matches = append(matches, s)
		}
	}
	return matches
}
