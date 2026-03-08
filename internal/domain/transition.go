package domain

type ActionKind string

const (
	ActionMove   ActionKind = "move"
	ActionEdit   ActionKind = "edit"
	ActionDelete ActionKind = "delete"
)

type TransitionClass string

const (
	TransitionNormal   TransitionClass = "normal"
	TransitionRecovery TransitionClass = "recovery"
)

type TaskAction struct {
	ID                   string
	Label                string
	Shortcut             string
	Kind                 ActionKind
	From                 Lane
	To                   Lane
	Class                TransitionClass
	RequiresConfirmation bool
}

var transitionActions = []TaskAction{
	// Normal operational transitions
	{ID: "move_backlog_todo", Label: "todo", Shortcut: "m", Kind: ActionMove, From: LaneBacklog, To: LaneTodo, Class: TransitionNormal},
	{ID: "move_backlog_archived", Label: "archive", Shortcut: "x", Kind: ActionMove, From: LaneBacklog, To: LaneArchived, Class: TransitionNormal, RequiresConfirmation: true},
	{ID: "move_todo_doing", Label: "doing", Shortcut: "m", Kind: ActionMove, From: LaneTodo, To: LaneDoing, Class: TransitionNormal},
	{ID: "move_todo_parking", Label: "park", Shortcut: "p", Kind: ActionMove, From: LaneTodo, To: LaneParking, Class: TransitionNormal},
	{ID: "move_doing_halt", Label: "halt", Shortcut: "s", Kind: ActionMove, From: LaneDoing, To: LaneHalt, Class: TransitionNormal},
	{ID: "move_halt_doing", Label: "resume halt", Shortcut: "s", Kind: ActionMove, From: LaneHalt, To: LaneDoing, Class: TransitionNormal},
	{ID: "move_doing_parking", Label: "park", Shortcut: "p", Kind: ActionMove, From: LaneDoing, To: LaneParking, Class: TransitionNormal},
	{ID: "move_parking_doing", Label: "unpark", Shortcut: "p", Kind: ActionMove, From: LaneParking, To: LaneDoing, Class: TransitionNormal},
	{ID: "move_doing_done", Label: "done", Shortcut: "m", Kind: ActionMove, From: LaneDoing, To: LaneDone, Class: TransitionNormal},
	{ID: "move_done_archived", Label: "archive", Shortcut: "m", Kind: ActionMove, From: LaneDone, To: LaneArchived, Class: TransitionNormal},

	// Recovery transitions
	{ID: "move_todo_backlog", Label: "backlog", Shortcut: "b", Kind: ActionMove, From: LaneTodo, To: LaneBacklog, Class: TransitionRecovery, RequiresConfirmation: true},
	{ID: "move_doing_todo", Label: "todo", Shortcut: "b", Kind: ActionMove, From: LaneDoing, To: LaneTodo, Class: TransitionRecovery, RequiresConfirmation: true},
	{ID: "move_done_doing", Label: "reopen doing", Shortcut: "b", Kind: ActionMove, From: LaneDone, To: LaneDoing, Class: TransitionRecovery, RequiresConfirmation: true},
	{ID: "move_archived_done", Label: "restore done", Shortcut: "b", Kind: ActionMove, From: LaneArchived, To: LaneDone, Class: TransitionRecovery, RequiresConfirmation: true},
	{ID: "move_archived_backlog", Label: "restore backlog", Shortcut: "x", Kind: ActionMove, From: LaneArchived, To: LaneBacklog, Class: TransitionRecovery, RequiresConfirmation: true},
}

func AllowedTaskActions(lane Lane) []TaskAction {
	actions := make([]TaskAction, 0)
	for _, action := range transitionActions {
		if action.From == lane {
			actions = append(actions, action)
		}
	}
	if lane == LaneBacklog {
		actions = append(actions,
			TaskAction{
				ID:       "edit_task",
				Label:    "edit",
				Shortcut: "e",
				Kind:     ActionEdit,
				From:     LaneBacklog,
			},
			TaskAction{
				ID:                   "delete_task",
				Label:                "delete",
				Shortcut:             "X",
				Kind:                 ActionDelete,
				From:                 LaneBacklog,
				RequiresConfirmation: true,
			},
		)
	}
	return actions
}

func CanTransition(from, to Lane) (TaskAction, bool) {
	for _, action := range transitionActions {
		if action.From == from && action.To == to {
			return action, true
		}
	}
	return TaskAction{}, false
}
