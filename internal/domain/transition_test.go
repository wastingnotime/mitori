package domain

import "testing"

func TestCanTransition_NormalAndRecovery(t *testing.T) {
	tests := []struct {
		name                string
		from                Lane
		to                  Lane
		wantAllowed         bool
		wantRequiresConfirm bool
	}{
		// Normal
		{"backlog_to_todo", LaneBacklog, LaneTodo, true, false},
		{"backlog_to_archived", LaneBacklog, LaneArchived, true, true},
		{"todo_to_doing", LaneTodo, LaneDoing, true, false},
		{"todo_to_parking", LaneTodo, LaneParking, true, false},
		{"doing_to_halt", LaneDoing, LaneHalt, true, false},
		{"halt_to_doing", LaneHalt, LaneDoing, true, false},
		{"doing_to_parking", LaneDoing, LaneParking, true, false},
		{"parking_to_doing", LaneParking, LaneDoing, true, false},
		{"doing_to_done", LaneDoing, LaneDone, true, false},
		{"done_to_archived", LaneDone, LaneArchived, true, false},

		// Recovery
		{"todo_to_backlog", LaneTodo, LaneBacklog, true, true},
		{"doing_to_todo", LaneDoing, LaneTodo, true, true},
		{"done_to_doing", LaneDone, LaneDoing, true, true},
		{"archived_to_done", LaneArchived, LaneDone, true, true},
		{"archived_to_backlog", LaneArchived, LaneBacklog, true, true},

		// Invalid representatives
		{"backlog_to_done_invalid", LaneBacklog, LaneDone, false, false},
		{"todo_to_archived_invalid", LaneTodo, LaneArchived, false, false},
		{"halt_to_done_invalid", LaneHalt, LaneDone, false, false},
		{"archived_to_doing_invalid", LaneArchived, LaneDoing, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := CanTransition(tt.from, tt.to)
			if ok != tt.wantAllowed {
				t.Fatalf("allowed mismatch: got=%v want=%v", ok, tt.wantAllowed)
			}
			if !tt.wantAllowed {
				return
			}
			if got.RequiresConfirmation != tt.wantRequiresConfirm {
				t.Fatalf("requires confirmation mismatch: got=%v want=%v", got.RequiresConfirmation, tt.wantRequiresConfirm)
			}
		})
	}
}

func TestAllowedTaskActions_LaneScoped(t *testing.T) {
	backlog := AllowedTaskActions(LaneBacklog)
	if !hasAction(backlog, ActionEdit, "") {
		t.Fatal("expected edit action in backlog")
	}
	if !hasAction(backlog, ActionDelete, "") {
		t.Fatal("expected delete action in backlog")
	}
	if !hasMove(backlog, LaneArchived) {
		t.Fatal("expected archive transition in backlog")
	}
	if !hasDeleteConfirm(backlog) {
		t.Fatal("expected delete action to require confirmation")
	}

	todo := AllowedTaskActions(LaneTodo)
	if hasAction(todo, ActionEdit, "") || hasAction(todo, ActionDelete, "") {
		t.Fatal("did not expect edit/delete outside backlog")
	}
	if hasMove(todo, LaneArchived) {
		t.Fatal("did not expect archive transition in todo")
	}

	doing := AllowedTaskActions(LaneDoing)
	if hasMove(doing, LaneArchived) {
		t.Fatal("did not expect archive transition in doing")
	}

	done := AllowedTaskActions(LaneDone)
	if !hasMove(done, LaneArchived) {
		t.Fatal("expected archive transition in done")
	}
}

func hasMove(actions []TaskAction, to Lane) bool {
	for _, a := range actions {
		if a.Kind == ActionMove && a.To == to {
			return true
		}
	}
	return false
}

func hasAction(actions []TaskAction, kind ActionKind, to Lane) bool {
	for _, a := range actions {
		if a.Kind == kind {
			if kind != ActionMove || a.To == to {
				return true
			}
		}
	}
	return false
}

func hasDeleteConfirm(actions []TaskAction) bool {
	for _, a := range actions {
		if a.Kind == ActionDelete {
			return a.RequiresConfirmation
		}
	}
	return false
}
