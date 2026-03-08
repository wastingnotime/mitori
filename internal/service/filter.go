package service

import (
	"strings"

	"github.com/wastingnotime/mitori/internal/domain"
)

type TaskFilter struct {
	ProjectID  string
	Initiative domain.Initiative
	EnergyType domain.EnergyType
	TaskType   domain.TaskType
	Query      string
}

func MatchTaskFilter(task domain.Task, projectName string, filter TaskFilter) bool {
	if filter.ProjectID != "" && task.ProjectID != filter.ProjectID {
		return false
	}
	if filter.Initiative != "" && task.Initiative != filter.Initiative {
		return false
	}
	if filter.EnergyType != "" && task.EnergyType != filter.EnergyType {
		return false
	}
	if filter.TaskType != "" && task.Type != filter.TaskType {
		return false
	}
	if strings.TrimSpace(filter.Query) == "" {
		return true
	}

	q := strings.ToLower(strings.TrimSpace(filter.Query))
	return strings.Contains(strings.ToLower(task.Title), q) ||
		strings.Contains(strings.ToLower(task.Description), q) ||
		strings.Contains(strings.ToLower(projectName), q)
}
