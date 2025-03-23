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

TM loads its configuration from environment variables. More details coming soon.
