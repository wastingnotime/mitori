package store

import (
	"context"

	"github.com/wastingnotime/mitori/internal/domain"
)

type Snapshot struct {
	Projects []domain.Project `json:"projects"`
	Tasks    []domain.Task    `json:"tasks"`
	Events   []domain.Event   `json:"events"`
}

type Store interface {
	Init(context.Context) error
	Load(context.Context) (Snapshot, error)
	Save(context.Context, Snapshot) error
	Path() string
}
