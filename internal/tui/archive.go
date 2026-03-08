package tui

import (
	"fmt"
	"strings"
)

func (m Model) viewArchive() string {
	lines := []string{
		m.styles.header.Render("ARCHIVE"),
		"",
	}

	if len(m.archiveTasks) == 0 {
		lines = append(lines, m.styles.hint.Render("No archived tasks match the current filters."))
	} else {
		for idx, t := range m.archiveTasks {
			project := "-"
			if p, ok := m.projects[t.ProjectID]; ok {
				project = p.Name
			}
			archivedAt := "-"
			if t.ArchivedAt != nil {
				archivedAt = t.ArchivedAt.Format("2006-01-02 15:04")
			}
			line := fmt.Sprintf("%s | %s | %s | %s | archived %s", t.Title, project, t.Initiative, t.Type, archivedAt)
			if idx == m.archiveIdx {
				lines = append(lines, m.styles.cardActive.Render(line))
			} else {
				lines = append(lines, "- "+line)
			}
		}
		if task, ok := m.currentArchiveTask(); ok {
			lines = append(lines, "", m.styles.hint.Render(m.actionHintForActions(m.taskActions(task))))
		}
	}

	lines = append(lines, "", m.styles.status.Render(m.status))
	lines = append(lines, m.styles.hint.Render("j/k select  y confirm  Esc back/cancel  / search/filter  c clear filters"))
	return m.styles.app.Render(strings.Join(lines, "\n"))
}

func (m Model) viewSearch() string {
	lines := []string{
		m.styles.header.Render("FILTERS"),
		"",
		"Type plain text to search over title, description, and project name.",
		"Token formats supported:",
		`project turtle   initiative cz   energy produces   type discovery`,
		`project:"collision playground"   initiative:cz   energy:produces   type:feat`,
		"",
		m.searchInput.View(),
		"",
		m.styles.hint.Render("enter apply  esc cancel"),
	}
	return m.styles.app.Render(strings.Join(lines, "\n"))
}
