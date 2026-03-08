package tui

import "github.com/charmbracelet/lipgloss"

type styles struct {
	app          lipgloss.Style
	header       lipgloss.Style
	column       lipgloss.Style
	columnActive lipgloss.Style
	card         lipgloss.Style
	cardActive   lipgloss.Style
	status       lipgloss.Style
	hint         lipgloss.Style
	error        lipgloss.Style
	sectionTitle lipgloss.Style
}

func defaultStyles() styles {
	return styles{
		app:          lipgloss.NewStyle().Padding(1, 1),
		header:       lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")),
		column:       lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1),
		columnActive: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("86")).Padding(0, 1),
		card:         lipgloss.NewStyle().Padding(0, 0).Foreground(lipgloss.Color("252")),
		cardActive:   lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("63")).Padding(0, 1),
		status:       lipgloss.NewStyle().Foreground(lipgloss.Color("121")),
		hint:         lipgloss.NewStyle().Foreground(lipgloss.Color("244")),
		error:        lipgloss.NewStyle().Foreground(lipgloss.Color("203")),
		sectionTitle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("141")),
	}
}
