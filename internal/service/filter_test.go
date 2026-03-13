package service

import (
	"testing"

	"github.com/wastingnotime/mitori/internal/domain"
)

func TestMatchTaskFilter(t *testing.T) {
	task := domain.Task{
		ID:          "tsk_1",
		Title:       "Allow fullscreen mode",
		Description: "Fix collision viewport clipping",
		Initiative:  domain.InitiativeCZ,
		ProjectID:   "prj_1",
		Type:        domain.TaskTypeFix,
		EnergyType:  domain.EnergyProduces,
	}
	projectName := "Collision Playground"

	tests := []struct {
		name   string
		filter TaskFilter
		want   bool
	}{
		{"match_by_project_id", TaskFilter{ProjectID: "prj_1"}, true},
		{"mismatch_by_project_id", TaskFilter{ProjectID: "prj_x"}, false},
		{"match_by_initiative", TaskFilter{Initiative: domain.InitiativeCZ}, true},
		{"mismatch_by_initiative", TaskFilter{Initiative: domain.InitiativeWNT}, false},
		{"match_by_energy", TaskFilter{EnergyType: domain.EnergyProduces}, true},
		{"mismatch_by_energy", TaskFilter{EnergyType: domain.EnergyConsumes}, false},
		{"mismatch_by_energy_neutral", TaskFilter{EnergyType: domain.EnergyNeutral}, false},
		{"match_by_task_type", TaskFilter{TaskType: domain.TaskTypeFix}, true},
		{"mismatch_by_task_type", TaskFilter{TaskType: domain.TaskTypeFeat}, false},
		{"match_by_query_title", TaskFilter{Query: "fullscreen"}, true},
		{"match_by_query_description", TaskFilter{Query: "viewport"}, true},
		{"match_by_query_project", TaskFilter{Query: "playground"}, true},
		{"mismatch_by_query", TaskFilter{Query: "unrelated"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchTaskFilter(task, projectName, tt.filter)
			if got != tt.want {
				t.Fatalf("match mismatch: got=%v want=%v", got, tt.want)
			}
		})
	}
}
