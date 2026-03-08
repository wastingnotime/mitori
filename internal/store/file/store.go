package file

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/store"
)

type Store struct {
	path string
}

func New(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Path() string {
	return s.path
}

func (s *Store) Init(_ context.Context) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	_, err := os.Stat(s.path)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	data := EmptyData()
	return s.write(data)
}

func (s *Store) Load(ctx context.Context) (store.Snapshot, error) {
	select {
	case <-ctx.Done():
		return store.Snapshot{}, ctx.Err()
	default:
	}

	if err := s.Init(ctx); err != nil {
		return store.Snapshot{}, err
	}

	b, err := os.ReadFile(s.path)
	if err != nil {
		return store.Snapshot{}, err
	}

	if len(b) == 0 {
		empty := EmptyData()
		return toSnapshot(empty), nil
	}

	var data Data
	if err := json.Unmarshal(b, &data); err != nil {
		return store.Snapshot{}, err
	}
	if data.Projects == nil {
		data.Projects = []domain.Project{}
	}
	if data.Tasks == nil {
		data.Tasks = []domain.Task{}
	}
	if data.Events == nil {
		data.Events = []domain.Event{}
	}
	return toSnapshot(data), nil
}

func (s *Store) Save(ctx context.Context, snap store.Snapshot) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := s.Init(ctx); err != nil {
		return err
	}

	data := Data{
		Projects: snap.Projects,
		Tasks:    snap.Tasks,
		Events:   snap.Events,
	}
	return s.write(data)
}

func (s *Store) write(data Data) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(s.path, b, 0o644)
}

func toSnapshot(data Data) store.Snapshot {
	return store.Snapshot{
		Projects: data.Projects,
		Tasks:    data.Tasks,
		Events:   data.Events,
	}
}
