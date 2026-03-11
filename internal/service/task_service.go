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

type EditTaskInput struct {
	Title       string
	Description string
	ProjectName string
	Initiative  domain.Initiative
	Type        domain.TaskType
	Loop        domain.Loop
	EnergyType  domain.EnergyType
	Nature      domain.Nature
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
	if !domain.ValidLane(in.Lane) || in.Lane == domain.LaneArchived {
		in.Lane = domain.LaneBacklog
	}

	snap, err := s.store.Load(ctx)
	if err != nil {
		return domain.Task{}, err
	}
	normalizeLegacyArchive(&snap)

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
	appendEvent(&snap, domain.Event{
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
	normalizeLegacyArchive(&snap)

	out := make([]domain.Task, 0, len(snap.Tasks))
	for _, t := range snap.Tasks {
		if t.Lane != domain.LaneArchived && t.ArchivedAt == nil {
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
	normalizeLegacyArchive(&snap)

	projects := mapProjects(snap.Projects)
	out := make([]domain.Task, 0)
	for _, t := range snap.Tasks {
		if t.Lane != domain.LaneArchived && t.ArchivedAt == nil {
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
	normalizeLegacyArchive(&snap)
	for _, t := range snap.Tasks {
		if t.ID == taskID {
			return t, true, nil
		}
	}
	return domain.Task{}, false, nil
}

func (s *TaskService) AllowedActionsForTask(task domain.Task) []domain.TaskAction {
	return domain.AllowedTaskActions(task.Lane)
}

func (s *TaskService) Transition(ctx context.Context, taskID string, to domain.Lane, confirmed bool) error {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return err
	}
	normalizeLegacyArchive(&snap)

	task, ok := findTask(snap.Tasks, taskID)
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	from := task.Lane
	if from == to {
		return nil
	}

	tr, allowed := domain.CanTransition(from, to)
	if !allowed {
		return fmt.Errorf("transition not allowed from %s to %s", from, to)
	}
	if tr.RequiresConfirmation && !confirmed {
		return fmt.Errorf("confirmation required for transition %s -> %s", from, to)
	}

	now := time.Now().UTC()
	oldLane := task.Lane
	if to == domain.LaneArchived {
		task.Lane = domain.LaneArchived
		task.ArchivedAt = &now
		task.Position = 0
		normalizeLanePositions(snap, oldLane)
	} else {
		task.Lane = to
		task.Position = nextLanePosition(snap, to)
		if oldLane == domain.LaneArchived {
			task.ArchivedAt = nil
		}
		normalizeLanePositions(snap, oldLane)
		normalizeLanePositions(snap, to)
	}
	task.UpdatedAt = now

	appendEvent(&snap, domain.Event{
		ID:        newID("evt"),
		TaskID:    task.ID,
		Type:      domain.EventLaneChanged,
		Timestamp: now,
		Payload: map[string]interface{}{
			"from":      from,
			"to":        to,
			"class":     tr.Class,
			"confirmed": confirmed,
		},
	})
	if to == domain.LaneArchived {
		appendEvent(&snap, domain.Event{
			ID:        newID("evt"),
			TaskID:    task.ID,
			Type:      domain.EventTaskArchived,
			Timestamp: now,
			Payload: map[string]interface{}{
				"from_lane": from,
			},
		})
	}

	return s.store.Save(ctx, snap)
}

func (s *TaskService) MoveLane(ctx context.Context, taskID string, lane domain.Lane) error {
	return s.Transition(ctx, taskID, lane, false)
}

func (s *TaskService) MarkDone(ctx context.Context, taskID string) error {
	return s.Transition(ctx, taskID, domain.LaneDone, false)
}

func (s *TaskService) SendBacklog(ctx context.Context, taskID string) error {
	return s.Transition(ctx, taskID, domain.LaneBacklog, false)
}

func (s *TaskService) Archive(ctx context.Context, taskID string) error {
	return s.Transition(ctx, taskID, domain.LaneArchived, false)
}

func (s *TaskService) Touch(ctx context.Context, taskID string) error {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return err
	}
	normalizeLegacyArchive(&snap)

	now := time.Now().UTC()
	task, ok := findTask(snap.Tasks, taskID)
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	if task.Lane == domain.LaneArchived {
		return fmt.Errorf("touch not allowed in archived lane")
	}

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

func (s *TaskService) EditTitle(ctx context.Context, taskID string, title string) error {
	task, ok, err := s.Get(ctx, taskID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	projectName := "general"
	snap, err := s.store.Load(ctx)
	if err != nil {
		return err
	}
	for _, p := range snap.Projects {
		if p.ID == task.ProjectID {
			projectName = p.Name
			break
		}
	}
	return s.Edit(ctx, taskID, EditTaskInput{
		Title:       title,
		Description: task.Description,
		ProjectName: projectName,
		Initiative:  task.Initiative,
		Type:        task.Type,
		Loop:        task.Loop,
		EnergyType:  task.EnergyType,
		Nature:      task.Nature,
	})
}

func (s *TaskService) Edit(ctx context.Context, taskID string, in EditTaskInput) error {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return fmt.Errorf("title is required")
	}
	description := strings.TrimSpace(in.Description)
	projectName := strings.TrimSpace(in.ProjectName)
	if projectName == "" {
		return fmt.Errorf("project is required")
	}
	if !domain.ValidInitiative(in.Initiative) {
		return fmt.Errorf("initiative is invalid")
	}
	if !domain.ValidTaskType(in.Type) {
		return fmt.Errorf("task type is invalid")
	}
	if !domain.ValidLoop(in.Loop) {
		return fmt.Errorf("loop is invalid")
	}
	if !domain.ValidEnergyType(in.EnergyType) {
		return fmt.Errorf("energy type is invalid")
	}
	if !domain.ValidNature(in.Nature) {
		return fmt.Errorf("nature is invalid")
	}

	project, err := s.projectSvc.Ensure(ctx, in.Initiative, projectName)
	if err != nil {
		return err
	}

	snap, err := s.store.Load(ctx)
	if err != nil {
		return err
	}
	normalizeLegacyArchive(&snap)

	now := time.Now().UTC()
	task, ok := findTask(snap.Tasks, taskID)
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	if !canEdit(task.Lane) {
		return fmt.Errorf("edit only allowed in backlog")
	}

	task.Title = title
	task.Description = description
	task.ProjectID = project.ID
	task.Initiative = in.Initiative
	task.Type = in.Type
	task.Loop = in.Loop
	task.EnergyType = in.EnergyType
	task.Nature = in.Nature
	task.UpdatedAt = now
	appendEvent(&snap, domain.Event{
		ID:        newID("evt"),
		TaskID:    taskID,
		Type:      domain.EventTaskUpdated,
		Timestamp: now,
		Payload: map[string]interface{}{
			"action": "edit",
			"fields": []string{
				"title",
				"description",
				"project",
				"initiative",
				"type",
				"loop",
				"energy_type",
				"nature",
			},
		},
	})
	return s.store.Save(ctx, snap)
}

func (s *TaskService) Delete(ctx context.Context, taskID string, confirmed bool) error {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return err
	}
	normalizeLegacyArchive(&snap)

	taskIdx := -1
	for i, t := range snap.Tasks {
		if t.ID == taskID {
			taskIdx = i
			if !canDelete(t.Lane) {
				return fmt.Errorf("delete only allowed in backlog")
			}
			break
		}
	}
	if taskIdx == -1 {
		return fmt.Errorf("task not found: %s", taskID)
	}
	if !confirmed {
		return fmt.Errorf("confirmation required for delete")
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
	normalizeLegacyArchive(&snap)

	now := time.Now().UTC()
	task, ok := findTask(snap.Tasks, taskID)
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}
	if task.Lane == domain.LaneArchived {
		return fmt.Errorf("cannot reorder archived task")
	}

	type laneRef struct {
		idx int
		pos int
	}
	refs := make([]laneRef, 0)
	for i, t := range snap.Tasks {
		if t.Lane != task.Lane {
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

func canEdit(lane domain.Lane) bool {
	for _, action := range domain.AllowedTaskActions(lane) {
		if action.Kind == domain.ActionEdit {
			return true
		}
	}
	return false
}

func canDelete(lane domain.Lane) bool {
	for _, action := range domain.AllowedTaskActions(lane) {
		if action.Kind == domain.ActionDelete {
			return true
		}
	}
	return false
}

func normalizeLegacyArchive(snap *store.Snapshot) {
	for i := range snap.Tasks {
		if snap.Tasks[i].ArchivedAt != nil && snap.Tasks[i].Lane != domain.LaneArchived {
			snap.Tasks[i].Lane = domain.LaneArchived
			snap.Tasks[i].Position = 0
		}
	}
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
		if t.Lane == lane && t.Position > maxPos {
			maxPos = t.Position
		}
	}
	return maxPos + 1
}

func minLanePosition(snap store.Snapshot, lane domain.Lane) int {
	minPos := 1
	found := false
	for _, t := range snap.Tasks {
		if t.Lane == lane {
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
	if lane == domain.LaneArchived {
		return
	}
	indexes := make([]int, 0)
	for i, t := range snap.Tasks {
		if t.Lane == lane {
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
	if lane == domain.LaneArchived {
		return len(domain.LaneOrder)
	}
	return len(domain.LaneOrder) + 1
}
