package domain

import "slices"

type Initiative string

const (
	InitiativeWNT Initiative = "wnt"
	InitiativeCZ  Initiative = "cz"
	InitiativeDM  Initiative = "dm"
)

var InitiativeValues = []Initiative{
	InitiativeWNT,
	InitiativeCZ,
	InitiativeDM,
}

func ValidInitiative(v Initiative) bool {
	return slices.Contains(InitiativeValues, v)
}

type TaskType string

const (
	TaskTypeDiscovery TaskType = "discovery"
	TaskTypeFeat      TaskType = "feat"
	TaskTypeRefact    TaskType = "refact"
	TaskTypeChore     TaskType = "chore"
	TaskTypeFix       TaskType = "fix"
)

var TaskTypeValues = []TaskType{
	TaskTypeDiscovery,
	TaskTypeFeat,
	TaskTypeRefact,
	TaskTypeChore,
	TaskTypeFix,
}

func ValidTaskType(v TaskType) bool {
	return slices.Contains(TaskTypeValues, v)
}

type Loop string

const (
	LoopCoreValue Loop = "core_value"
	LoopGrowth    Loop = "growth"
	LoopRevenue   Loop = "revenue"
	LoopProduct   Loop = "product"
)

var LoopValues = []Loop{
	LoopCoreValue,
	LoopGrowth,
	LoopRevenue,
	LoopProduct,
}

func ValidLoop(v Loop) bool {
	return slices.Contains(LoopValues, v)
}

type EnergyType string

const (
	EnergyProduces EnergyType = "produces"
	EnergyNeutral  EnergyType = "neutral"
	EnergyConsumes EnergyType = "consumes"
)

var EnergyTypeValues = []EnergyType{
	EnergyProduces,
	EnergyNeutral,
	EnergyConsumes,
}

func ValidEnergyType(v EnergyType) bool {
	return slices.Contains(EnergyTypeValues, v)
}

type Nature string

const (
	NatureMushin  Nature = "mushin"
	NatureGambaru Nature = "gambaru"
	NatureShizen  Nature = "shizen"
	NatureGaman   Nature = "gaman"
)

var NatureValues = []Nature{
	NatureMushin,
	NatureGambaru,
	NatureShizen,
	NatureGaman,
}

func ValidNature(v Nature) bool {
	return slices.Contains(NatureValues, v)
}

type Lane string

const (
	LaneBacklog  Lane = "backlog"
	LaneTodo     Lane = "todo"
	LaneDoing    Lane = "doing"
	LaneHalt     Lane = "halt"
	LaneParking  Lane = "parking"
	LaneDone     Lane = "done"
	LaneArchived Lane = "archived"
)

var AllLanes = []Lane{
	LaneBacklog,
	LaneTodo,
	LaneDoing,
	LaneHalt,
	LaneParking,
	LaneDone,
	LaneArchived,
}

var LaneOrder = []Lane{
	LaneBacklog,
	LaneTodo,
	LaneDoing,
	LaneHalt,
	LaneParking,
	LaneDone,
}

func ValidLane(v Lane) bool {
	return slices.Contains(AllLanes, v)
}

func LaneTitle(v Lane) string {
	switch v {
	case LaneBacklog:
		return "BACKLOG"
	case LaneTodo:
		return "TODO"
	case LaneDoing:
		return "DOING"
	case LaneHalt:
		return "HALT"
	case LaneParking:
		return "PARKING"
	case LaneDone:
		return "DONE"
	case LaneArchived:
		return "ARCHIVED"
	default:
		return string(v)
	}
}

type EventType string

const (
	EventTaskCreated  EventType = "task_created"
	EventTaskUpdated  EventType = "task_updated"
	EventLaneChanged  EventType = "lane_changed"
	EventTaskParked   EventType = "task_parked"
	EventTaskUnparked EventType = "task_unparked"
	EventTaskArchived EventType = "task_archived"
)

func ValidEventType(v EventType) bool {
	return slices.Contains([]EventType{
		EventTaskCreated,
		EventTaskUpdated,
		EventLaneChanged,
		EventTaskParked,
		EventTaskUnparked,
		EventTaskArchived,
	}, v)
}
