package service

import (
	"context"
	"fmt"

	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/store"
)

type DoctorReport struct {
	Issues []string
}

func (r DoctorReport) OK() bool {
	return len(r.Issues) == 0
}

type DoctorService struct {
	store store.Store
}

func NewDoctorService(st store.Store) *DoctorService {
	return &DoctorService{store: st}
}

func (s *DoctorService) Check(ctx context.Context) (DoctorReport, error) {
	snap, err := s.store.Load(ctx)
	if err != nil {
		return DoctorReport{}, err
	}

	report := DoctorReport{Issues: []string{}}
	projectByID := make(map[string]domain.Project, len(snap.Projects))
	taskByID := make(map[string]domain.Task, len(snap.Tasks))

	for _, p := range snap.Projects {
		if p.ID == "" {
			report.Issues = append(report.Issues, "project with empty id")
		}
		if _, ok := projectByID[p.ID]; ok {
			report.Issues = append(report.Issues, fmt.Sprintf("duplicate project id: %s", p.ID))
		}
		projectByID[p.ID] = p
		if !domain.ValidInitiative(p.Initiative) {
			report.Issues = append(report.Issues, fmt.Sprintf("project %q has invalid initiative %q", p.Name, p.Initiative))
		}
	}

	for _, t := range snap.Tasks {
		if t.ID == "" {
			report.Issues = append(report.Issues, "task with empty id")
		}
		if _, ok := taskByID[t.ID]; ok {
			report.Issues = append(report.Issues, fmt.Sprintf("duplicate task id: %s", t.ID))
		}
		taskByID[t.ID] = t

		if _, ok := projectByID[t.ProjectID]; !ok {
			report.Issues = append(report.Issues, fmt.Sprintf("task %q references unknown project id %q", t.ID, t.ProjectID))
		}
		if !domain.ValidInitiative(t.Initiative) {
			report.Issues = append(report.Issues, fmt.Sprintf("task %q has invalid initiative %q", t.ID, t.Initiative))
		}
		if !domain.ValidLane(t.Lane) {
			report.Issues = append(report.Issues, fmt.Sprintf("task %q has invalid lane %q", t.ID, t.Lane))
		}
	}

	for _, evt := range snap.Events {
		if evt.ID == "" {
			report.Issues = append(report.Issues, "event with empty id")
		}
		if _, ok := taskByID[evt.TaskID]; !ok {
			report.Issues = append(report.Issues, fmt.Sprintf("event %q references unknown task id %q", evt.ID, evt.TaskID))
		}
		if !domain.ValidEventType(evt.Type) {
			report.Issues = append(report.Issues, fmt.Sprintf("event %q has invalid type %q", evt.ID, evt.Type))
		}
	}

	return report, nil
}
