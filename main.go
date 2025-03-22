package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		fmt.Println("tmux not found")
		return
	}

	//we only want the first argument from the cli,
	// if there are less than 1 error, of there are more than 1 jsut ignore the rest
	if len(os.Args) < 2 {
		fmt.Println("Please provide a session name")
		return
	}

	//the input will be a session name
	sName := os.Args[1]

	config, err := loadConfigFromEnv()
	if err != nil {
		fmt.Println("Error loading config:", err)
	}

	//check existing session
	session, err := matchExistingSession(sName, func(name string) bool {
		return exec.Command("tmux", "has-session", "-t", name).Run() == nil
	})
	if err != nil {
		fmt.Println("Error matching existing session:", err)
	}

	//check pre defined session
	if session == nil && len(config.pds) > 0 {
		session, err = matchPreDefinedSession(sName, config.pds)
		if err != nil {
			fmt.Println("Error matching pre defined session:", err)
			return
		}
	}

	//check smart session directories
	if session == nil && len(config.sds) > 0 {
		session, err = matchSmartSessionDirectories(sName, config.sds)
		if err != nil {
			fmt.Println("Error matching smart session directories:", err)
			return
		}
	}

	//nothing found, create a new session the the current directory
	if session == nil {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("Error getting current working directory:", err)
		}
		session = NewSession(sName, cwd, false)
	}

	config.debugMsg(fmt.Sprintf("Session: %s, dir: %s, exists: %t", session.name, session.dir, session.exists))

	// if our founds sessions do not exist yet create one
	// first we set the os dir to the session dir
	// then we create a new session
	if session.exists == false {
		config.debugMsg(fmt.Sprintf("Creating new session: %s and setting cwd to %s", session.name, session.dir))
		err := os.Chdir(session.dir)
		if err != nil {
			fmt.Println("Error changing directory:", err)
			return
		}
		syscall.Exec(tmuxPath, []string{"tmux", "new-session", "-s", session.name}, os.Environ())
	}

	// finally join the session
	config.debugMsg(fmt.Sprintf("Attaching to session: %s", session.name))
	syscall.Exec(tmuxPath, []string{"tmux", "attach-session", "-t", session.name}, os.Environ())
}
