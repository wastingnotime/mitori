package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/wastingnotime/mitori/internal/domain"
)

func (m Model) viewBoard() string {
	header := m.styles.header.Render("MITORI v1.0  Observability board")

	columns := make([]string, 0, len(domain.LaneOrder))
	for i, lane := range domain.LaneOrder {
		tasks := m.board[lane]
		lines := []string{
			fmt.Sprintf("%s (%d)", domain.LaneTitle(lane), len(tasks)),
			"",
		}

		selected := m.selectedCardIndex(lane)
		for idx, t := range tasks {
			projectName := "-"
			if p, ok := m.projects[t.ProjectID]; ok {
				projectName = p.Name
			}
			line := fmt.Sprintf("%s | %s | %s | %s", t.Title, projectName, t.Initiative, t.EnergyType)
			if idx == selected && i == m.laneIdx {
				lines = append(lines, m.styles.cardActive.Render(line))
			} else {
				lines = append(lines, m.styles.card.Render(line))
			}
		}
		if len(tasks) == 0 {
			lines = append(lines, m.styles.hint.Render("(empty)"))
		}

		colStyle := m.styles.column
		if i == m.laneIdx {
			colStyle = m.styles.columnActive
		}
		columns = append(columns, colStyle.Width(max(18, m.width/6-2)).Render(strings.Join(lines, "\n")))
	}

	status := m.styles.status.Render(m.status)
	hint := m.styles.hint.Render("h/l lanes  j/k tasks  enter detail  a add  g project  ? help  q quit")
	board := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	return m.styles.app.Render(lipgloss.JoinVertical(lipgloss.Left, header, "", board, "", status, hint))
}
