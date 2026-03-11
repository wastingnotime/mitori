package tui

import (
	"fmt"
	"strings"

	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/service"
)

func (m Model) viewConstellation() string {
	lines := []string{
		m.styles.header.Render("PROJECT CONSTELLATION"),
		m.styles.hint.Render("Projects are the continuity unit. This view shows ecosystem spread at a glance."),
		"",
	}

	totalIdx := 0
	for _, group := range m.constellation.Groups {
		lines = append(lines, m.styles.sectionTitle.Render(strings.ToUpper(string(group.Initiative))))
		for _, entry := range group.Projects {
			prefix := "  "
			if totalIdx == m.constellationIdx {
				prefix = "> "
			}
			lines = append(lines, prefix+renderConstellationProject(entry))
			totalIdx++
		}
		lines = append(lines, "")
	}

	if totalIdx == 0 {
		lines = append(lines, m.styles.hint.Render("(no projects)"))
	}

	project, ok := m.currentConstellationProject()
	if ok {
		lines = append(lines, m.styles.hint.Render("selected: "+project.Project.Name))
	}
	lines = append(lines, m.styles.hint.Render("j/k move  Enter project view  Esc back  q quit"))
	return m.styles.app.Render(strings.Join(lines, "\n"))
}

func renderConstellationProject(p service.ConstellationProject) string {
	summary := fmt.Sprintf(
		"%s  b:%d t:%d d:%d h:%d p:%d dn:%d a:%d  total:%d  status:%s",
		p.Project.Name,
		p.CountsByLane[domain.LaneBacklog],
		p.CountsByLane[domain.LaneTodo],
		p.CountsByLane[domain.LaneDoing],
		p.CountsByLane[domain.LaneHalt],
		p.CountsByLane[domain.LaneParking],
		p.CountsByLane[domain.LaneDone],
		p.CountsByLane[domain.LaneArchived],
		p.TotalTasks,
		p.Status,
	)
	if p.LastActivityAt != nil {
		summary += "  last:" + p.LastActivityAt.Format("2006-01-02")
	}
	return summary
}
