package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {

	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		fmt.Println("tmux not found")
		return
	}
	
  config, err := loadConfigFromEnv()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	ts := NewTmuxService(tmuxPath, &TmuxCommandRunner{})
	sf := NewSessionFinder(ts.HasSession, config)

	//we only want the first argument from the cli,
	//if there are less than 1 error, of there are more than 1 jsut ignore the rest
	if len(os.Args) < 2 {
		fmt.Println("Please provide a session name")
		return
	}

	//the input will be a session name
	session, err := sf.Find(os.Args[1])
	if err != nil {
		fmt.Println("Error finding session:", err)
	}

	config.debugMsg(fmt.Sprintf("Session: %s, dir: %s, exists: %t", session.name, session.dir, session.exists))

	// if our founds sessions do not exist yet create one
	// first we set the os dir to the session dir
	// then we create a new session
	if session.exists == false {
		config.debugMsg(fmt.Sprintf("Creating new session: %s and setting cwd to %s", session.name, session.dir))
		ts.NewSession(session)
	}

	// finally join the session
	config.debugMsg(fmt.Sprintf("Attaching to session: %s", session.name))
	ts.AttachSession(session)
}
