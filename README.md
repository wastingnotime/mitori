# Mitori

Mitori is a local-first terminal Kanban for personal work observability.

## v0.2 scope

- single-user
- local JSON storage
- CLI + Bubble Tea TUI
- no deadlines, metrics, sync, or team features
- observability and ergonomics improvements over v1

Board lanes:

`backlog` `todo` `doing` `halt` `parking` `done` `archived`

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
go run ./cmd/mitori add "cz - collision playground - allow fullscreen mode"
go run ./cmd/mitori projects
go run ./cmd/mitori doctor
go run ./cmd/mitori archive
go run ./cmd/mitori archive crash
go run ./cmd/mitori import          # placeholder, not implemented yet
go run ./cmd/mitori export          # placeholder, not implemented yet
```

Default data path:

`~/.config/mitori/data.json`

You can override the data file path with:

`MITORI_DATA_PATH=/path/to/data.json`

## Structured input parser

Input pattern:

`initiative - project - title`

Compatible pattern:

`initiative - project -> title`

Example:

`cz - collision playground - allow fullscreen mode`

Parsed fields:

- initiative: `cz`
- project: `collision playground`
- title: `allow fullscreen mode`

Deterministic type inference is applied from title text:

- `fix` keywords: bug/fix/crash/error/regress/patch...
- `refact` keywords: refactor/cleanup/restructure/rename...
- `discovery` keywords: investigate/research/explore/spike...
- `feat` keywords: add/implement/enable/support/allow...
- fallback: `chore`

## TUI additions in v0.2

- Board filters with `/`:
  - free-text search over title, description, project
  - token forms: `project turtle`, `/project turtle`, `project:"name"`, `initiative cz`, `energy produces`, `type discovery`
- Archive browser with `A` (respects active filters)
- Manual reorder within lane with `K` (up) and `J` (down)
- Lane-valid actions are shown per selected task and executed via semantic keys (for example: `p park/unpark`, `d done`, `b backlog`, `x archive`)
- Recovery/destructive actions require `y` confirmation
- Project view (`g`) now includes:
  - counts by lane
  - recent project activity
- Task detail includes payload-aware event history
- Transition model:
  - normal and recovery transitions are centralized in domain/service logic
  - `backlog -> archived` requires confirmation
  - `done -> archived` is a normal forward transition
  - archived restore paths (`archived -> done/backlog`) require confirmation
  - edit/delete are backlog-only lane actions
- Navigation:
  - `q` quits app
  - `Esc` backs out of screen/input
