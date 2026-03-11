package export

import (
	"strings"
	"testing"
	"time"

	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/service"
)

func TestRenderBoardMarkdown_BasicStructure(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	out := RenderBoardMarkdown(BoardExport{
		Type:          "board_export",
		GeneratedAt:   now,
		ActiveFilters: "none",
		LaneCounts: map[domain.Lane]int{
			domain.LaneBacklog: 1,
		},
		Energy: service.BoardEnergySummary{
			Total: service.EnergySummary{Produces: 1, Consumes: 0},
		},
		TasksByLane: map[domain.Lane][]TaskItem{
			domain.LaneBacklog: {
				{Title: "allow fullscreen mode", Project: "collision playground", Initiative: domain.InitiativeCZ, Type: domain.TaskTypeFeat, EnergyType: domain.EnergyProduces, Nature: domain.NatureMushin},
			},
		},
	})

	want := []string{
		"# Mitori Board Export",
		"## BACKLOG (1)",
		"- allow fullscreen mode",
		"- active_filters: none",
	}
	for _, token := range want {
		if !strings.Contains(out, token) {
			t.Fatalf("markdown missing %q\noutput:\n%s", token, out)
		}
	}
}

func TestRenderProjectMarkdown_BasicStructure(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	out := RenderProjectMarkdown(ProjectExport{
		Type:        "project_export",
		GeneratedAt: now,
		Project: domain.Project{
			ID:          "p1",
			Name:        "collision playground",
			Initiative:  domain.InitiativeCZ,
			Description: "physics testbed",
		},
		LaneCounts: map[domain.Lane]int{
			domain.LaneDoing: 1,
		},
		EnergyTotal: service.EnergySummary{Produces: 0, Consumes: 1},
		TasksByLane: map[domain.Lane][]TaskItem{
			domain.LaneDoing: {
				{Title: "fix crash", Type: domain.TaskTypeFix, EnergyType: domain.EnergyConsumes, Nature: domain.NatureGaman},
			},
		},
		RecentEvents: []EventItem{
			{TaskID: "t1", Type: domain.EventTaskCreated, Timestamp: now},
		},
	})

	want := []string{
		"# Mitori Project Export",
		"## collision playground [cz]",
		"### DOING (1)",
		"## Recent Events",
	}
	for _, token := range want {
		if !strings.Contains(out, token) {
			t.Fatalf("markdown missing %q\noutput:\n%s", token, out)
		}
	}
}

func TestRenderTaskMarkdown_BasicStructure(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	out := RenderTaskMarkdown(TaskExport{
		Type:        "task_export",
		GeneratedAt: now,
		Task: TaskItem{
			ID:          "t1",
			Title:       "allow fullscreen mode",
			Description: "more room for testing",
			ProjectID:   "p1",
			Project:     "collision playground",
			Initiative:  domain.InitiativeCZ,
			Type:        domain.TaskTypeFeat,
			Loop:        domain.LoopProduct,
			EnergyType:  domain.EnergyProduces,
			Nature:      domain.NatureMushin,
			Lane:        domain.LaneTodo,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		Events: []EventItem{
			{Type: domain.EventTaskCreated, Timestamp: now},
		},
	})

	want := []string{
		"# Mitori Task Export",
		"## allow fullscreen mode",
		"- energy_type: produces",
		"## Event History",
	}
	for _, token := range want {
		if !strings.Contains(out, token) {
			t.Fatalf("markdown missing %q\noutput:\n%s", token, out)
		}
	}
}

func TestRenderArchiveMarkdown_BasicStructure(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	out := RenderArchiveMarkdown(ArchiveExport{
		Type:        "archive_export",
		GeneratedAt: now,
		GroupMode:   "by_month",
		Groups: []ArchiveGroup{
			{
				Label: "2026-03",
				Tasks: []TaskItem{
					{Title: "finished migration", Project: "collision playground", Initiative: domain.InitiativeCZ, Type: domain.TaskTypeRefact, EnergyType: domain.EnergyConsumes, Nature: domain.NatureShizen},
				},
			},
		},
	})

	want := []string{
		"# Mitori Archive Export",
		"- group_mode: by_month",
		"## 2026-03",
		"- finished migration",
	}
	for _, token := range want {
		if !strings.Contains(out, token) {
			t.Fatalf("markdown missing %q\noutput:\n%s", token, out)
		}
	}
}
