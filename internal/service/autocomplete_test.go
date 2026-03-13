package service

import (
	"context"
	"path/filepath"
	"slices"
	"testing"

	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/store"
	"github.com/wastingnotime/mitori/internal/store/file"
)

func TestStructuredValues_IncludeExpectedEnums(t *testing.T) {
	t.Parallel()

	energy := StructuredValues("energy")
	if !slices.Contains(energy, "produces") || !slices.Contains(energy, "neutral") || !slices.Contains(energy, "consumes") {
		t.Fatalf("unexpected energy values: %+v", energy)
	}
	initiatives := StructuredValues("initiative")
	if !slices.Contains(initiatives, string(domain.InitiativeCZ)) {
		t.Fatalf("expected initiative values to include cz: %+v", initiatives)
	}
}

func TestTaskService_StructuredSuggestions_ProjectAndEnums(t *testing.T) {
	t.Parallel()

	st := file.New(filepath.Join(t.TempDir(), "data.json"))
	err := st.Save(context.Background(), store.Snapshot{
		Projects: []domain.Project{
			{ID: "p1", Name: "collision playground", Initiative: domain.InitiativeCZ},
			{ID: "p2", Name: "core shell", Initiative: domain.InitiativeWNT},
		},
	})
	if err != nil {
		t.Fatalf("seed save failed: %v", err)
	}

	svc := NewTaskService(st, NewProjectService(st))

	projectSuggestions, err := svc.StructuredSuggestions(context.Background(), "project", "co")
	if err != nil {
		t.Fatalf("project suggestions failed: %v", err)
	}
	if len(projectSuggestions) != 2 {
		t.Fatalf("expected 2 project suggestions, got %d (%+v)", len(projectSuggestions), projectSuggestions)
	}

	energySuggestions, err := svc.StructuredSuggestions(context.Background(), "energy", "n")
	if err != nil {
		t.Fatalf("energy suggestions failed: %v", err)
	}
	if len(energySuggestions) != 1 || energySuggestions[0] != "neutral" {
		t.Fatalf("expected neutral energy suggestion, got %+v", energySuggestions)
	}
}
