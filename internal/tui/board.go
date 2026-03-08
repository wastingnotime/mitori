package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/wastingnotime/mitori/internal/domain"
)

func (m Model) viewBoard() string {
	header := m.styles.header.Render("MITORI v0.2  Observability board")
	filterLine := m.styles.hint.Render("filters: " + m.filterLabel)

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
			cardWidth := max(16, m.width/6-6)
			line1 := truncateRight(strings.TrimSpace(t.Title), cardWidth)
			line2 := truncateRight(fmt.Sprintf("%s · %s · %s", projectName, t.Initiative, t.EnergyType), cardWidth)
			card := line1 + "\n" + m.styles.hint.Render(line2)
			if idx == selected && i == m.laneIdx {
				lines = append(lines, m.styles.cardActive.Render(card))
			} else {
				lines = append(lines, m.styles.card.Render(card))
			}
			lines = append(lines, "")
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
	hint := m.styles.hint.Render("h/l lanes  j/k tasks  J/K reorder  y confirm  t touch  / filters  A archive  a add  g project  ? help  q quit")
	actionHint := m.styles.hint.Render(m.actionHint())
	board := lipgloss.JoinHorizontal(lipgloss.Top, columns...)
	return m.styles.app.Render(lipgloss.JoinVertical(lipgloss.Left, header, filterLine, "", board, "", status, actionHint, hint))
}

func truncateRight(s string, maxLen int) string {
	if maxLen <= 0 || len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
