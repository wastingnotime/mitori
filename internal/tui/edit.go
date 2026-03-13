package tui

import "strings"

func (m Model) viewEdit() string {
	fields := make([]string, 0, len(editFieldKeys()))
	current := m.editFieldKey()
	for _, key := range editFieldKeys() {
		prefix := "  "
		if key == current {
			prefix = "> "
		}
		value := m.currentEditValueByKey(key)
		fields = append(fields, prefix+editFieldLabel(key)+": "+value)
	}

	lines := []string{
		m.styles.header.Render("EDIT TASK"),
		"",
		"Edit is allowed only for tasks in BACKLOG.",
		"",
		strings.Join(fields, "\n"),
		"",
		m.styles.sectionTitle.Render("edit value"),
		m.editInput.View(),
		"",
		m.styles.hint.Render("tab/shift+tab field  ctrl+n autocomplete  enter save  esc cancel (double Esc if unsaved)"),
	}
	return m.styles.app.Render(strings.Join(lines, "\n"))
}

func (m Model) currentEditValueByKey(key string) string {
	switch key {
	case "title":
		return m.editDraft.Title
	case "description":
		return m.editDraft.Description
	case "project":
		return m.editDraft.ProjectName
	case "initiative":
		return string(m.editDraft.Initiative)
	case "type":
		return string(m.editDraft.Type)
	case "loop":
		return string(m.editDraft.Loop)
	case "energy":
		return string(m.editDraft.EnergyType)
	case "nature":
		return string(m.editDraft.Nature)
	default:
		return ""
	}
}
