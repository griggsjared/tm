package main

import (
	"fmt"
	"os"
)

func main() {

	config, err := LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	NewApp(config).Run()
}
