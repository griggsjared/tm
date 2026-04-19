package fzf

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Runner struct {
	path string
}

// NewRunner creates a new Runner with an optional custom path.
// If path is empty, it searches for fzf in PATH.
// Customize fzf behavior via FZF_DEFAULT_OPTS in your environment.
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

func (r *Runner) Path() string {
	return r.path
}

func (r *Runner) Select(items []string, query string) (int, bool, error) {
	if !r.IsAvailable() {
		return 0, false, fmt.Errorf("fzf is not available")
	}

	args := make([]string, 0, 4)
	args = append(args, "--select-1", "--exit-0", "--with-nth=2..", fmt.Sprintf("--query=%s", query))

	cmd := exec.Command(r.path, args...)

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return 0, false, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	cmd.Stderr = os.Stderr

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return 0, false, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return 0, false, fmt.Errorf("failed to start fzf: %w", err)
	}

	go func() {
		defer stdinPipe.Close()
		for i, item := range items {
			fmt.Fprintf(stdinPipe, "%d\t%s\n", i+1, item)
		}
	}()

	output, err := io.ReadAll(stdoutPipe)
	if err != nil {
		return 0, false, fmt.Errorf("failed to read fzf output: %w", err)
	}
	selected := strings.TrimSpace(string(output))

	err = cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 || exitErr.ExitCode() == 130 {
				return 0, false, nil
			}
		}
		return 0, false, fmt.Errorf("fzf error: %w", err)
	}

	if selected == "" {
		return 0, false, nil
	}

	parts := strings.SplitN(selected, "\t", 2)
	if len(parts) == 0 {
		return 0, false, nil
	}

	idx, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, false, fmt.Errorf("failed to parse selection index: %w", err)
	}

	return idx - 1, true, nil
}
