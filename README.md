# TM - Tmux Session Manager

A smart tmux session manager that helps you create and attach to tmux sessions quickly.

## Features

- Automatically finds and attaches to existing tmux sessions
- Support for pre-defined sessions
- Smart directory matching for project-based sessions
- Falls back to creating sessions in the current directory
- Simple command-line interface

## Installation

```
go install github.com/griggsjared/tm@latest
```

## Usage

```
tm session-name
```

The application will:
1. Look for existing sessions with the provided name
2. Check pre-defined sessions configuration
3. Search smart directory configurations for matching projects
4. Create a new session in the current directory if nothing matches

## Configuration

TM loads its configuration from both environment variables and a YAML config file.

### Environment Variables

- `TM_DEBUG`: Enable debug mode (`true` or `false`). Defaults to `false`. Used to show extra debug information.
- `TM_TMUX_PATH`: Path to the tmux binary. Defualt to `tmux` found in your path.
- `TM_CONFIG_PATH`: Path to the config file. Defaults to `~/.config/tm/config.yaml`.

### Config File

The config file is a YAML file located at `~/.config/tm/config.yaml` by default. It defines pre-defined sessions and smart directories.

#### Example config:

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
