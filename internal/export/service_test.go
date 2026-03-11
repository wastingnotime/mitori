package export

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/store"
)

type stubStore struct {
	snap store.Snapshot
}

func (s *stubStore) Init(context.Context) error { return nil }
func (s *stubStore) Load(context.Context) (store.Snapshot, error) {
	return s.snap, nil
}
func (s *stubStore) Save(context.Context, store.Snapshot) error { return nil }
func (s *stubStore) Path() string                               { return "" }

func TestService_BuildBoard(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	svc := NewService(&stubStore{
		snap: store.Snapshot{
			Projects: []domain.Project{
				{ID: "p1", Name: "collision playground", Initiative: domain.InitiativeCZ},
			},
			Tasks: []domain.Task{
				{ID: "t1", Title: "task backlog", ProjectID: "p1", Initiative: domain.InitiativeCZ, Type: domain.TaskTypeFeat, EnergyType: domain.EnergyProduces, Nature: domain.NatureMushin, Lane: domain.LaneBacklog, Position: 2, CreatedAt: now, UpdatedAt: now},
				{ID: "t2", Title: "task doing", ProjectID: "p1", Initiative: domain.InitiativeCZ, Type: domain.TaskTypeFix, EnergyType: domain.EnergyConsumes, Nature: domain.NatureGaman, Lane: domain.LaneDoing, Position: 1, CreatedAt: now, UpdatedAt: now},
				{ID: "t3", Title: "task archived", ProjectID: "p1", Initiative: domain.InitiativeCZ, Type: domain.TaskTypeChore, EnergyType: domain.EnergyProduces, Nature: domain.NatureShizen, Lane: domain.LaneArchived, Position: 0, CreatedAt: now, UpdatedAt: now},
			},
		},
	})

	out, err := svc.BuildBoard(context.Background())
	if err != nil {
		t.Fatalf("BuildBoard failed: %v", err)
	}
	if out.Type != "board_export" {
		t.Fatalf("unexpected export type: %q", out.Type)
	}
	if out.LaneCounts[domain.LaneBacklog] != 1 || out.LaneCounts[domain.LaneDoing] != 1 {
		t.Fatalf("unexpected lane counts: %+v", out.LaneCounts)
	}
	if out.LaneCounts[domain.LaneArchived] != 0 {
		t.Fatalf("archived tasks should not be counted in board lanes: %+v", out.LaneCounts)
	}
	if out.Energy.Total.Produces != 1 || out.Energy.Total.Consumes != 1 {
		t.Fatalf("unexpected board energy total: %+v", out.Energy.Total)
	}
	if got := len(out.TasksByLane[domain.LaneArchived]); got != 0 {
		t.Fatalf("expected no archived tasks in board export, got %d", got)
	}
}

func TestService_BuildProject(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	archivedAt := now.Add(2 * time.Hour)
	svc := NewService(&stubStore{
		snap: store.Snapshot{
			Projects: []domain.Project{
				{ID: "p1", Name: "collision playground", Initiative: domain.InitiativeCZ},
				{ID: "p2", Name: "other project", Initiative: domain.InitiativeDM},
			},
			Tasks: []domain.Task{
				{ID: "t1", Title: "todo task", ProjectID: "p1", Initiative: domain.InitiativeCZ, Type: domain.TaskTypeFeat, EnergyType: domain.EnergyProduces, Nature: domain.NatureMushin, Lane: domain.LaneTodo, Position: 0, CreatedAt: now, UpdatedAt: now},
				{ID: "t2", Title: "doing task", ProjectID: "p1", Initiative: domain.InitiativeCZ, Type: domain.TaskTypeFix, EnergyType: domain.EnergyConsumes, Nature: domain.NatureGaman, Lane: domain.LaneDoing, Position: 1, CreatedAt: now, UpdatedAt: now},
				{ID: "t3", Title: "archived task", ProjectID: "p1", Initiative: domain.InitiativeCZ, Type: domain.TaskTypeChore, EnergyType: domain.EnergyProduces, Nature: domain.NatureShizen, Lane: domain.LaneArchived, Position: 2, CreatedAt: now, UpdatedAt: now, ArchivedAt: &archivedAt},
				{ID: "t4", Title: "foreign task", ProjectID: "p2", Initiative: domain.InitiativeDM, Type: domain.TaskTypeFeat, EnergyType: domain.EnergyProduces, Nature: domain.NatureMushin, Lane: domain.LaneTodo, Position: 0, CreatedAt: now, UpdatedAt: now},
			},
			Events: []domain.Event{
				{ID: "e1", TaskID: "t2", Type: domain.EventLaneChanged, Timestamp: now.Add(1 * time.Minute)},
				{ID: "e2", TaskID: "t1", Type: domain.EventTaskCreated, Timestamp: now},
				{ID: "e3", TaskID: "t4", Type: domain.EventTaskCreated, Timestamp: now.Add(2 * time.Minute)},
			},
		},
	})

	out, err := svc.BuildProject(context.Background(), "collision playground")
	if err != nil {
		t.Fatalf("BuildProject failed: %v", err)
	}
	if out.Type != "project_export" {
		t.Fatalf("unexpected export type: %q", out.Type)
	}
	if out.Project.ID != "p1" {
		t.Fatalf("unexpected project selected: %+v", out.Project)
	}
	if out.LaneCounts[domain.LaneTodo] != 1 || out.LaneCounts[domain.LaneDoing] != 1 {
		t.Fatalf("unexpected lane counts: %+v", out.LaneCounts)
	}
	if len(out.Archived) != 1 {
		t.Fatalf("expected one archived task in project export, got %d", len(out.Archived))
	}
	if out.EnergyTotal.Produces != 2 || out.EnergyTotal.Consumes != 1 {
		t.Fatalf("unexpected project energy total: %+v", out.EnergyTotal)
	}
	if len(out.RecentEvents) != 2 {
		t.Fatalf("expected project-scoped events only, got %d", len(out.RecentEvents))
	}
}

func TestService_BuildTask(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	svc := NewService(&stubStore{
		snap: store.Snapshot{
			Projects: []domain.Project{
				{ID: "p1", Name: "collision playground", Initiative: domain.InitiativeCZ},
			},
			Tasks: []domain.Task{
				{ID: "t1", Title: "task one", Description: "detail", ProjectID: "p1", Initiative: domain.InitiativeCZ, Type: domain.TaskTypeFeat, Loop: domain.LoopProduct, EnergyType: domain.EnergyProduces, Nature: domain.NatureMushin, Lane: domain.LaneDoing, Position: 0, CreatedAt: now, UpdatedAt: now},
			},
			Events: []domain.Event{
				{ID: "e2", TaskID: "t1", Type: domain.EventTaskUpdated, Timestamp: now.Add(2 * time.Minute)},
				{ID: "e1", TaskID: "t1", Type: domain.EventTaskCreated, Timestamp: now},
			},
		},
	})

	out, err := svc.BuildTask(context.Background(), "t1")
	if err != nil {
		t.Fatalf("BuildTask failed: %v", err)
	}
	if out.Type != "task_export" {
		t.Fatalf("unexpected export type: %q", out.Type)
	}
	if out.Task.Project != "collision playground" {
		t.Fatalf("unexpected task project name: %q", out.Task.Project)
	}
	if len(out.Events) != 2 {
		t.Fatalf("unexpected events length: %d", len(out.Events))
	}
	if out.Events[0].ID != "e1" || out.Events[1].ID != "e2" {
		t.Fatalf("expected events sorted ascending by time, got: %+v", out.Events)
	}
}

func TestService_BuildArchive(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	a1 := now.Add(-24 * time.Hour)
	a2 := now.Add(-48 * time.Hour)
	svc := NewService(&stubStore{
		snap: store.Snapshot{
			Projects: []domain.Project{
				{ID: "p1", Name: "collision playground", Initiative: domain.InitiativeCZ},
			},
			Tasks: []domain.Task{
				{ID: "t1", Title: "archived one", ProjectID: "p1", Initiative: domain.InitiativeCZ, Type: domain.TaskTypeFeat, EnergyType: domain.EnergyProduces, Nature: domain.NatureMushin, Lane: domain.LaneArchived, Position: 1, CreatedAt: now, UpdatedAt: now, ArchivedAt: &a1},
				{ID: "t2", Title: "archived two", ProjectID: "p1", Initiative: domain.InitiativeCZ, Type: domain.TaskTypeFix, EnergyType: domain.EnergyConsumes, Nature: domain.NatureGaman, Lane: domain.LaneArchived, Position: 2, CreatedAt: now, UpdatedAt: now, ArchivedAt: &a2},
			},
		},
	})

	out, err := svc.BuildArchive(context.Background())
	if err != nil {
		t.Fatalf("BuildArchive failed: %v", err)
	}
	if out.Type != "archive_export" {
		t.Fatalf("unexpected export type: %q", out.Type)
	}
	if out.GroupMode != "by_month" {
		t.Fatalf("expected by_month grouping, got: %s", out.GroupMode)
	}
	if len(out.Groups) == 0 {
		t.Fatal("expected at least one archive group")
	}
}

func TestService_NotFoundErrors(t *testing.T) {
	t.Parallel()

	svc := NewService(&stubStore{snap: store.Snapshot{}})

	if _, err := svc.BuildProject(context.Background(), "unknown"); err == nil || !strings.Contains(err.Error(), "unknown project") {
		t.Fatalf("expected unknown project error, got: %v", err)
	}
	if _, err := svc.BuildTask(context.Background(), "unknown"); err == nil || !strings.Contains(err.Error(), "unknown task") {
		t.Fatalf("expected unknown task error, got: %v", err)
	}
}
