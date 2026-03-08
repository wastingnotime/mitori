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
		Position:    nextLanePosition(snap, in.Lane),
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
	sortTasks(out)
	return out, nil
}

func (s *TaskService) ListArchivedFiltered(ctx context.Context, filter TaskFilter) ([]domain.Task, map[string]domain.Project, error) {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return nil, nil, err
	}
	projects := mapProjects(snap.Projects)
	out := make([]domain.Task, 0)
	for _, t := range snap.Tasks {
		if t.ArchivedAt == nil {
			continue
		}
		projectName := ""
		if p, ok := projects[t.ProjectID]; ok {
			projectName = p.Name
		}
		if MatchTaskFilter(t, projectName, filter) {
			out = append(out, t)
		}
	}
	slices.SortFunc(out, func(a, b domain.Task) int {
		if a.ArchivedAt != nil && b.ArchivedAt != nil && !a.ArchivedAt.Equal(*b.ArchivedAt) {
			if a.ArchivedAt.After(*b.ArchivedAt) {
				return -1
			}
			return 1
		}
		if a.UpdatedAt.After(b.UpdatedAt) {
			return -1
		}
		if a.UpdatedAt.Before(b.UpdatedAt) {
			return 1
		}
		return strings.Compare(a.ID, b.ID)
	})
	return out, projects, nil
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

	snap, err := s.store.Load(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	task, ok := findTask(snap.Tasks, taskID)
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	if task.ArchivedAt != nil {
		return fmt.Errorf("task is archived: %s", taskID)
	}
	if task.Lane == lane {
		return nil
	}

	oldLane := task.Lane
	newPos := nextLanePosition(snap, lane)
	task.Lane = lane
	task.Position = newPos
	task.UpdatedAt = now
	normalizeLanePositions(snap, oldLane)
	normalizeLanePositions(snap, lane)
	appendEvent(&snap, domain.Event{
		ID:        newID("evt"),
		TaskID:    task.ID,
		Type:      domain.EventLaneChanged,
		Timestamp: now,
		Payload: map[string]interface{}{
			"from": oldLane,
			"to":   lane,
		},
	})
	return s.store.Save(ctx, snap)
}

func (s *TaskService) ToggleParking(ctx context.Context, taskID string) error {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	task, ok := findTask(snap.Tasks, taskID)
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	if task.ArchivedAt != nil {
		return fmt.Errorf("task is archived: %s", taskID)
	}

	oldLane := task.Lane
	eventType := domain.EventTaskParked
	payload := map[string]interface{}{"from": oldLane}

	if oldLane == domain.LaneParking {
		newPos := nextLanePosition(snap, domain.LaneTodo)
		task.Lane = domain.LaneTodo
		task.Position = newPos
		eventType = domain.EventTaskUnparked
		payload = map[string]interface{}{"to": domain.LaneTodo}
	} else {
		newPos := nextLanePosition(snap, domain.LaneParking)
		task.Lane = domain.LaneParking
		task.Position = newPos
	}
	task.UpdatedAt = now

	normalizeLanePositions(snap, oldLane)
	normalizeLanePositions(snap, task.Lane)
	appendEvent(&snap, domain.Event{
		ID:        newID("evt"),
		TaskID:    task.ID,
		Type:      eventType,
		Timestamp: now,
		Payload:   payload,
	})
	return s.store.Save(ctx, snap)
}

func (s *TaskService) Touch(ctx context.Context, taskID string) error {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	task, ok := findTask(snap.Tasks, taskID)
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	if task.ArchivedAt != nil {
		return fmt.Errorf("task is archived: %s", taskID)
	}

	// Bump touched task to the top of its current lane for better recency visibility.
	task.Position = minLanePosition(snap, task.Lane) - 1
	task.LastTouchedAt = &now
	task.UpdatedAt = now
	normalizeLanePositions(snap, task.Lane)
	appendEvent(&snap, domain.Event{
		ID:        newID("evt"),
		TaskID:    task.ID,
		Type:      domain.EventTaskUpdated,
		Timestamp: now,
		Payload: map[string]interface{}{
			"action": "touch",
			"bumped": true,
		},
	})
	return s.store.Save(ctx, snap)
}

func (s *TaskService) MarkDone(ctx context.Context, taskID string) error {
	return s.MoveLane(ctx, taskID, domain.LaneDone)
}

func (s *TaskService) SendBacklog(ctx context.Context, taskID string) error {
	return s.MoveLane(ctx, taskID, domain.LaneBacklog)
}

func (s *TaskService) Archive(ctx context.Context, taskID string) error {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	task, ok := findTask(snap.Tasks, taskID)
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	if task.ArchivedAt != nil {
		return nil
	}
	if task.Lane != domain.LaneDone {
		return fmt.Errorf("archive is allowed only when task is in done")
	}

	fromLane := task.Lane
	task.ArchivedAt = &now
	task.UpdatedAt = now
	normalizeLanePositions(snap, fromLane)
	appendEvent(&snap, domain.Event{
		ID:        newID("evt"),
		TaskID:    task.ID,
		Type:      domain.EventTaskArchived,
		Timestamp: now,
		Payload: map[string]interface{}{
			"from_lane": fromLane,
		},
	})
	return s.store.Save(ctx, snap)
}

func (s *TaskService) EditTitle(ctx context.Context, taskID string, title string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return fmt.Errorf("title is required")
	}

	snap, err := s.store.Load(ctx)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	task, ok := findTask(snap.Tasks, taskID)
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	if task.ArchivedAt != nil {
		return fmt.Errorf("task is archived: %s", taskID)
	}
	if task.Lane != domain.LaneBacklog {
		return fmt.Errorf("edit is allowed only when task is in backlog")
	}
	task.Title = title
	task.UpdatedAt = now
	appendEvent(&snap, domain.Event{
		ID:        newID("evt"),
		TaskID:    taskID,
		Type:      domain.EventTaskUpdated,
		Timestamp: now,
		Payload: map[string]interface{}{
			"action": "edit",
		},
	})
	return s.store.Save(ctx, snap)
}

func (s *TaskService) Delete(ctx context.Context, taskID string) error {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return err
	}

	taskIdx := -1
	for i, t := range snap.Tasks {
		if t.ID == taskID {
			taskIdx = i
			if t.ArchivedAt != nil {
				return fmt.Errorf("task is archived: %s", taskID)
			}
			if t.Lane != domain.LaneBacklog {
				return fmt.Errorf("delete is allowed only when task is in backlog")
			}
			break
		}
	}
	if taskIdx == -1 {
		return fmt.Errorf("task not found: %s", taskID)
	}

	lane := snap.Tasks[taskIdx].Lane
	snap.Tasks = slices.Delete(snap.Tasks, taskIdx, taskIdx+1)
	filteredEvents := make([]domain.Event, 0, len(snap.Events))
	for _, evt := range snap.Events {
		if evt.TaskID != taskID {
			filteredEvents = append(filteredEvents, evt)
		}
	}
	snap.Events = filteredEvents
	normalizeLanePositions(snap, lane)
	return s.store.Save(ctx, snap)
}

func (s *TaskService) ReorderInLane(ctx context.Context, taskID string, direction int) error {
	if direction != -1 && direction != 1 {
		return fmt.Errorf("direction must be -1 or 1")
	}
	snap, err := s.store.Load(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	task, ok := findTask(snap.Tasks, taskID)
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	if task.ArchivedAt != nil {
		return fmt.Errorf("task is archived: %s", taskID)
	}

	type laneRef struct {
		idx int
		pos int
	}
	refs := make([]laneRef, 0)
	for i, t := range snap.Tasks {
		if t.ArchivedAt != nil || t.Lane != task.Lane {
			continue
		}
		refs = append(refs, laneRef{idx: i, pos: t.Position})
	}
	slices.SortFunc(refs, func(a, b laneRef) int { return a.pos - b.pos })
	current := -1
	for i, ref := range refs {
		if snap.Tasks[ref.idx].ID == taskID {
			current = i
			break
		}
	}
	if current == -1 {
		return fmt.Errorf("task not found in lane ordering: %s", taskID)
	}

	target := current + direction
	if target < 0 || target >= len(refs) {
		return nil
	}

	a := refs[current].idx
	b := refs[target].idx
	snap.Tasks[a].Position, snap.Tasks[b].Position = snap.Tasks[b].Position, snap.Tasks[a].Position
	snap.Tasks[a].UpdatedAt = now
	snap.Tasks[b].UpdatedAt = now
	normalizeLanePositions(snap, task.Lane)
	appendEvent(&snap, domain.Event{
		ID:        newID("evt"),
		TaskID:    taskID,
		Type:      domain.EventTaskUpdated,
		Timestamp: now,
		Payload: map[string]interface{}{
			"action":    "reorder",
			"direction": direction,
		},
	})
	return s.store.Save(ctx, snap)
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

func appendEvent(snap *store.Snapshot, evt domain.Event) {
	snap.Events = append(snap.Events, evt)
}

func findTask(tasks []domain.Task, taskID string) (*domain.Task, bool) {
	for i := range tasks {
		if tasks[i].ID == taskID {
			return &tasks[i], true
		}
	}
	return nil, false
}

func nextLanePosition(snap store.Snapshot, lane domain.Lane) int {
	maxPos := 0
	for _, t := range snap.Tasks {
		if t.ArchivedAt == nil && t.Lane == lane && t.Position > maxPos {
			maxPos = t.Position
		}
	}
	return maxPos + 1
}

func minLanePosition(snap store.Snapshot, lane domain.Lane) int {
	minPos := 1
	found := false
	for _, t := range snap.Tasks {
		if t.ArchivedAt == nil && t.Lane == lane {
			if !found || t.Position < minPos {
				minPos = t.Position
				found = true
			}
		}
	}
	if !found {
		return 1
	}
	return minPos
}

func normalizeLanePositions(snap store.Snapshot, lane domain.Lane) {
	indexes := make([]int, 0)
	for i, t := range snap.Tasks {
		if t.ArchivedAt == nil && t.Lane == lane {
			indexes = append(indexes, i)
		}
	}
	slices.SortFunc(indexes, func(a, b int) int {
		if snap.Tasks[a].Position != snap.Tasks[b].Position {
			return snap.Tasks[a].Position - snap.Tasks[b].Position
		}
		return strings.Compare(snap.Tasks[a].ID, snap.Tasks[b].ID)
	})
	for i, idx := range indexes {
		snap.Tasks[idx].Position = i + 1
	}
}

func mapProjects(projects []domain.Project) map[string]domain.Project {
	out := make(map[string]domain.Project, len(projects))
	for _, p := range projects {
		out[p.ID] = p
	}
	return out
}

func sortTasks(tasks []domain.Task) {
	slices.SortFunc(tasks, func(a, b domain.Task) int {
		if a.Lane != b.Lane {
			return laneIndex(a.Lane) - laneIndex(b.Lane)
		}
		if a.Position != b.Position {
			return a.Position - b.Position
		}
		return strings.Compare(a.ID, b.ID)
	})
}

func laneIndex(lane domain.Lane) int {
	for i, l := range domain.LaneOrder {
		if l == lane {
			return i
		}
	}
	return len(domain.LaneOrder)
}
