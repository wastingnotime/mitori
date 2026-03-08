package tui

import (
	"fmt"
	"strings"
	"time"
)

func (m Model) viewTaskDetail() string {
	task, ok := m.currentTask()
	if !ok {
		return m.styles.app.Render("No task selected.\n\nPress q to go back.")
	}

	projectName := "-"
	if p, exists := m.projects[task.ProjectID]; exists {
		projectName = p.Name
	}

	var events []string
	for _, evt := range m.recentEvents {
		events = append(events, fmt.Sprintf("- %s  %s", evt.Timestamp.Format(time.RFC3339), evt.Type))
	}
	if len(events) == 0 {
		events = append(events, "- no events")
	}

	lines := []string{
		m.styles.header.Render("TASK DETAIL"),
		"",
		fmt.Sprintf("title: %s", task.Title),
		fmt.Sprintf("description: %s", task.Description),
		fmt.Sprintf("project: %s", projectName),
		fmt.Sprintf("initiative: %s", task.Initiative),
		fmt.Sprintf("type: %s", task.Type),
		fmt.Sprintf("loop: %s", task.Loop),
		fmt.Sprintf("energy: %s", task.EnergyType),
		fmt.Sprintf("nature: %s", task.Nature),
		fmt.Sprintf("lane: %s", task.Lane),
		fmt.Sprintf("created_at: %s", task.CreatedAt.Format(time.RFC3339)),
		fmt.Sprintf("updated_at: %s", task.UpdatedAt.Format(time.RFC3339)),
		fmt.Sprintf("last_touched_at: %s", formatTimePtr(task.LastTouchedAt)),
		fmt.Sprintf("archived_at: %s", formatTimePtr(task.ArchivedAt)),
		"",
		m.styles.sectionTitle.Render("recent events"),
		strings.Join(events, "\n"),
		"",
		m.styles.hint.Render("q back  p park/unpark  t touch  d done  b backlog  x archive"),
	}
	return m.styles.app.Render(strings.Join(lines, "\n"))
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format(time.RFC3339)
}
