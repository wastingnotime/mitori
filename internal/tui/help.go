package tui

import "strings"

func (m Model) viewHelp() string {
	lines := []string{
		m.styles.header.Render("HELP"),
		"",
		"h/l  move between lanes",
		"j/k  move within lane",
		"J/K  reorder task within lane",
		"enter open task",
		"",
		"a    add task",
		"y    confirm pending recovery/destructive action",
		"t    touch task",
		"",
		"g    project view",
		"/    search/filter input",
		"A    archive browser",
		"c    clear active filters",
		"?    help",
		"Esc  back/cancel/close",
		"q    quit app",
		"",
		m.styles.hint.Render("Esc back"),
	}
	return m.styles.app.Render(strings.Join(lines, "\n"))
}
