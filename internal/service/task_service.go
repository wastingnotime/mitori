package service

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/store"
)

type CreateTaskInput struct {
	Title       string
	Description string
	Initiative  domain.Initiative
	ProjectName string
	Type        domain.TaskType
	Loop        domain.Loop
	EnergyType  domain.EnergyType
	Nature      domain.Nature
	Lane        domain.Lane
}

type TaskService struct {
	store      store.Store
	projectSvc *ProjectService
}

func NewTaskService(st store.Store, projectSvc *ProjectService) *TaskService {
	return &TaskService{
		store:      st,
		projectSvc: projectSvc,
	}
}

func (s *TaskService) Create(ctx context.Context, in CreateTaskInput) (domain.Task, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return domain.Task{}, fmt.Errorf("title is required")
	}

	initiative := in.Initiative
	if !domain.ValidInitiative(initiative) {
		initiative = domain.InitiativeWNT
	}

	project, err := s.projectSvc.Ensure(ctx, initiative, in.ProjectName)
	if err != nil {
		return domain.Task{}, err
	}

	if !domain.ValidTaskType(in.Type) {
		in.Type = domain.TaskTypeChore
	}
	if !domain.ValidLoop(in.Loop) {
		in.Loop = domain.LoopProduct
	}
	if !domain.ValidEnergyType(in.EnergyType) {
		in.EnergyType = domain.EnergyProduces
	}
	if !domain.ValidNature(in.Nature) {
		in.Nature = domain.NatureMushin
	}
	if !domain.ValidLane(in.Lane) {
		in.Lane = domain.LaneBacklog
	}

	snap, err := s.store.Load(ctx)
	if err != nil {
		return domain.Task{}, err
	}

	position := 1
	for _, t := range snap.Tasks {
		if t.Lane == in.Lane && t.ArchivedAt == nil && t.Position >= position {
			position = t.Position + 1
		}
	}

	now := time.Now().UTC()
	task := domain.Task{
		ID:          newID("tsk"),
		Title:       title,
		Description: strings.TrimSpace(in.Description),
		Initiative:  initiative,
		ProjectID:   project.ID,
		Type:        in.Type,
		Loop:        in.Loop,
		EnergyType:  in.EnergyType,
		Nature:      in.Nature,
		Lane:        in.Lane,
		Position:    position,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	snap.Tasks = append(snap.Tasks, task)
	snap.Events = append(snap.Events, domain.Event{
		ID:        newID("evt"),
		TaskID:    task.ID,
		Type:      domain.EventTaskCreated,
		Timestamp: now,
	})
	if err := s.store.Save(ctx, snap); err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

func (s *TaskService) ListActive(ctx context.Context) ([]domain.Task, error) {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Task, 0, len(snap.Tasks))
	for _, t := range snap.Tasks {
		if t.ArchivedAt == nil {
			out = append(out, t)
		}
	}
	slices.SortFunc(out, func(a, b domain.Task) int {
		if a.Lane != b.Lane {
			return laneIndex(a.Lane) - laneIndex(b.Lane)
		}
		if a.Position != b.Position {
			return a.Position - b.Position
		}
		return strings.Compare(a.ID, b.ID)
	})
	return out, nil
}

func (s *TaskService) ListByProjectID(ctx context.Context, projectID string) ([]domain.Task, error) {
	tasks, err := s.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Task, 0)
	for _, t := range tasks {
		if t.ProjectID == projectID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (s *TaskService) Get(ctx context.Context, taskID string) (domain.Task, bool, error) {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return domain.Task{}, false, err
	}
	for _, t := range snap.Tasks {
		if t.ID == taskID {
			return t, true, nil
		}
	}
	return domain.Task{}, false, nil
}

func (s *TaskService) MoveLane(ctx context.Context, taskID string, lane domain.Lane) error {
	if !domain.ValidLane(lane) {
		return fmt.Errorf("invalid lane: %s", lane)
	}
	return s.updateTask(ctx, taskID, func(t *domain.Task, now time.Time) (domain.EventType, map[string]interface{}, bool) {
		old := t.Lane
		if old == lane {
			return "", nil, false
		}
		t.Lane = lane
		t.UpdatedAt = now
		return domain.EventLaneChanged, map[string]interface{}{
			"from": old,
			"to":   lane,
		}, true
	})
}

func (s *TaskService) ToggleParking(ctx context.Context, taskID string) error {
	return s.updateTask(ctx, taskID, func(t *domain.Task, now time.Time) (domain.EventType, map[string]interface{}, bool) {
		if t.Lane == domain.LaneParking {
			t.Lane = domain.LaneTodo
			t.UpdatedAt = now
			return domain.EventTaskUnparked, map[string]interface{}{"to": domain.LaneTodo}, true
		}
		from := t.Lane
		t.Lane = domain.LaneParking
		t.UpdatedAt = now
		return domain.EventTaskParked, map[string]interface{}{"from": from}, true
	})
}

func (s *TaskService) Touch(ctx context.Context, taskID string) error {
	return s.updateTask(ctx, taskID, func(t *domain.Task, now time.Time) (domain.EventType, map[string]interface{}, bool) {
		t.UpdatedAt = now
		t.LastTouchedAt = &now
		return domain.EventTaskUpdated, map[string]interface{}{"action": "touch"}, true
	})
}

func (s *TaskService) MarkDone(ctx context.Context, taskID string) error {
	return s.MoveLane(ctx, taskID, domain.LaneDone)
}

func (s *TaskService) SendBacklog(ctx context.Context, taskID string) error {
	return s.MoveLane(ctx, taskID, domain.LaneBacklog)
}

func (s *TaskService) Archive(ctx context.Context, taskID string) error {
	return s.updateTask(ctx, taskID, func(t *domain.Task, now time.Time) (domain.EventType, map[string]interface{}, bool) {
		if t.ArchivedAt != nil {
			return "", nil, false
		}
		t.ArchivedAt = &now
		t.UpdatedAt = now
		return domain.EventTaskArchived, nil, true
	})
}

func (s *TaskService) RecentEvents(ctx context.Context, taskID string, limit int) ([]domain.Event, error) {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Event, 0)
	for _, evt := range snap.Events {
		if evt.TaskID == taskID {
			out = append(out, evt)
		}
	}
	slices.SortFunc(out, func(a, b domain.Event) int {
		if a.Timestamp.Equal(b.Timestamp) {
			return strings.Compare(a.ID, b.ID)
		}
		if a.Timestamp.Before(b.Timestamp) {
			return -1
		}
		return 1
	})
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out, nil
}

type taskMutator func(task *domain.Task, now time.Time) (eventType domain.EventType, payload map[string]interface{}, changed bool)

func (s *TaskService) updateTask(ctx context.Context, taskID string, mut taskMutator) error {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	for i := range snap.Tasks {
		if snap.Tasks[i].ID != taskID {
			continue
		}
		evtType, payload, changed := mut(&snap.Tasks[i], now)
		if !changed {
			return nil
		}
		if evtType != "" {
			snap.Events = append(snap.Events, domain.Event{
				ID:        newID("evt"),
				TaskID:    taskID,
				Type:      evtType,
				Timestamp: now,
				Payload:   payload,
			})
		}
		return s.store.Save(ctx, snap)
	}
	return fmt.Errorf("task not found: %s", taskID)
}

func laneIndex(lane domain.Lane) int {
	for i, l := range domain.LaneOrder {
		if l == lane {
			return i
		}
	}
	return len(domain.LaneOrder)
}
