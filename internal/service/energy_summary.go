package service

import "github.com/wastingnotime/mitori/internal/domain"

type EnergySummary struct {
	Produces int
	Consumes int
}

type BoardEnergySummary struct {
	Total  EnergySummary
	ByLane map[domain.Lane]EnergySummary
	Note   string
}

type ProjectEnergySummary struct {
	Total  EnergySummary
	ByLane map[domain.Lane]EnergySummary
	Note   string
}

func BuildBoardEnergySummary(board map[domain.Lane][]domain.Task) BoardEnergySummary {
	byLane := make(map[domain.Lane]EnergySummary)
	total := EnergySummary{}
	for lane, tasks := range board {
		s := summarizeTasksEnergy(tasks)
		byLane[lane] = s
		total.Produces += s.Produces
		total.Consumes += s.Consumes
	}
	return BoardEnergySummary{
		Total:  total,
		ByLane: byLane,
		Note:   energyNote(byLane),
	}
}

func BuildProjectEnergySummary(tasks []domain.Task) ProjectEnergySummary {
	byLane := make(map[domain.Lane]EnergySummary)
	total := EnergySummary{}
	for _, task := range tasks {
		entry := byLane[task.Lane]
		switch task.EnergyType {
		case domain.EnergyProduces:
			entry.Produces++
			total.Produces++
		case domain.EnergyConsumes:
			entry.Consumes++
			total.Consumes++
		}
		byLane[task.Lane] = entry
	}
	return ProjectEnergySummary{
		Total:  total,
		ByLane: byLane,
		Note:   energyNote(byLane),
	}
}

func summarizeTasksEnergy(tasks []domain.Task) EnergySummary {
	var s EnergySummary
	for _, task := range tasks {
		switch task.EnergyType {
		case domain.EnergyProduces:
			s.Produces++
		case domain.EnergyConsumes:
			s.Consumes++
		}
	}
	return s
}

func energyNote(byLane map[domain.Lane]EnergySummary) string {
	activeLanes := []domain.Lane{
		domain.LaneTodo,
		domain.LaneDoing,
		domain.LaneParking,
		domain.LaneHalt,
	}
	active := EnergySummary{}
	for _, lane := range activeLanes {
		entry := byLane[lane]
		active.Produces += entry.Produces
		active.Consumes += entry.Consumes
	}
	if active.Consumes > active.Produces {
		return "active lanes are consume-heavy"
	}
	if active.Consumes > 0 && active.Produces == 0 {
		return "active lanes have no producing tasks"
	}
	return ""
}
