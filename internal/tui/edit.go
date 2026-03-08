package tui

import "strings"

func (m Model) viewEdit() string {
	lines := []string{
		m.styles.header.Render("EDIT TASK"),
		"",
		"Edit is allowed only for tasks in BACKLOG.",
		"",
		m.editInput.View(),
		"",
		m.styles.hint.Render("enter save  esc cancel"),
	}
	return m.styles.app.Render(strings.Join(lines, "\n"))
}
