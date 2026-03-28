package fzf

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Runner struct {
	path string
}

// NewRunner creates a new Runner with an optional custom path.
// If path is empty, it searches for fzf in PATH.
func NewRunner(path string) *Runner {
	if path == "" {
		if fzfPath, err := exec.LookPath("fzf"); err == nil {
			path = fzfPath
		}
	}
	return &Runner{path: path}
}

func (r *Runner) IsAvailable() bool {
	return r.path != ""
}

func (r *Runner) Select(items []string, query string) (string, bool, error) {
	if !r.IsAvailable() {
		return "", false, fmt.Errorf("fzf is not available")
	}

	args := []string{
		"--height=40%",
		"--reverse",
		"--select-1",
		"--exit-0",
		fmt.Sprintf("--query=%s", query),
	}

	cmd := exec.Command(r.path, args...)

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return "", false, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	cmd.Stderr = os.Stderr

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", false, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", false, fmt.Errorf("failed to start fzf: %w", err)
	}

	go func() {
		defer stdinPipe.Close()
		for _, item := range items {
			fmt.Fprintln(stdinPipe, item)
		}
	}()

	output, err := io.ReadAll(stdoutPipe)
	if err != nil {
		return "", false, fmt.Errorf("failed to read fzf output: %w", err)
	}
	selected := strings.TrimSpace(string(output))

	err = cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 || exitErr.ExitCode() == 130 {
				return "", false, nil
			}
		}
		return "", false, fmt.Errorf("fzf error: %w", err)
	}

	if selected == "" {
		return "", false, nil
	}

	return selected, true, nil
}
