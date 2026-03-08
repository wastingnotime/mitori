package tui

import "strings"

func (m Model) viewHelp() string {
	lines := []string{
		m.styles.header.Render("HELP"),
		"",
		"h/l  move between lanes",
		"j/k  move within lane",
		"enter open task",
		"",
		"a    add task",
		"e    edit task (not implemented in v1)",
		"m    move task to next lane",
		"",
		"p    park/unpark task",
		"t    touch task",
		"d    mark done",
		"b    send to backlog",
		"x    archive task",
		"",
		"g    project view",
		"/    filter (not implemented in v1)",
		"?    help",
		"q    quit/back",
		"",
		m.styles.hint.Render("q back"),
	}
	return m.styles.app.Render(strings.Join(lines, "\n"))
}
