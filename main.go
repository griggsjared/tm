package main

import (
	"fmt"
	"os/exec"
)

func main() {

	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		fmt.Println("tmux not found")
		return
	}

	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	newApp(config, tmuxPath).run()
}
