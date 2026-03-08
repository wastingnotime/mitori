package tui

import (
	"fmt"
	"strings"
)

func (m Model) viewQuickAdd() string {
	lines := []string{
		m.styles.header.Render("QUICK ADD"),
		"",
		`Input examples:`,
		`- plain title`,
		`- cz - collision playground -> allow fullscreen mode`,
		"",
		m.quickInput.View(),
	}
	if m.quickErr != "" {
		lines = append(lines, m.styles.error.Render(m.quickErr))
	}
	lines = append(lines, "", m.styles.hint.Render("enter create  esc cancel"))
	return m.styles.app.Render(strings.Join(lines, "\n"))
}

func (m *Model) resetQuickAdd() {
	m.quickErr = ""
	m.quickInput.SetValue("")
}

func (m *Model) setQuickError(err error) {
	if err == nil {
		m.quickErr = ""
		return
	}
	m.quickErr = fmt.Sprintf("error: %v", err)
}
