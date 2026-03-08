package service

import (
	"context"
	"slices"
	"strings"
	"time"

	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/store"
)

type ProjectService struct {
	store store.Store
}

func NewProjectService(st store.Store) *ProjectService {
	return &ProjectService{store: st}
}

func (s *ProjectService) List(ctx context.Context) ([]domain.Project, error) {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return nil, err
	}
	projects := slices.Clone(snap.Projects)
	slices.SortFunc(projects, func(a, b domain.Project) int {
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})
	return projects, nil
}

func (s *ProjectService) Ensure(ctx context.Context, initiative domain.Initiative, name string) (domain.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "general"
	}
	if !domain.ValidInitiative(initiative) {
		initiative = domain.InitiativeWNT
	}

	snap, err := s.store.Load(ctx)
	if err != nil {
		return domain.Project{}, err
	}

	for _, p := range snap.Projects {
		if strings.EqualFold(p.Name, name) && p.Initiative == initiative {
			return p, nil
		}
	}

	now := time.Now().UTC()
	project := domain.Project{
		ID:         newID("prj"),
		Name:       name,
		Initiative: initiative,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	snap.Projects = append(snap.Projects, project)
	if err := s.store.Save(ctx, snap); err != nil {
		return domain.Project{}, err
	}
	return project, nil
}
