package file

import (
	"github.com/wastingnotime/mitori/internal/domain"
)

type Data struct {
	Projects []domain.Project `json:"projects"`
	Tasks    []domain.Task    `json:"tasks"`
	Events   []domain.Event   `json:"events"`
}

func EmptyData() Data {
	return Data{
		Projects: []domain.Project{},
		Tasks:    []domain.Task{},
		Events:   []domain.Event{},
	}
}
