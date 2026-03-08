package file

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/store"
)

func TestStore_SaveLoadRoundtrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "data.json")
	st := New(path)

	now := time.Now().UTC().Truncate(time.Second)
	archivedAt := now.Add(1 * time.Hour)
	snap := store.Snapshot{
		Projects: []domain.Project{
			{
				ID:         "prj_1",
				Name:       "collision playground",
				Initiative: domain.InitiativeCZ,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
		},
		Tasks: []domain.Task{
			{
				ID:         "tsk_1",
				Title:      "allow fullscreen mode",
				Initiative: domain.InitiativeCZ,
				ProjectID:  "prj_1",
				Type:       domain.TaskTypeFeat,
				Loop:       domain.LoopProduct,
				EnergyType: domain.EnergyProduces,
				Nature:     domain.NatureMushin,
				Lane:       domain.LaneArchived,
				Position:   0,
				CreatedAt:  now,
				UpdatedAt:  now,
				ArchivedAt: &archivedAt,
			},
		},
		Events: []domain.Event{
			{
				ID:        "evt_1",
				TaskID:    "tsk_1",
				Type:      domain.EventTaskCreated,
				Timestamp: now,
			},
		},
	}

	if err := st.Save(ctx, snap); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	got, err := st.Load(ctx)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if len(got.Projects) != 1 || got.Projects[0].Name != "collision playground" {
		t.Fatalf("unexpected projects roundtrip: %+v", got.Projects)
	}
	if len(got.Tasks) != 1 {
		t.Fatalf("unexpected tasks length: %d", len(got.Tasks))
	}
	task := got.Tasks[0]
	if task.Lane != domain.LaneArchived {
		t.Fatalf("expected archived lane, got=%q", task.Lane)
	}
	if task.ArchivedAt == nil || !task.ArchivedAt.Equal(archivedAt) {
		t.Fatalf("expected archived timestamp preserved, got=%v", task.ArchivedAt)
	}
	if len(got.Events) != 1 || got.Events[0].ID != "evt_1" {
		t.Fatalf("unexpected events roundtrip: %+v", got.Events)
	}
}
