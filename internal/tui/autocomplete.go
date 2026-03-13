package tui

import (
	"fmt"
	"strings"
)

func (m Model) autocompleteEditField() (string, bool) {
	field := m.editFieldKey()
	switch field {
	case "project", "initiative", "type", "loop", "energy", "nature":
	default:
		return "", false
	}
	queryField := field
	if field == "energy" {
		queryField = "energy_type"
	}
	suggestions, err := m.taskSvc.StructuredSuggestions(m.ctx, queryField, m.editInput.Value())
	if err != nil || len(suggestions) == 0 {
		return "", false
	}
	return suggestions[0], true
}

func (m Model) autocompleteQuickAdd() (string, bool) {
	raw := m.quickInput.Value()
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}

	if !strings.Contains(trimmed, " - ") {
		suggestions, err := m.taskSvc.StructuredSuggestions(m.ctx, "initiative", trimmed)
		if err != nil || len(suggestions) == 0 {
			return "", false
		}
		return suggestions[0], true
	}

	if strings.Contains(trimmed, "->") {
		parts := strings.SplitN(trimmed, "->", 2)
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		leftParts := strings.SplitN(left, " - ", 2)
		if len(leftParts) != 2 || strings.TrimSpace(right) != "" {
			return "", false
		}
		initiative := strings.TrimSpace(leftParts[0])
		projectPrefix := strings.TrimSpace(leftParts[1])
		suggestions, err := m.taskSvc.StructuredSuggestions(m.ctx, "project", projectPrefix)
		if err != nil || len(suggestions) == 0 {
			return "", false
		}
		return fmt.Sprintf("%s - %s -> ", initiative, suggestions[0]), true
	}

	parts := strings.SplitN(trimmed, " - ", 3)
	if len(parts) == 2 {
		initiative := strings.TrimSpace(parts[0])
		projectPrefix := strings.TrimSpace(parts[1])
		suggestions, err := m.taskSvc.StructuredSuggestions(m.ctx, "project", projectPrefix)
		if err != nil || len(suggestions) == 0 {
			return "", false
		}
		return fmt.Sprintf("%s - %s - ", initiative, suggestions[0]), true
	}
	return "", false
}
