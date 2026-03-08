# Mitori — Ground Rules for AI Coding Sessions

## Purpose

Mitori is a **local-first terminal Kanban application** for **personal work observability**.

It is **not** a productivity platform.
It is **not** a team workflow system.
It is **not** a SaaS product.

The purpose of Mitori is to help the user **see work clearly and steer manually**.

---

## Product Principles

### 1. Observability first
The primary value of Mitori is visibility.

If a feature does not improve:
- clarity
- readability
- state awareness
- manual decision support

then it is probably out of scope.

---

### 2. Manual steering over automation
Mitori should help the user **decide**, not decide for the user.

Do not introduce:
- automatic prioritization
- enforced balancing
- recommendation engines
- hidden workflow automation

unless explicitly requested later.

---

### 3. Projects are the true balancing unit
Initiative is only a classifier.

The real working container is:

```text
project
````

The user may work across many projects in parallel.
This is intentional and must not be treated as an anti-pattern.

Do not impose “single active project” assumptions.

---

### 4. Simplicity over cleverness

Prefer:

* boring code
* explicit data flow
* simple state transitions
* minimal abstraction

Avoid speculative architecture.

---

### 5. Local-first always

Mitori must work fully offline.

Do not introduce:

* server dependencies
* cloud sync
* remote APIs
* background daemons

unless explicitly requested.

---

## Scope of v1.0

v1.0 includes only:

* projects
* tasks
* lanes
* parking
* audit events
* JSON file storage
* CLI entry points
* Bubble Tea TUI
* deterministic task parsing

v1.0 does **not** include:

* AI classification
* deadlines
* reminders
* acceptance criteria
* metrics dashboards
* multi-user support
* sync
* plugin systems

---

## Board Semantics

The board lanes are:

```text
backlog
todo
doing
halt
parking
done
```

### Meanings

**backlog**
High-level intents or tasks not yet ready.

**todo**
Tasks ready to start.

**doing**
Tasks currently being worked on.

**halt**
Tasks blocked by external factors.

**parking**
Tasks intentionally paused because attention shifted.

**done**
Finished tasks awaiting final review or archival.

Do not rename these lanes unless explicitly asked.

---

## Domain Model Rules

### Task

A task is the central entity.

A task must support:

* title
* description
* initiative
* project reference
* task type
* loop
* energy type
* nature
* lane
* timestamps
* archive state

### Project

A project is the true unit of continuity.

A project must:

* belong to an initiative
* group related tasks
* remain visible in the UI

### Event

Events are append-only audit records.

Use them to record meaningful state changes, especially:

* creation
* updates
* lane changes
* park/unpark
* archive

Do not overcomplicate events into full event sourcing.

Events are an audit trail, not the primary source of truth.

---

## Storage Rules

v1 uses a single JSON file.

Requirements:

* human-inspectable
* easy to back up
* stable format
* minimal surprise

Storage must sit behind an interface so other backends can be added later.

Do not implement SQLite in v1 unless explicitly requested.

Do not prematurely optimize persistence.

---

## TUI Rules

The TUI is the primary interface.

Use:

* Bubble Tea
* Bubbles
* Lip Gloss

### TUI goals

* fast navigation
* clear project visibility
* readable task cards
* low cognitive load
* keyboard-first interaction

### TUI non-goals

* flashy animation
* decorative complexity
* overloaded dashboards
* dense enterprise-style panels

The board must be understandable at a glance.

---

## CLI Rules

The CLI should remain thin and useful.

Required commands:

* `mitori`
* `mitori init`
* `mitori add`
* `mitori projects`
* `mitori doctor`

Do not add many subcommands unless clearly useful.

---

## Parsing Rules

Structured task input must be deterministic.

Example:

```text
cz - collision playground -> allow fullscreen mode
```

Expected parse:

* initiative = cz
* project = collision playground
* title = allow fullscreen mode

Start with simple parsing rules.
Do not introduce AI parsing in v1.

---

## AI Rules

AI support is not part of v1 behavior.

If AI-related code is prepared, it must be:

* optional
* interface-based
* disabled by default
* non-blocking to the rest of the app

Do not make the app depend on any model provider.

---

## Architectural Rules

### Keep business logic out of TUI code

Bubble Tea code should render state and dispatch actions.

Business rules belong in services.

### Keep domain free from UI concerns

Domain types must not import Bubble Tea or storage details.

### Keep storage isolated

Persistence logic should not leak into domain or TUI code.

### Prefer explicit services

Use small service methods for:

* creating tasks
* moving tasks
* parking tasks
* touching tasks
* listing board state
* listing project state

---

## Code Style Rules

Prefer:

* small files
* explicit names
* straightforward functions
* low magic
* low coupling

Avoid:

* unnecessary generics
* reflection
* meta-framework patterns
* premature plugin architecture
* giant interfaces

Use interfaces only where they provide real separation:

* storage backend
* optional classifier
* clock/time abstraction if useful for tests

---

## Testing Rules

Focus tests on:

* parser behavior
* lane transitions
* task lifecycle
* storage roundtrip
* event recording

Do not overinvest in TUI snapshot tests in v1.

Prioritize confidence in:

* domain behavior
* services
* persistence

---

## UX Rules

Mitori should feel like:

* a studio instrument
* a workbench
* a board you can read quickly

It should **not** feel like:

* Jira
* Trello clone with bloat
* personal KPI tracker
* habit/productivity gamification tool

No gamification.

No fake urgency.

No pressure mechanics.

---

## Open Source Intent

Mitori is an open-source application intended for:

* collaboration
* visible evolution
* referenceable lineage

The code should be readable and welcoming to contributors.

Avoid obscure architecture that makes contribution harder.

---

## Decision Heuristic

When uncertain, choose the option that maximizes:

1. clarity
2. local simplicity
3. manual control
4. project visibility
5. future maintainability

Do not choose based on:

* trendiness
* abstraction purity
* hypothetical scale
* imagined enterprise use cases

---

## Anti-Drift Rules

Do not gradually transform Mitori into:

* a scheduler
* a planner
* a recommendation engine
* an analytics dashboard
* a collaboration platform

Mitori is a **Kanban observability tool**.

Protect that identity.

---

## If You Need to Extend the System

Before adding a new concept, ask:

1. Does this improve observability?
2. Does this preserve manual steering?
3. Does this fit local-first usage?
4. Is this necessary for v1?
5. Is there a simpler version?

If the answer is unclear, do not add it.

---

## Final Rule

Prefer a smaller honest tool over a larger confused one.
