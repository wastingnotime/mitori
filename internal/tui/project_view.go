package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/wastingnotime/mitori/internal/domain"
)

func (m Model) viewProject() string {
	if m.projectViewID == "" {
		return m.styles.app.Render("No project selected.\n\nPress q to go back.")
	}

	project := m.overview.Project
	if project.ID == "" {
		return m.styles.app.Render("Project not found.\n\nPress q to go back.")
	}

	lines := []string{
		m.styles.header.Render("PROJECT VIEW"),
		fmt.Sprintf("%s  [%s]", project.Name, project.Initiative),
		fmt.Sprintf("project_id: %s", project.ID),
		"",
		m.styles.sectionTitle.Render("overview"),
	}

	for _, lane := range domain.LaneOrder {
		lines = append(lines, fmt.Sprintf("- %s: %d", strings.ToLower(domain.LaneTitle(lane)), m.overview.CountsByLane[lane]))
	}

	lines = append(lines, "", m.styles.sectionTitle.Render("tasks by lane"))
	for _, lane := range domain.LaneOrder {
		lines = append(lines, m.styles.sectionTitle.Render(domain.LaneTitle(lane)))
		count := 0
		for _, t := range m.board[lane] {
			if t.ProjectID == project.ID {
				lines = append(lines, fmt.Sprintf("- %s [%s]", t.Title, t.Type))
				count++
			}
		}
		if count == 0 {
			lines = append(lines, m.styles.hint.Render("(empty)"))
		}
		lines = append(lines, "")
	}

	lines = append(lines, m.styles.sectionTitle.Render("recent activity"))
	if len(m.overview.RecentEvents) == 0 {
		lines = append(lines, m.styles.hint.Render("(no recent events)"))
	} else {
		for _, evt := range m.overview.RecentEvents {
			title := m.overview.TaskTitleByID[evt.TaskID]
			lines = append(lines, fmt.Sprintf("- %s  %s  %s", evt.Timestamp.Format(time.RFC3339), title, formatEvent(evt)))
		}
	}
	lines = append(lines, "", m.styles.hint.Render("q back"))
	return m.styles.app.Render(strings.Join(lines, "\n"))
}
