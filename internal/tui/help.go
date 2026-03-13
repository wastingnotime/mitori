package tui

import (
	"strings"
)

func (m Model) viewHelp() string {
	lines := m.helpLines()
	viewportHeight := max(10, m.height-6)
	if m.helpScroll > max(0, len(lines)-viewportHeight) {
		m.helpScroll = max(0, len(lines)-viewportHeight)
	}
	start := m.helpScroll
	end := min(len(lines), start+viewportHeight)
	body := strings.Join(lines[start:end], "\n")

	footer := m.styles.hint.Render("j/k scroll  Esc back/cancel  q quit")
	title := m.styles.header.Render("HELP + ABOUT")
	return m.styles.app.Render(strings.Join([]string{title, "", body, "", footer}, "\n"))
}

func (m Model) helpLines() []string {
	lines := []string{
		m.styles.sectionTitle.Render("About Mitori"),
		"Mitori is a local-first terminal Kanban for personal work observability.",
		"It exists to make work visible so you can steer manually.",
		"It is not a productivity tracker, team PM tool, or planning SaaS.",
		"The name points to seeing a situation clearly at a glance and grasping its structure.",
		"Projects are the true continuity unit; use constellation view to read the full ecosystem.",
		"",
		m.styles.sectionTitle.Render("Lane Meanings"),
		"- backlog: high-level intents and refinement queue",
		"- todo: ready to start",
		"- doing: active work",
		"- halt: externally blocked",
		"- parking: intentionally paused",
		"- done: finished and pending archive/review",
		"- archived: closed work retained for history",
		"",
		m.styles.sectionTitle.Render("Key Semantics"),
		"- m: move forward in the happy path",
		"- b: move backward / reopen",
		"- s: halt / resume halt",
		"- p: park / resume parked work",
		"- x: archive boundary",
		"- q: quit",
		"- Esc: back/cancel",
		"",
		m.styles.sectionTitle.Render("Transition Semantics"),
		"- transitions are lane-scoped and validated in domain/service",
		"- recovery transitions require confirmation (y)",
		"- destructive boundaries require confirmation where defined",
		"- edit/delete are backlog-only actions",
		"",
		m.styles.sectionTitle.Render("Available Commands"),
		"- h/l: move between board lanes",
		"- j/k: move within lane or scroll help/archive list",
		"- J/K: reorder within lane",
		"- t: touch task",
		"- lane actions: shown dynamically for selected task",
		"- a: quick add (up/down history, ctrl+n autocomplete)",
		"- g: project view",
		"- C: project constellation view",
		"- /: filter input",
		"- c: clear filters",
		"- A: archive browser",
		"- ?: open this Help + About screen",
	}
	return lines
}
