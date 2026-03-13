package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/service"
)

func (m Model) viewProject() string {
	if m.projectViewID == "" {
		return m.styles.app.Render("No project selected.\n\nPress Esc to go back.")
	}

	project := m.overview.Project
	if project.ID == "" {
		return m.styles.app.Render("Project not found.\n\nPress Esc to go back.")
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

	lines = append(lines, "", m.styles.sectionTitle.Render("energy"))
	lines = append(lines, fmt.Sprintf("- total produces: %d", m.overview.EnergyTotal.Produces))
	lines = append(lines, fmt.Sprintf("- total neutral: %d", m.overview.EnergyTotal.Neutral))
	lines = append(lines, fmt.Sprintf("- total consumes: %d", m.overview.EnergyTotal.Consumes))
	for _, lane := range []domain.Lane{domain.LaneTodo, domain.LaneDoing, domain.LaneParking, domain.LaneHalt, domain.LaneDone} {
		entry := m.overview.EnergyByLane[lane]
		lines = append(lines, fmt.Sprintf("- %s: produces %d, neutral %d, consumes %d", strings.ToLower(domain.LaneTitle(lane)), entry.Produces, entry.Neutral, entry.Consumes))
	}
	if m.overview.EnergyNote != "" {
		lines = append(lines, m.styles.hint.Render("note: "+m.overview.EnergyNote))
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
			lines = append(lines, fmt.Sprintf("- %s  %s  %s", evt.Timestamp.Format(time.RFC3339), title, service.FormatEvent(evt)))
		}
	}
	lines = append(lines, "", m.styles.hint.Render("Esc back"))
	return m.styles.app.Render(strings.Join(lines, "\n"))
}
