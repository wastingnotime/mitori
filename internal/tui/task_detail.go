package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/wastingnotime/mitori/internal/domain"
)

func (m Model) viewTaskDetail() string {
	task, ok := m.currentTask()
	if !ok {
		return m.styles.app.Render("No task selected.\n\nPress Esc to go back.")
	}

	projectName := "-"
	if p, exists := m.projects[task.ProjectID]; exists {
		projectName = p.Name
	}

	var events []string
	for _, evt := range m.recentEvents {
		events = append(events, fmt.Sprintf("- %s  %s", evt.Timestamp.Format(time.RFC3339), formatEvent(evt)))
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
		m.styles.hint.Render(m.actionHint() + "  t:touch  y:confirm  Esc:back  q:quit"),
	}
	return m.styles.app.Render(strings.Join(lines, "\n"))
}

func formatEvent(evt domain.Event) string {
	switch evt.Type {
	case domain.EventLaneChanged:
		return fmt.Sprintf("lane_changed (%v -> %v)", evt.Payload["from"], evt.Payload["to"])
	case domain.EventTaskParked:
		return fmt.Sprintf("task_parked (from %v)", evt.Payload["from"])
	case domain.EventTaskUnparked:
		return fmt.Sprintf("task_unparked (to %v)", evt.Payload["to"])
	case domain.EventTaskArchived:
		if from, ok := evt.Payload["from_lane"]; ok {
			return fmt.Sprintf("task_archived (from %v)", from)
		}
		return "task_archived"
	case domain.EventTaskUpdated:
		if action, ok := evt.Payload["action"]; ok {
			return fmt.Sprintf("task_updated (%v)", action)
		}
		return "task_updated"
	default:
		return string(evt.Type)
	}
}

func formatTimePtr(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format(time.RFC3339)
}
