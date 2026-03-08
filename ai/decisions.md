# Mitori — Architecture Decisions Log

This file records meaningful technical and product decisions made during development.

Its purpose is to:
- preserve reasoning
- reduce drift across AI coding sessions
- avoid re-discussing solved problems
- make future refactors easier

---

## Status Legend

- **proposed** → under consideration
- **accepted** → chosen and active
- **deprecated** → no longer preferred, but still present
- **superseded** → replaced by another decision

---

# ADR-0001 — Product identity

- **Status:** accepted
- **Date:** 2026-03-07

## Decision

Mitori is a **local-first terminal Kanban application for personal work observability**.

## Rationale

The goal is to make work visible so the user can steer manually.
Mitori is not intended to be:
- a productivity tracker
- a team workflow platform
- a recommendation engine
- a project planning SaaS

## Consequences

All features must support:
- observability
- manual control
- low cognitive overhead

Features that push Mitori toward metrics, automation, or collaboration should be rejected unless explicitly requested.

---

# ADR-0002 — Project is the true balancing unit

- **Status:** accepted
- **Date:** 2026-03-07

## Decision

The main work container is **project**, not initiative.

## Rationale

Initiative is only a classifier.
The user actually balances and moves across multiple projects in parallel.

Examples:
- initiative: CZ
- project: collision playground

The user intentionally interlaces projects to regulate curiosity, anxiety, and sustained progress.

## Consequences

UI and services must make project identity highly visible.
Do not model the workflow as “one active initiative.”
Avoid assumptions that parallel projects are a problem to solve.

---

# ADR-0003 — Board lanes

- **Status:** accepted
- **Date:** 2026-03-07

## Decision

The board lanes are:

```text
backlog
todo
doing
halt
parking
done
````

## Rationale

These lanes reflect the user’s real physical Kanban process.

Their meanings are:

* backlog: high-level intents not yet refined
* todo: tasks ready to start
* doing: currently active work
* halt: externally blocked work
* parking: intentionally paused work
* done: finished work awaiting review or archival

## Consequences

These lane names and semantics should remain stable in v1.
Do not introduce workflow states beyond these unless explicitly requested.

---

# ADR-0004 — v1 storage backend

* **Status:** accepted
* **Date:** 2026-03-07

## Decision

v1 uses a single local JSON file as the primary storage backend.

## Rationale

The user wants:

* local-first behavior
* low friction
* no OS-level dependency issues
* human-inspectable data
* easy backups

A JSON file is the simplest honest solution for v1.

## Consequences

Storage must still sit behind an abstraction so future backends can be added later.
Do not implement SQLite in v1 unless explicitly requested.

---

# ADR-0005 — Append-only audit events

* **Status:** accepted
* **Date:** 2026-03-07

## Decision

Mitori stores append-only events for significant task state changes.

## Rationale

The user wants observability and lightweight tracking:

* who/when/what
* movement history
* visibility of change

However, Mitori is not doing full event sourcing.

## Consequences

Events are an audit trail, not the primary source of truth.
The canonical current state remains in tasks and projects.

Examples of event types:

* task_created
* task_updated
* lane_changed
* task_parked
* task_unparked
* task_archived

---

# ADR-0006 — TUI as primary interface

* **Status:** accepted
* **Date:** 2026-03-07

## Decision

The primary interface is a terminal UI built with Bubble Tea.

## Rationale

The user prefers a CLI/TUI application over a web UI.
The application is personal, local, and terminal-native.

Bubble Tea fits:

* board navigation
* keyboard-driven interaction
* multiple screens
* low ceremony local usage

## Consequences

The TUI must remain lightweight and readable.
Do not build a web UI in v1.

---

# ADR-0007 — CLI scope stays thin

* **Status:** accepted
* **Date:** 2026-03-07

## Decision

The CLI should provide only a few thin entry points:

* `mitori`
* `mitori init`
* `mitori add`
* `mitori projects`
* `mitori doctor`

## Rationale

The TUI is the main interface.
The CLI exists for fast entry, initialization, inspection, and maintenance.

## Consequences

Avoid turning the CLI into a large command tree unless there is clear value.

---

# ADR-0008 — Deterministic parsing before AI

* **Status:** accepted
* **Date:** 2026-03-07

## Decision

Structured task parsing in v1 is deterministic and rule-based.

## Rationale

The user wants fast entry such as:

```text
cz - collision playground -> allow fullscreen mode
```

This can be parsed reliably with simple rules.
AI is not necessary for v1 and would add dependency and unpredictability.

## Consequences

Implement a parser that extracts:

* initiative
* project
* title

Type inference may be heuristic or left explicit.
AI classification may be introduced later behind an interface, but must not be required.

---

# ADR-0009 — No productivity theater

* **Status:** accepted
* **Date:** 2026-03-07

## Decision

v1 will not include:

* deadlines
* reminders
* story points
* productivity scoring
* dashboards
* recommendation engines
* forced WIP logic

## Rationale

The user’s goal is observability, not behavioral enforcement.
The system should support manual balancing, not automated pressure.

## Consequences

Reject features that turn Mitori into:

* a KPI tracker
* a habit app
* a gamified planner
* a team PM tool

---

# ADR-0010 — License choice

* **Status:** proposed
* **Date:** 2026-03-07

## Decision

Mitori is likely to use **MPL-2.0**.

## Rationale

The intent is to encourage:

* collaboration
* visible evolution
* reference to the original lineage

The project is an application, not primarily a reusable library.
A slightly protective license fits better than MIT.

## Consequences

Until confirmed, repository scaffolding may refer to the license as pending.
Once confirmed, add:

* `LICENSE`
* README note explaining the collaboration/continuity intent

---

# ADR-0011 — Domain/service/store separation

* **Status:** accepted
* **Date:** 2026-03-07

## Decision

The codebase should separate:

* domain
* service
* store
* tui

## Rationale

This keeps:

* business logic out of UI code
* storage details out of domain code
* future changes easier

## Consequences

Bubble Tea code should dispatch actions and render state.
Business rules belong in services.
Storage logic belongs in store packages.

---

# ADR-0012 — Simplicity over speculative flexibility

* **Status:** accepted
* **Date:** 2026-03-07

## Decision

Mitori will prefer simple, explicit code over abstract or speculative architecture.

## Rationale

This is a personal tool in v1.
The main risk is architectural drift, not scale.

## Consequences

Avoid:

* unnecessary generics
* reflection-heavy patterns
* plugin systems
* broad interface hierarchies
* premature extension points

Introduce abstraction only when it serves a concrete need.

---

# Open Questions

Use this section for unresolved items.

## Q-0001 — Should initiatives remain fixed or become user-defined?

* **Status:** open
* Current default initiatives are:

  * WNT
  * CZ
  * DM

## Q-0002 — Should task type inference be mandatory or optional?

* **Status:** open

## Q-0003 — Should board ordering be purely manual or partially driven by last_touched_at?

* **Status:** open

## Q-0004 — Should archived tasks stay in the main JSON file or move to a separate archive file later?

* **Status:** open

---

# Superseded Decisions

Move replaced decisions here rather than deleting them.

*None yet.*
