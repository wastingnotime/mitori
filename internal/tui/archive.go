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
		for _, t := range m.archiveTasks {
			project := "-"
			if p, ok := m.projects[t.ProjectID]; ok {
				project = p.Name
			}
			archivedAt := "-"
			if t.ArchivedAt != nil {
				archivedAt = t.ArchivedAt.Format("2006-01-02 15:04")
			}
			lines = append(lines, fmt.Sprintf("- %s | %s | %s | %s | archived %s", t.Title, project, t.Initiative, t.Type, archivedAt))
		}
	}

	lines = append(lines, "", m.styles.hint.Render("q back  / search/filter  c clear filters"))
	return m.styles.app.Render(strings.Join(lines, "\n"))
}

func (m Model) viewSearch() string {
	lines := []string{
		m.styles.header.Render("FILTERS"),
		"",
		"Type plain text to search over title, description, and project name.",
		"Optional deterministic tokens:",
		`project:"<name>"  initiative:<wnt|cz|dm>  energy:<produces|consumes>  type:<discovery|feat|refact|chore|fix>`,
		"",
		m.searchInput.View(),
		"",
		m.styles.hint.Render("enter apply  esc cancel  c clear"),
	}
	return m.styles.app.Render(strings.Join(lines, "\n"))
}
