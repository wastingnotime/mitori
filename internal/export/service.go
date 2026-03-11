package export

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/service"
	"github.com/wastingnotime/mitori/internal/store"
)

type Service struct {
	store store.Store
}

func NewService(st store.Store) *Service {
	return &Service{store: st}
}

func (s *Service) BuildBoard(ctx context.Context) (BoardExport, error) {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return BoardExport{}, err
	}

	projects := indexProjects(snap.Projects)
	tasksByLane := make(map[domain.Lane][]TaskItem, len(domain.LaneOrder))
	laneCounts := make(map[domain.Lane]int, len(domain.LaneOrder))
	board := make(map[domain.Lane][]domain.Task, len(domain.LaneOrder))
	for _, lane := range domain.LaneOrder {
		tasksByLane[lane] = []TaskItem{}
		laneCounts[lane] = 0
		board[lane] = []domain.Task{}
	}

	for _, task := range snap.Tasks {
		if task.Lane == domain.LaneArchived {
			continue
		}
		project := projects[task.ProjectID]
		item := toTaskItem(task, project.Name)
		tasksByLane[task.Lane] = append(tasksByLane[task.Lane], item)
		laneCounts[task.Lane]++
		board[task.Lane] = append(board[task.Lane], task)
	}
	for lane := range tasksByLane {
		sortTaskItems(tasksByLane[lane])
	}

	return BoardExport{
		Type:          "board_export",
		GeneratedAt:   time.Now().UTC(),
		ActiveFilters: "none",
		LaneCounts:    laneCounts,
		Energy:        service.BuildBoardEnergySummary(board),
		TasksByLane:   tasksByLane,
	}, nil
}

func (s *Service) BuildProject(ctx context.Context, projectRef string) (ProjectExport, error) {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return ProjectExport{}, err
	}

	project, found := findProject(snap.Projects, projectRef)
	if !found {
		return ProjectExport{}, fmt.Errorf("unknown project: %s", projectRef)
	}

	tasksByLane := make(map[domain.Lane][]TaskItem, len(domain.LaneOrder))
	laneCounts := make(map[domain.Lane]int, len(domain.LaneOrder))
	archived := make([]TaskItem, 0)
	for _, lane := range domain.LaneOrder {
		tasksByLane[lane] = []TaskItem{}
		laneCounts[lane] = 0
	}

	projectTasks := make([]domain.Task, 0)
	taskIDs := make(map[string]struct{})
	for _, task := range snap.Tasks {
		if task.ProjectID != project.ID {
			continue
		}
		taskIDs[task.ID] = struct{}{}
		projectTasks = append(projectTasks, task)
		item := toTaskItem(task, project.Name)
		if task.Lane == domain.LaneArchived {
			archived = append(archived, item)
			continue
		}
		tasksByLane[task.Lane] = append(tasksByLane[task.Lane], item)
		laneCounts[task.Lane]++
	}
	for lane := range tasksByLane {
		sortTaskItems(tasksByLane[lane])
	}
	sortTaskItems(archived)

	energy := service.BuildProjectEnergySummary(projectTasks)
	events := collectProjectEvents(snap.Events, taskIDs, 40)
	return ProjectExport{
		Type:         "project_export",
		GeneratedAt:  time.Now().UTC(),
		Project:      project,
		LaneCounts:   laneCounts,
		EnergyTotal:  energy.Total,
		EnergyByLane: energy.ByLane,
		EnergyNote:   energy.Note,
		TasksByLane:  tasksByLane,
		Archived:     archived,
		RecentEvents: events,
	}, nil
}

func (s *Service) BuildTask(ctx context.Context, taskID string) (TaskExport, error) {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return TaskExport{}, err
	}

	task, found := findTask(snap.Tasks, taskID)
	if !found {
		return TaskExport{}, fmt.Errorf("unknown task: %s", taskID)
	}
	projects := indexProjects(snap.Projects)
	projectName := ""
	if p, ok := projects[task.ProjectID]; ok {
		projectName = p.Name
	}

	events := make([]EventItem, 0)
	for _, evt := range snap.Events {
		if evt.TaskID == task.ID {
			events = append(events, EventItem{
				ID:        evt.ID,
				TaskID:    evt.TaskID,
				Type:      evt.Type,
				Timestamp: evt.Timestamp,
				Payload:   evt.Payload,
			})
		}
	}
	slices.SortFunc(events, func(a, b EventItem) int {
		if a.Timestamp.Equal(b.Timestamp) {
			return strings.Compare(a.ID, b.ID)
		}
		if a.Timestamp.Before(b.Timestamp) {
			return -1
		}
		return 1
	})

	return TaskExport{
		Type:        "task_export",
		GeneratedAt: time.Now().UTC(),
		Task:        toTaskItem(task, projectName),
		Events:      events,
	}, nil
}

func (s *Service) BuildArchive(ctx context.Context) (ArchiveExport, error) {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return ArchiveExport{}, err
	}
	projects := indexProjects(snap.Projects)

	groupsByMonth := map[string][]TaskItem{}
	groupsByProject := map[string][]TaskItem{}
	allWithMonth := true

	for _, task := range snap.Tasks {
		if task.Lane != domain.LaneArchived {
			continue
		}
		projectName := ""
		if p, ok := projects[task.ProjectID]; ok {
			projectName = p.Name
		}
		item := toTaskItem(task, projectName)
		if task.ArchivedAt != nil {
			month := task.ArchivedAt.Format("2006-01")
			groupsByMonth[month] = append(groupsByMonth[month], item)
		} else {
			allWithMonth = false
			key := projectName
			if key == "" {
				key = "unknown-project"
			}
			groupsByProject[key] = append(groupsByProject[key], item)
		}
	}

	mode := "by_month"
	groupMap := groupsByMonth
	if !allWithMonth {
		mode = "by_project"
		groupMap = groupsByProject
	}

	labels := make([]string, 0, len(groupMap))
	for label := range groupMap {
		labels = append(labels, label)
	}
	slices.Sort(labels)

	groups := make([]ArchiveGroup, 0, len(labels))
	for _, label := range labels {
		items := groupMap[label]
		sortTaskItems(items)
		groups = append(groups, ArchiveGroup{
			Label: label,
			Tasks: items,
		})
	}

	return ArchiveExport{
		Type:        "archive_export",
		GeneratedAt: time.Now().UTC(),
		GroupMode:   mode,
		Groups:      groups,
	}, nil
}

func indexProjects(projects []domain.Project) map[string]domain.Project {
	out := make(map[string]domain.Project, len(projects))
	for _, p := range projects {
		out[p.ID] = p
	}
	return out
}

func findProject(projects []domain.Project, ref string) (domain.Project, bool) {
	ref = strings.TrimSpace(strings.ToLower(ref))
	for _, p := range projects {
		if strings.ToLower(p.ID) == ref || strings.ToLower(p.Name) == ref {
			return p, true
		}
	}
	return domain.Project{}, false
}

func findTask(tasks []domain.Task, id string) (domain.Task, bool) {
	for _, task := range tasks {
		if task.ID == id {
			return task, true
		}
	}
	return domain.Task{}, false
}

func toTaskItem(task domain.Task, projectName string) TaskItem {
	return TaskItem{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		ProjectID:   task.ProjectID,
		Project:     projectName,
		Initiative:  task.Initiative,
		Type:        task.Type,
		Loop:        task.Loop,
		EnergyType:  task.EnergyType,
		Nature:      task.Nature,
		Lane:        task.Lane,
		Position:    task.Position,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
		ArchivedAt:  task.ArchivedAt,
	}
}

func sortTaskItems(items []TaskItem) {
	slices.SortFunc(items, func(a, b TaskItem) int {
		if a.Position != b.Position {
			return a.Position - b.Position
		}
		return strings.Compare(a.ID, b.ID)
	})
}

func collectProjectEvents(events []domain.Event, taskIDs map[string]struct{}, limit int) []EventItem {
	out := make([]EventItem, 0)
	for _, evt := range events {
		if _, ok := taskIDs[evt.TaskID]; !ok {
			continue
		}
		out = append(out, EventItem{
			ID:        evt.ID,
			TaskID:    evt.TaskID,
			Type:      evt.Type,
			Timestamp: evt.Timestamp,
			Payload:   evt.Payload,
		})
	}
	slices.SortFunc(out, func(a, b EventItem) int {
		if a.Timestamp.Equal(b.Timestamp) {
			return strings.Compare(a.ID, b.ID)
		}
		if a.Timestamp.After(b.Timestamp) {
			return -1
		}
		return 1
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}
