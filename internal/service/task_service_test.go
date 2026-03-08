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
