package main

import (
	"fmt"
	"os/exec"
)

func main() {

	config, err := loadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	newApp(config, tmuxPath).run()
}
