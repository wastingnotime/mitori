package domain

import "time"

type Project struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Initiative  Initiative `json:"initiative"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ArchivedAt  *time.Time `json:"archived_at,omitempty"`
}
