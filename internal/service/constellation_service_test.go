package service

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/store"
	"github.com/wastingnotime/mitori/internal/store/file"
)

func TestBoardService_ProjectConstellation_GroupsSortsAndCounts(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)
	p1 := domain.Project{ID: "p1", Name: "alpha", Initiative: domain.InitiativeWNT}
	p2 := domain.Project{ID: "p2", Name: "zeta", Initiative: domain.InitiativeWNT}
	p3 := domain.Project{ID: "p3", Name: "collision playground", Initiative: domain.InitiativeCZ}
	p4 := domain.Project{ID: "p4", Name: "quiet project", Initiative: domain.InitiativeDM}

	st := file.New(filepath.Join(t.TempDir(), "data.json"))
	mustSaveSnapshot(t, st, store.Snapshot{
		Projects: []domain.Project{p2, p1, p3, p4},
		Tasks: []domain.Task{
			{ID: "t1", ProjectID: "p1", Lane: domain.LaneDoing, CreatedAt: now, UpdatedAt: now.Add(1 * time.Minute), EnergyType: domain.EnergyProduces},
			{ID: "t2", ProjectID: "p1", Lane: domain.LaneTodo, CreatedAt: now, UpdatedAt: now.Add(2 * time.Minute), EnergyType: domain.EnergyConsumes},
			{ID: "t3", ProjectID: "p2", Lane: domain.LaneParking, CreatedAt: now, UpdatedAt: now.Add(3 * time.Minute), EnergyType: domain.EnergyProduces},
			{ID: "t4", ProjectID: "p3", Lane: domain.LaneHalt, CreatedAt: now, UpdatedAt: now.Add(4 * time.Minute), EnergyType: domain.EnergyConsumes},
			{ID: "t5", ProjectID: "p3", Lane: domain.LaneArchived, CreatedAt: now, UpdatedAt: now.Add(5 * time.Minute), EnergyType: domain.EnergyProduces},
		},
	})

	svc := NewBoardService(st)
	out, err := svc.ProjectConstellation(context.Background())
	if err != nil {
		t.Fatalf("ProjectConstellation failed: %v", err)
	}

	if len(out.Groups) != 3 {
		t.Fatalf("expected 3 initiative groups, got %d", len(out.Groups))
	}
	if out.Groups[0].Initiative != domain.InitiativeWNT || out.Groups[1].Initiative != domain.InitiativeCZ || out.Groups[2].Initiative != domain.InitiativeDM {
		t.Fatalf("unexpected initiative order: %+v", out.Groups)
	}

	wntProjects := out.Groups[0].Projects
	if len(wntProjects) != 2 {
		t.Fatalf("expected 2 WNT projects, got %d", len(wntProjects))
	}
	if wntProjects[0].Project.Name != "alpha" || wntProjects[1].Project.Name != "zeta" {
		t.Fatalf("expected alphabetical project order in WNT group, got: %s, %s", wntProjects[0].Project.Name, wntProjects[1].Project.Name)
	}

	alpha := wntProjects[0]
	if alpha.TotalTasks != 2 {
		t.Fatalf("expected alpha total 2, got %d", alpha.TotalTasks)
	}
	if alpha.CountsByLane[domain.LaneDoing] != 1 || alpha.CountsByLane[domain.LaneTodo] != 1 {
		t.Fatalf("unexpected alpha lane counts: %+v", alpha.CountsByLane)
	}
	if alpha.Status != "active" {
		t.Fatalf("expected alpha status active, got %q", alpha.Status)
	}

	czProject := out.Groups[1].Projects[0]
	if czProject.TotalTasks != 2 {
		t.Fatalf("expected CZ project total 2, got %d", czProject.TotalTasks)
	}
	if czProject.CountsByLane[domain.LaneHalt] != 1 || czProject.CountsByLane[domain.LaneArchived] != 1 {
		t.Fatalf("unexpected CZ lane counts: %+v", czProject.CountsByLane)
	}
	if czProject.Status != "halted" {
		t.Fatalf("expected halted status, got %q", czProject.Status)
	}

	quiet := out.Groups[2].Projects[0]
	if quiet.TotalTasks != 0 {
		t.Fatalf("expected quiet project total 0, got %d", quiet.TotalTasks)
	}
	if quiet.Status != "quiet" {
		t.Fatalf("expected quiet status, got %q", quiet.Status)
	}
}

func TestConstellationStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		count map[domain.Lane]int
		want  string
	}{
		{
			name:  "active",
			count: map[domain.Lane]int{domain.LaneDoing: 1, domain.LaneParking: 2, domain.LaneHalt: 1},
			want:  "active",
		},
		{
			name:  "halted",
			count: map[domain.Lane]int{domain.LaneDoing: 0, domain.LaneHalt: 2, domain.LaneParking: 3},
			want:  "halted",
		},
		{
			name:  "parked",
			count: map[domain.Lane]int{domain.LaneDoing: 0, domain.LaneHalt: 0, domain.LaneParking: 2},
			want:  "parked",
		},
		{
			name:  "quiet",
			count: map[domain.Lane]int{domain.LaneDoing: 0, domain.LaneHalt: 0, domain.LaneParking: 0},
			want:  "quiet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := constellationStatus(tt.count); got != tt.want {
				t.Fatalf("status mismatch: got=%q want=%q", got, tt.want)
			}
		})
	}
}

func mustSaveSnapshot(t *testing.T, st *file.Store, snap store.Snapshot) {
	t.Helper()
	if err := st.Save(context.Background(), snap); err != nil {
		t.Fatalf("save snapshot failed: %v", err)
	}
}
