package service

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/wastingnotime/mitori/internal/domain"
)

type ConstellationProject struct {
	Project        domain.Project
	CountsByLane   map[domain.Lane]int
	TotalTasks     int
	Status         string
	LastActivityAt *time.Time
}

type ConstellationGroup struct {
	Initiative domain.Initiative
	Projects   []ConstellationProject
}

type ProjectConstellation struct {
	Groups []ConstellationGroup
}

func (s *BoardService) ProjectConstellation(ctx context.Context) (ProjectConstellation, error) {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return ProjectConstellation{}, err
	}

	entries := make(map[string]ConstellationProject, len(snap.Projects))
	for _, project := range snap.Projects {
		counts := make(map[domain.Lane]int, len(domain.AllLanes))
		for _, lane := range domain.AllLanes {
			counts[lane] = 0
		}
		entries[project.ID] = ConstellationProject{
			Project:      project,
			CountsByLane: counts,
			Status:       "quiet",
		}
	}

	for _, task := range snap.Tasks {
		entry, ok := entries[task.ProjectID]
		if !ok {
			continue
		}
		if domain.ValidLane(task.Lane) {
			entry.CountsByLane[task.Lane]++
		}
		entry.TotalTasks++
		entry.Status = constellationStatus(entry.CountsByLane)
		entry.LastActivityAt = latestActivity(entry.LastActivityAt, task)
		entries[task.ProjectID] = entry
	}

	grouped := make(map[domain.Initiative][]ConstellationProject)
	for _, entry := range entries {
		grouped[entry.Project.Initiative] = append(grouped[entry.Project.Initiative], entry)
	}

	seen := make(map[domain.Initiative]struct{}, len(grouped))
	groups := make([]ConstellationGroup, 0, len(grouped))
	for _, initiative := range domain.InitiativeValues {
		projects := grouped[initiative]
		if len(projects) == 0 {
			continue
		}
		slices.SortFunc(projects, compareConstellationProjects)
		groups = append(groups, ConstellationGroup{
			Initiative: initiative,
			Projects:   projects,
		})
		seen[initiative] = struct{}{}
	}

	extra := make([]domain.Initiative, 0)
	for initiative := range grouped {
		if _, ok := seen[initiative]; !ok {
			extra = append(extra, initiative)
		}
	}
	slices.SortFunc(extra, func(a, b domain.Initiative) int {
		return strings.Compare(string(a), string(b))
	})
	for _, initiative := range extra {
		projects := grouped[initiative]
		slices.SortFunc(projects, compareConstellationProjects)
		groups = append(groups, ConstellationGroup{
			Initiative: initiative,
			Projects:   projects,
		})
	}

	return ProjectConstellation{Groups: groups}, nil
}

func compareConstellationProjects(a, b ConstellationProject) int {
	an := strings.ToLower(strings.TrimSpace(a.Project.Name))
	bn := strings.ToLower(strings.TrimSpace(b.Project.Name))
	if an != bn {
		return strings.Compare(an, bn)
	}
	return strings.Compare(a.Project.ID, b.Project.ID)
}

func constellationStatus(counts map[domain.Lane]int) string {
	if counts[domain.LaneDoing] > 0 {
		return "active"
	}
	if counts[domain.LaneHalt] > 0 {
		return "halted"
	}
	if counts[domain.LaneParking] > 0 {
		return "parked"
	}
	return "quiet"
}

func latestActivity(current *time.Time, task domain.Task) *time.Time {
	candidates := make([]time.Time, 0, 4)
	candidates = append(candidates, task.CreatedAt, task.UpdatedAt)
	if task.LastTouchedAt != nil {
		candidates = append(candidates, *task.LastTouchedAt)
	}
	if task.ArchivedAt != nil {
		candidates = append(candidates, *task.ArchivedAt)
	}

	var latest time.Time
	for _, ts := range candidates {
		if ts.IsZero() {
			continue
		}
		if latest.IsZero() || ts.After(latest) {
			latest = ts
		}
	}
	if latest.IsZero() {
		return current
	}
	if current == nil || latest.After(*current) {
		cp := latest
		return &cp
	}
	return current
}
