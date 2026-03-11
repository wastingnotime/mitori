package export

import (
	"time"

	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/service"
)

type TaskItem struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	ProjectID   string            `json:"project_id"`
	Project     string            `json:"project"`
	Initiative  domain.Initiative `json:"initiative"`
	Type        domain.TaskType   `json:"type"`
	Loop        domain.Loop       `json:"loop"`
	EnergyType  domain.EnergyType `json:"energy_type"`
	Nature      domain.Nature     `json:"nature"`
	Lane        domain.Lane       `json:"lane"`
	Position    int               `json:"position"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	ArchivedAt  *time.Time        `json:"archived_at,omitempty"`
}

type EventItem struct {
	ID        string                 `json:"id"`
	TaskID    string                 `json:"task_id"`
	Type      domain.EventType       `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}

type BoardExport struct {
	Type          string                     `json:"type"`
	GeneratedAt   time.Time                  `json:"generated_at"`
	ActiveFilters string                     `json:"active_filters"`
	LaneCounts    map[domain.Lane]int        `json:"lane_counts"`
	Energy        service.BoardEnergySummary `json:"energy"`
	TasksByLane   map[domain.Lane][]TaskItem `json:"tasks_by_lane"`
}

type ProjectExport struct {
	Type         string                                `json:"type"`
	GeneratedAt  time.Time                             `json:"generated_at"`
	Project      domain.Project                        `json:"project"`
	LaneCounts   map[domain.Lane]int                   `json:"lane_counts"`
	EnergyTotal  service.EnergySummary                 `json:"energy_total"`
	EnergyByLane map[domain.Lane]service.EnergySummary `json:"energy_by_lane"`
	EnergyNote   string                                `json:"energy_note"`
	TasksByLane  map[domain.Lane][]TaskItem            `json:"tasks_by_lane"`
	Archived     []TaskItem                            `json:"archived_tasks,omitempty"`
	RecentEvents []EventItem                           `json:"recent_events,omitempty"`
}

type TaskExport struct {
	Type        string      `json:"type"`
	GeneratedAt time.Time   `json:"generated_at"`
	Task        TaskItem    `json:"task"`
	Events      []EventItem `json:"events"`
}

type ArchiveGroup struct {
	Label string     `json:"label"`
	Tasks []TaskItem `json:"tasks"`
}

type ArchiveExport struct {
	Type        string         `json:"type"`
	GeneratedAt time.Time      `json:"generated_at"`
	GroupMode   string         `json:"group_mode"`
	Groups      []ArchiveGroup `json:"groups"`
}
