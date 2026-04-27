package fzf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Runner interface {
	Output(path string, args []string) ([]byte, error)
	Run(path string, args []string, stdin io.Reader, stderr io.Writer) ([]byte, int, error)
}

type FzfRunner struct{}

func NewRunner() *FzfRunner {
	return &FzfRunner{}
}

func (r *FzfRunner) Output(path string, args []string) ([]byte, error) {
	return exec.Command(path, args...).Output()
}

func (r *FzfRunner) Run(path string, args []string, stdin io.Reader, stderr io.Writer) ([]byte, int, error) {
	cmd := exec.Command(path, args...)
	cmd.Stdin = stdin
	cmd.Stderr = stderr
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			return stdout.Bytes(), exitErr.ExitCode(), err
		}
		return stdout.Bytes(), -1, err
	}
	return stdout.Bytes(), 0, nil
}

type Client struct {
	runner Runner
	path   string
}

// NewClient creates a new Client with the given Runner and optional custom path.
// Customize fzf behavior via FZF_DEFAULT_OPTS in your environment.
func NewClient(r Runner, path string) *Client {
	return &Client{
		runner: r,
		path:   path,
	}
}

func (c *Client) IsAvailable() bool {
	return c.path != ""
}

func (c *Client) Path() string {
	return c.path
}

func (c *Client) Version() string {
	output, err := c.runner.Output(c.path, []string{"--version"})
	if err != nil {
		return ""
	}
	parts := strings.Fields(string(output))
	if len(parts) == 0 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

func (c *Client) Select(items []string, query string) (int, bool, error) {
	if !c.IsAvailable() {
		return 0, false, fmt.Errorf("fzf is not available")
	}

	args := make([]string, 0, 4)
	args = append(args, "--select-1", "--exit-0", "--with-nth=2..", fmt.Sprintf("--query=%s", query))

	var stdin bytes.Buffer
	for i, item := range items {
		fmt.Fprintf(&stdin, "%d\t%s\n", i+1, item)
	}

	output, exitCode, err := c.runner.Run(c.path, args, &stdin, os.Stderr)
	if err != nil {
		if exitCode == 1 || exitCode == 130 {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("fzf error: %w", err)
	}

	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return 0, false, nil
	}

	parts := strings.SplitN(selected, "\t", 2)

	idx, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, false, fmt.Errorf("failed to parse selection index: %w", err)
	}

	return idx - 1, true, nil
}
