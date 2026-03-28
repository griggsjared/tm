# TM - Tmux Session Manager

A smart tmux session manager with fuzzy-finding support.

## Features

- **Fuzzy finding** with fzf integration for quick session selection
- Automatically finds and attaches to existing tmux sessions
- Support for pre-defined sessions with aliases
- Smart directory matching for project-based sessions
- Simple command-line interface

## Installation

```
go install github.com/griggsjared/tm@latest
```

**Optional dependency:** [fzf](https://github.com/junegunn/fzf) for fuzzy-finding. Install it for the best experience.

## Usage

### Quick start

```bash
tm              # Opens fzf with all available sessions
tm session-name # Exact match or fuzzy search
tm ls           # List active sessions
tm ls-all       # List all sessions (including pre-defined)
```

### How it works

1. **Exact match**: If you provide a session name and it exists, attaches immediately
2. **Single partial match**: If only one session matches your input, attaches immediately
3. **Multiple matches**: Opens fzf with matching sessions for you to select
4. **No matches**: Opens fzf with all sessions, or prints list if fzf unavailable

## Configuration

TM loads configuration from environment variables and a YAML config file.

### Environment Variables

- `TM_DEBUG`: Enable debug mode (`true` or `false`). Defaults to `false`.
- `TM_TMUX_PATH`: Path to the tmux binary. Defaults to `tmux` in PATH.
- `TM_FZF_PATH`: Path to the fzf binary. Defaults to `fzf` in PATH.
- `TM_FZF_OPTS`: Space-separated fzf options for customizing the UI. Defaults to `--height=20% --ansi --reverse`.
- `TM_CONFIG_PATH`: Path to the config file. Defaults to `~/.config/tm/config.yaml`.

### Config File

Located at `~/.config/tm/config.yaml` by default.

```yaml
sessions:
  -
    dir: ~/.config/app1
    name: app1
    aliases:
      - config1
      - settings1
  -
    dir: ~/.config/app2
    name: app2
    aliases:
      - config2
      - settings2

smart_directories:
  - ~/projects
  - ~/work
  - ~/school
```

## Development

```
make test    # Run tests with coverage
make build   # Build the binary
make lint    # Run linter
make fmt     # Format code
```

### Project Structure

```
tm/
├── main.go              # Entry point and wiring
├── Makefile             # Development commands
├── internal/
│   ├── app/             # Application orchestration
│   ├── config/          # Configuration loading
│   ├── fzf/             # Fuzzy finding integration
│   ├── session/         # Session domain (Finder)
│   └── tmux/            # Tmux client
```

### Architecture

- **app**: Orchestrates the flow - handles fuzzy selection, filtering, and tmux operations
- **session.Finder**: Discovers sessions from multiple sources (tmux, pre-defined, smart directories)
- **tmux.Client**: Low-level tmux operations (create, attach, check existence)
- **fzf**: Fuzzy finding integration (optional)

### Building from Source

```bash
make build
./tm
```
