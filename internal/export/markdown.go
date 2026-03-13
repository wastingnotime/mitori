package export

import (
	"fmt"
	"strings"

	"github.com/wastingnotime/mitori/internal/domain"
)

func RenderBoardMarkdown(data BoardExport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Mitori Board Export\n\n")
	fmt.Fprintf(&b, "- generated_at: %s\n", data.GeneratedAt.Format("2006-01-02 15:04:05 MST"))
	fmt.Fprintf(&b, "- active_filters: %s\n", data.ActiveFilters)
	fmt.Fprintf(&b, "- energy total: produces %d, neutral %d, consumes %d\n", data.Energy.Total.Produces, data.Energy.Total.Neutral, data.Energy.Total.Consumes)
	if data.Energy.Note != "" {
		fmt.Fprintf(&b, "- note: %s\n", data.Energy.Note)
	}
	fmt.Fprintf(&b, "\n")

	for _, lane := range domain.LaneOrder {
		fmt.Fprintf(&b, "## %s (%d)\n", domain.LaneTitle(lane), data.LaneCounts[lane])
		tasks := data.TasksByLane[lane]
		if len(tasks) == 0 {
			fmt.Fprintf(&b, "- (empty)\n\n")
			continue
		}
		for _, task := range tasks {
			fmt.Fprintf(&b, "- %s\n", task.Title)
			fmt.Fprintf(&b, "  - project: %s\n", task.Project)
			fmt.Fprintf(&b, "  - meta: %s · %s · %s · %s\n", task.Initiative, task.Type, task.EnergyType, task.Nature)
		}
		fmt.Fprintf(&b, "\n")
	}
	return b.String()
}

func RenderProjectMarkdown(data ProjectExport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Mitori Project Export\n\n")
	fmt.Fprintf(&b, "## %s [%s]\n\n", data.Project.Name, data.Project.Initiative)
	fmt.Fprintf(&b, "- project_id: %s\n", data.Project.ID)
	fmt.Fprintf(&b, "- generated_at: %s\n", data.GeneratedAt.Format("2006-01-02 15:04:05 MST"))
	if strings.TrimSpace(data.Project.Description) != "" {
		fmt.Fprintf(&b, "- description: %s\n", data.Project.Description)
	}
	fmt.Fprintf(&b, "- energy total: produces %d, neutral %d, consumes %d\n", data.EnergyTotal.Produces, data.EnergyTotal.Neutral, data.EnergyTotal.Consumes)
	if data.EnergyNote != "" {
		fmt.Fprintf(&b, "- note: %s\n", data.EnergyNote)
	}
	fmt.Fprintf(&b, "\n")

	for _, lane := range domain.LaneOrder {
		fmt.Fprintf(&b, "### %s (%d)\n", domain.LaneTitle(lane), data.LaneCounts[lane])
		tasks := data.TasksByLane[lane]
		if len(tasks) == 0 {
			fmt.Fprintf(&b, "- (empty)\n\n")
			continue
		}
		for _, task := range tasks {
			fmt.Fprintf(&b, "- %s\n", task.Title)
			fmt.Fprintf(&b, "  - type: %s | energy: %s | nature: %s\n", task.Type, task.EnergyType, task.Nature)
		}
		fmt.Fprintf(&b, "\n")
	}

	if len(data.Archived) > 0 {
		fmt.Fprintf(&b, "### ARCHIVED (%d)\n", len(data.Archived))
		for _, task := range data.Archived {
			fmt.Fprintf(&b, "- %s\n", task.Title)
		}
		fmt.Fprintf(&b, "\n")
	}

	fmt.Fprintf(&b, "## Recent Events\n")
	if len(data.RecentEvents) == 0 {
		fmt.Fprintf(&b, "- (none)\n")
		return b.String()
	}
	for _, evt := range data.RecentEvents {
		fmt.Fprintf(&b, "- %s  %s  (task: %s)\n", evt.Timestamp.Format("2006-01-02 15:04"), evt.Type, evt.TaskID)
	}
	return b.String()
}

func RenderTaskMarkdown(data TaskExport) string {
	task := data.Task
	var b strings.Builder
	fmt.Fprintf(&b, "# Mitori Task Export\n\n")
	fmt.Fprintf(&b, "## %s\n\n", task.Title)
	fmt.Fprintf(&b, "- task_id: %s\n", task.ID)
	fmt.Fprintf(&b, "- generated_at: %s\n", data.GeneratedAt.Format("2006-01-02 15:04:05 MST"))
	fmt.Fprintf(&b, "- description: %s\n", task.Description)
	fmt.Fprintf(&b, "- project: %s (%s)\n", task.Project, task.ProjectID)
	fmt.Fprintf(&b, "- initiative: %s\n", task.Initiative)
	fmt.Fprintf(&b, "- type: %s\n", task.Type)
	fmt.Fprintf(&b, "- loop: %s\n", task.Loop)
	fmt.Fprintf(&b, "- energy_type: %s\n", task.EnergyType)
	fmt.Fprintf(&b, "- nature: %s\n", task.Nature)
	fmt.Fprintf(&b, "- lane: %s\n", task.Lane)
	fmt.Fprintf(&b, "- created_at: %s\n", task.CreatedAt.Format("2006-01-02 15:04:05 MST"))
	fmt.Fprintf(&b, "- updated_at: %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05 MST"))
	if task.ArchivedAt != nil {
		fmt.Fprintf(&b, "- archived_at: %s\n", task.ArchivedAt.Format("2006-01-02 15:04:05 MST"))
	}
	fmt.Fprintf(&b, "\n## Event History\n")
	if len(data.Events) == 0 {
		fmt.Fprintf(&b, "- (none)\n")
		return b.String()
	}
	for _, evt := range data.Events {
		fmt.Fprintf(&b, "- %s  %s\n", evt.Timestamp.Format("2006-01-02 15:04"), evt.Type)
	}
	return b.String()
}

func RenderArchiveMarkdown(data ArchiveExport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Mitori Archive Export\n\n")
	fmt.Fprintf(&b, "- generated_at: %s\n", data.GeneratedAt.Format("2006-01-02 15:04:05 MST"))
	fmt.Fprintf(&b, "- group_mode: %s\n\n", data.GroupMode)
	if len(data.Groups) == 0 {
		fmt.Fprintf(&b, "- (empty archive)\n")
		return b.String()
	}

	for _, g := range data.Groups {
		fmt.Fprintf(&b, "## %s\n", g.Label)
		for _, task := range g.Tasks {
			fmt.Fprintf(&b, "- %s\n", task.Title)
			fmt.Fprintf(&b, "  - project: %s\n", task.Project)
			fmt.Fprintf(&b, "  - meta: %s · %s · %s · %s\n", task.Initiative, task.Type, task.EnergyType, task.Nature)
		}
		fmt.Fprintf(&b, "\n")
	}
	return b.String()
}
