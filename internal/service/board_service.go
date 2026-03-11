package service

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/store"
)

type BoardService struct {
	store store.Store
}

func NewBoardService(st store.Store) *BoardService {
	return &BoardService{store: st}
}

func (s *BoardService) Board(ctx context.Context) (map[domain.Lane][]domain.Task, map[string]domain.Project, error) {
	return s.BoardFiltered(ctx, TaskFilter{}, false)
}

func (s *BoardService) BoardFiltered(ctx context.Context, filter TaskFilter, includeArchived bool) (map[domain.Lane][]domain.Task, map[string]domain.Project, error) {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return nil, nil, err
	}

	projects := make(map[string]domain.Project, len(snap.Projects))
	for _, p := range snap.Projects {
		projects[p.ID] = p
	}

	board := make(map[domain.Lane][]domain.Task, len(domain.LaneOrder))
	for _, lane := range domain.LaneOrder {
		board[lane] = []domain.Task{}
	}
	for _, t := range snap.Tasks {
		if !includeArchived && t.Lane == domain.LaneArchived {
			continue
		}
		projectName := ""
		if p, ok := projects[t.ProjectID]; ok {
			projectName = p.Name
		}
		if !MatchTaskFilter(t, projectName, filter) {
			continue
		}
		board[t.Lane] = append(board[t.Lane], t)
	}

	for lane := range board {
		slices.SortFunc(board[lane], func(a, b domain.Task) int {
			if a.Position != b.Position {
				return a.Position - b.Position
			}
			return strings.Compare(a.ID, b.ID)
		})
	}
	return board, projects, nil
}

type ProjectOverview struct {
	Project       domain.Project
	CountsByLane  map[domain.Lane]int
	EnergyTotal   EnergySummary
	EnergyByLane  map[domain.Lane]EnergySummary
	EnergyNote    string
	RecentEvents  []domain.Event
	TaskTitleByID map[string]string
}

func (s *BoardService) ProjectOverview(ctx context.Context, projectID string, limit int) (ProjectOverview, error) {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return ProjectOverview{}, err
	}

	var project domain.Project
	found := false
	for _, p := range snap.Projects {
		if p.ID == projectID {
			project = p
			found = true
			break
		}
	}
	if !found {
		return ProjectOverview{}, fmt.Errorf("project not found: %s", projectID)
	}

	counts := make(map[domain.Lane]int, len(domain.LaneOrder))
	for _, lane := range domain.LaneOrder {
		counts[lane] = 0
	}
	counts[domain.LaneArchived] = 0
	projectTaskIDs := make(map[string]struct{})
	taskTitles := make(map[string]string)
	projectTasks := make([]domain.Task, 0)
	for _, t := range snap.Tasks {
		if t.ProjectID != projectID {
			continue
		}
		projectTasks = append(projectTasks, t)
		projectTaskIDs[t.ID] = struct{}{}
		taskTitles[t.ID] = t.Title
		if t.Lane != domain.LaneArchived {
			counts[t.Lane]++
		} else {
			counts[domain.LaneArchived]++
		}
	}
	energy := BuildProjectEnergySummary(projectTasks)

	events := make([]domain.Event, 0)
	for _, evt := range snap.Events {
		if _, ok := projectTaskIDs[evt.TaskID]; ok {
			events = append(events, evt)
		}
	}
	slices.SortFunc(events, func(a, b domain.Event) int {
		if a.Timestamp.Equal(b.Timestamp) {
			return strings.Compare(a.ID, b.ID)
		}
		if a.Timestamp.After(b.Timestamp) {
			return -1
		}
		return 1
	})
	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	return ProjectOverview{
		Project:       project,
		CountsByLane:  counts,
		EnergyTotal:   energy.Total,
		EnergyByLane:  energy.ByLane,
		EnergyNote:    energy.Note,
		RecentEvents:  events,
		TaskTitleByID: taskTitles,
	}, nil
}
