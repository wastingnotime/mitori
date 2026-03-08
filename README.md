# Mitori

Mitori is a local-first terminal Kanban for personal work observability.

## v1.0 scope

- single-user
- local JSON storage
- CLI + Bubble Tea TUI
- no deadlines, metrics, sync, or team features

Board lanes:

`backlog` `todo` `doing` `halt` `parking` `done`

## Requirements

- Go 1.22+

## Run

```bash
go run ./cmd/mitori init
go run ./cmd/mitori
```

## CLI commands

```bash
go run ./cmd/mitori                 # launch TUI board
go run ./cmd/mitori init            # initialize data file
go run ./cmd/mitori add "cz - collision playground -> allow fullscreen mode"
go run ./cmd/mitori projects
go run ./cmd/mitori doctor
```

Default data path:

`~/.config/mitori/data.json`

You can override the data file path with:

`MITORI_DATA_PATH=/path/to/data.json`

## Structured input parser

Input pattern:

`initiative - project -> title`

Example:

`cz - collision playground -> allow fullscreen mode`

Parsed fields:

- initiative: `cz`
- project: `collision playground`
- title: `allow fullscreen mode`
