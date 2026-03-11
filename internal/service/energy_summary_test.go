package service

import (
	"testing"

	"github.com/wastingnotime/mitori/internal/domain"
)

func TestBuildBoardEnergySummary_Counts(t *testing.T) {
	board := map[domain.Lane][]domain.Task{
		domain.LaneTodo: {
			{EnergyType: domain.EnergyProduces},
			{EnergyType: domain.EnergyConsumes},
		},
		domain.LaneDoing: {
			{EnergyType: domain.EnergyProduces},
		},
		domain.LaneParking: {
			{EnergyType: domain.EnergyConsumes},
			{EnergyType: domain.EnergyConsumes},
		},
		domain.LaneHalt: {
			{EnergyType: domain.EnergyConsumes},
		},
		domain.LaneDone: {
			{EnergyType: domain.EnergyProduces},
		},
	}

	got := BuildBoardEnergySummary(board)
	if got.Total.Produces != 3 || got.Total.Consumes != 4 {
		t.Fatalf("unexpected total energy summary: %+v", got.Total)
	}
	if got.ByLane[domain.LaneDoing].Produces != 1 || got.ByLane[domain.LaneDoing].Consumes != 0 {
		t.Fatalf("unexpected doing lane summary: %+v", got.ByLane[domain.LaneDoing])
	}
	if got.Note == "" {
		t.Fatal("expected descriptive note for consume-heavy active lanes")
	}
}

func TestBuildProjectEnergySummary_CountsByLane(t *testing.T) {
	tasks := []domain.Task{
		{Lane: domain.LaneTodo, EnergyType: domain.EnergyConsumes},
		{Lane: domain.LaneDoing, EnergyType: domain.EnergyProduces},
		{Lane: domain.LaneDoing, EnergyType: domain.EnergyProduces},
		{Lane: domain.LaneDone, EnergyType: domain.EnergyConsumes},
	}

	got := BuildProjectEnergySummary(tasks)
	if got.Total.Produces != 2 || got.Total.Consumes != 2 {
		t.Fatalf("unexpected project total: %+v", got.Total)
	}
	doing := got.ByLane[domain.LaneDoing]
	if doing.Produces != 2 || doing.Consumes != 0 {
		t.Fatalf("unexpected doing summary: %+v", doing)
	}
}

func TestBuildBoardEnergySummary_FilteredSubset(t *testing.T) {
	filteredBoard := map[domain.Lane][]domain.Task{
		domain.LaneTodo: {
			{EnergyType: domain.EnergyProduces},
		},
		domain.LaneDoing: {
			{EnergyType: domain.EnergyProduces},
		},
	}

	got := BuildBoardEnergySummary(filteredBoard)
	if got.Total.Produces != 2 || got.Total.Consumes != 0 {
		t.Fatalf("unexpected filtered summary: %+v", got.Total)
	}
	if got.Note != "" {
		t.Fatalf("expected no warning note for produce-heavy subset, got=%q", got.Note)
	}
}
