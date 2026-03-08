package domain

import "time"

type Task struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Initiative    Initiative `json:"initiative"`
	ProjectID     string     `json:"project_id"`
	Type          TaskType   `json:"type"`
	Loop          Loop       `json:"loop"`
	EnergyType    EnergyType `json:"energy_type"`
	Nature        Nature     `json:"nature"`
	Lane          Lane       `json:"lane"`
	Position      int        `json:"position"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	LastTouchedAt *time.Time `json:"last_touched_at,omitempty"`
	ArchivedAt    *time.Time `json:"archived_at,omitempty"`
}
