package service

import (
	"context"
	"slices"
	"strings"

	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/store"
)

type BoardService struct {
	store store.Store
}

func NewBoardService(st store.Store) *BoardService {
	return &BoardService{store: st}
}

func (s *BoardService) Board(ctx context.Context) (map[domain.Lane][]domain.Task, map[string]domain.Project, error) {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return nil, nil, err
	}

	projects := make(map[string]domain.Project, len(snap.Projects))
	for _, p := range snap.Projects {
		projects[p.ID] = p
	}

	board := make(map[domain.Lane][]domain.Task, len(domain.LaneOrder))
	for _, lane := range domain.LaneOrder {
		board[lane] = []domain.Task{}
	}
	for _, t := range snap.Tasks {
		if t.ArchivedAt != nil {
			continue
		}
		board[t.Lane] = append(board[t.Lane], t)
	}

	for lane := range board {
		slices.SortFunc(board[lane], func(a, b domain.Task) int {
			if a.Position != b.Position {
				return a.Position - b.Position
			}
			return strings.Compare(a.ID, b.ID)
		})
	}
	return board, projects, nil
}
