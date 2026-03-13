package service

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/store/file"
)

func TestTaskService_TransitionValidation(t *testing.T) {
	t.Parallel()

	svc := newTaskServiceForTest(t)
	ctx := context.Background()
	task := createTaskForTest(t, ctx, svc, "task transition validation")

	if err := svc.Transition(ctx, task.ID, domain.LaneTodo, false); err != nil {
		t.Fatalf("valid transition backlog->todo failed: %v", err)
	}

	if err := svc.Transition(ctx, task.ID, domain.LaneDone, false); err == nil {
		t.Fatal("expected invalid transition todo->done to fail")
	} else if !strings.Contains(err.Error(), "transition not allowed") {
		t.Fatalf("unexpected error for invalid transition: %v", err)
	}
}

func TestTaskService_RecoveryTransitionRequiresConfirmation(t *testing.T) {
	t.Parallel()

	svc := newTaskServiceForTest(t)
	ctx := context.Background()
	task := createTaskForTest(t, ctx, svc, "task recovery confirm")

	mustNoErr(t, svc.Transition(ctx, task.ID, domain.LaneTodo, false))

	err := svc.Transition(ctx, task.ID, domain.LaneBacklog, false)
	if err == nil || !strings.Contains(err.Error(), "confirmation required") {
		t.Fatalf("expected confirmation-required error, got: %v", err)
	}

	mustNoErr(t, svc.Transition(ctx, task.ID, domain.LaneBacklog, true))
}

func TestTaskService_DestructiveRules(t *testing.T) {
	t.Parallel()

	svc := newTaskServiceForTest(t)
	ctx := context.Background()

	task := createTaskForTest(t, ctx, svc, "archive from backlog")
	err := svc.Archive(ctx, task.ID)
	if err == nil || !strings.Contains(err.Error(), "confirmation required") {
		t.Fatalf("expected confirmation for backlog->archived, got: %v", err)
	}
	mustNoErr(t, svc.Transition(ctx, task.ID, domain.LaneArchived, true))

	task2 := createTaskForTest(t, ctx, svc, "delete confirm")
	err = svc.Delete(ctx, task2.ID, false)
	if err == nil || !strings.Contains(err.Error(), "confirmation required") {
		t.Fatalf("expected delete confirmation required error, got: %v", err)
	}
	mustNoErr(t, svc.Delete(ctx, task2.ID, true))
}

func TestTaskService_BacklogOnlyEditDelete(t *testing.T) {
	t.Parallel()

	svc := newTaskServiceForTest(t)
	ctx := context.Background()
	task := createTaskForTest(t, ctx, svc, "backlog only actions")

	mustNoErr(t, svc.Transition(ctx, task.ID, domain.LaneTodo, false))

	err := svc.EditTitle(ctx, task.ID, "edited title")
	if err == nil || !strings.Contains(err.Error(), "edit only allowed in backlog") {
		t.Fatalf("expected backlog-only edit error, got: %v", err)
	}

	err = svc.Delete(ctx, task.ID, true)
	if err == nil || !strings.Contains(err.Error(), "delete only allowed in backlog") {
		t.Fatalf("expected backlog-only delete error, got: %v", err)
	}
}

func TestTaskService_Edit_BacklogPersistsAndRecordsEvent(t *testing.T) {
	t.Parallel()

	svc := newTaskServiceForTest(t)
	ctx := context.Background()
	task := createTaskForTest(t, ctx, svc, "old title")

	err := svc.Edit(ctx, task.ID, EditTaskInput{
		Title:       "new title",
		Description: "new description",
		ProjectName: "new project",
		Initiative:  domain.InitiativeDM,
		Type:        domain.TaskTypeFix,
		Loop:        domain.LoopGrowth,
		EnergyType:  domain.EnergyConsumes,
		Nature:      domain.NatureGaman,
	})
	mustNoErr(t, err)

	edited, ok, err := svc.Get(ctx, task.ID)
	mustNoErr(t, err)
	if !ok {
		t.Fatalf("edited task not found: %s", task.ID)
	}
	if edited.Title != "new title" || edited.Description != "new description" {
		t.Fatalf("task fields not updated: %+v", edited)
	}
	if edited.Initiative != domain.InitiativeDM {
		t.Fatalf("initiative not updated: %s", edited.Initiative)
	}
	if edited.Type != domain.TaskTypeFix || edited.Loop != domain.LoopGrowth {
		t.Fatalf("type/loop not updated: type=%s loop=%s", edited.Type, edited.Loop)
	}
	if edited.EnergyType != domain.EnergyConsumes || edited.Nature != domain.NatureGaman {
		t.Fatalf("energy/nature not updated: energy=%s nature=%s", edited.EnergyType, edited.Nature)
	}

	events, err := svc.RecentEvents(ctx, task.ID, 20)
	mustNoErr(t, err)
	if len(events) == 0 {
		t.Fatalf("expected update events after edit")
	}
	last := events[len(events)-1]
	if last.Type != domain.EventTaskUpdated {
		t.Fatalf("expected last event task_updated, got %s", last.Type)
	}
	if action, ok := last.Payload["action"]; !ok || action != "edit" {
		t.Fatalf("expected edit action payload, got %+v", last.Payload)
	}
}

func TestTaskService_CreateAndEdit_NeutralEnergy(t *testing.T) {
	t.Parallel()

	svc := newTaskServiceForTest(t)
	ctx := context.Background()
	task, err := svc.Create(ctx, CreateTaskInput{
		Title:       "neutral task",
		Initiative:  domain.InitiativeWNT,
		ProjectName: "core shell",
		Type:        domain.TaskTypeChore,
		Loop:        domain.LoopCoreValue,
		EnergyType:  domain.EnergyNeutral,
		Nature:      domain.NatureShizen,
		Lane:        domain.LaneBacklog,
	})
	mustNoErr(t, err)

	if task.EnergyType != domain.EnergyNeutral {
		t.Fatalf("expected created task neutral energy, got %s", task.EnergyType)
	}

	err = svc.Edit(ctx, task.ID, EditTaskInput{
		Title:       task.Title,
		Description: task.Description,
		ProjectName: "core shell",
		Initiative:  task.Initiative,
		Type:        task.Type,
		Loop:        task.Loop,
		EnergyType:  domain.EnergyNeutral,
		Nature:      task.Nature,
	})
	mustNoErr(t, err)

	edited, ok, err := svc.Get(ctx, task.ID)
	mustNoErr(t, err)
	if !ok {
		t.Fatalf("task not found after edit: %s", task.ID)
	}
	if edited.EnergyType != domain.EnergyNeutral {
		t.Fatalf("expected neutral energy to be preserved, got %s", edited.EnergyType)
	}
}

func newTaskServiceForTest(t *testing.T) *TaskService {
	t.Helper()
	path := filepath.Join(t.TempDir(), "data.json")
	st := file.New(path)
	projectSvc := NewProjectService(st)
	return NewTaskService(st, projectSvc)
}

func createTaskForTest(t *testing.T, ctx context.Context, svc *TaskService, title string) domain.Task {
	t.Helper()
	task, err := svc.Create(ctx, CreateTaskInput{
		Title:       title,
		Initiative:  domain.InitiativeCZ,
		ProjectName: "collision playground",
		Type:        domain.TaskTypeFeat,
		Loop:        domain.LoopProduct,
		EnergyType:  domain.EnergyProduces,
		Nature:      domain.NatureMushin,
		Lane:        domain.LaneBacklog,
	})
	mustNoErr(t, err)
	return task
}

func mustNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
