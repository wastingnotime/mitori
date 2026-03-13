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
		`- cz - collision playground - allow fullscreen mode`,
		`- cz - collision playground -> allow fullscreen mode`,
		"",
		m.quickInput.View(),
	}
	if m.quickErr != "" {
		lines = append(lines, m.styles.error.Render(m.quickErr))
	}
	lines = append(lines, "", m.styles.hint.Render("up/down history  ctrl+n autocomplete  enter create  esc cancel"))
	return m.styles.app.Render(strings.Join(lines, "\n"))
}

func (m *Model) resetQuickAdd() {
	m.quickErr = ""
	m.quickInput.SetValue("")
	m.quickHistory.browsing = false
	m.quickHistory.idx = len(m.quickHistory.items)
	m.quickHistory.draft = ""
}

func (m *Model) setQuickError(err error) {
	if err == nil {
		m.quickErr = ""
		return
	}
	m.quickErr = fmt.Sprintf("error: %v", err)
}
