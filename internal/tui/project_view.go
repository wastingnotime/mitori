package tui

import (
	"fmt"
	"strings"

	"github.com/wastingnotime/mitori/internal/domain"
)

func (m Model) viewProject() string {
	if m.projectViewID == "" {
		return m.styles.app.Render("No project selected.\n\nPress q to go back.")
	}

	project, ok := m.projects[m.projectViewID]
	if !ok {
		return m.styles.app.Render("Project not found.\n\nPress q to go back.")
	}

	lines := []string{
		m.styles.header.Render("PROJECT VIEW"),
		fmt.Sprintf("%s (%s)", project.Name, project.Initiative),
		"",
	}

	for _, lane := range domain.LaneOrder {
		lines = append(lines, m.styles.sectionTitle.Render(domain.LaneTitle(lane)))
		count := 0
		for _, t := range m.board[lane] {
			if t.ProjectID == project.ID {
				lines = append(lines, "- "+t.Title)
				count++
			}
		}
		if count == 0 {
			lines = append(lines, m.styles.hint.Render("(empty)"))
		}
		lines = append(lines, "")
	}
	lines = append(lines, m.styles.hint.Render("q back"))
	return m.styles.app.Render(strings.Join(lines, "\n"))
}
