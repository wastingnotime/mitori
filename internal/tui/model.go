package tui

import (
	"context"
	"fmt"
	"slices"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/wastingnotime/mitori/internal/domain"
	"github.com/wastingnotime/mitori/internal/parser"
	"github.com/wastingnotime/mitori/internal/service"
)

type Model struct {
	ctx        context.Context
	taskSvc    *service.TaskService
	projectSvc *service.ProjectService
	boardSvc   *service.BoardService

	screen   screen
	width    int
	height   int
	styles   styles
	status   string
	quickErr string

	board         map[domain.Lane][]domain.Task
	projects      map[string]domain.Project
	laneIdx       int
	laneSelection map[domain.Lane]int
	recentEvents  []domain.Event
	projectViewID string

	quickInput textinput.Model
}

func NewModel(ctx context.Context, taskSvc *service.TaskService, projectSvc *service.ProjectService, boardSvc *service.BoardService) (Model, error) {
	input := textinput.New()
	input.Placeholder = "task title or structured input"
	input.Focus()
	input.CharLimit = 220
	input.Width = 72

	m := Model{
		ctx:           ctx,
		taskSvc:       taskSvc,
		projectSvc:    projectSvc,
		boardSvc:      boardSvc,
		screen:        screenBoard,
		styles:        defaultStyles(),
		status:        "Ready",
		laneSelection: make(map[domain.Lane]int, len(domain.LaneOrder)),
		quickInput:    input,
	}
	for _, lane := range domain.LaneOrder {
		m.laneSelection[lane] = 0
	}
	if err := m.reload(); err != nil {
		return Model{}, err
	}
	return m, nil
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch m.screen {
		case screenQuickAdd:
			return m.updateQuickAdd(msg)
		case screenTaskDetail:
			return m.updateTaskDetail(msg)
		case screenProjectView, screenHelp:
			if msg.String() == "q" || msg.String() == "esc" {
				m.screen = screenBoard
			}
			return m, nil
		default:
			return m.updateBoard(msg)
		}
	}
	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case screenTaskDetail:
		return m.viewTaskDetail()
	case screenQuickAdd:
		return m.viewQuickAdd()
	case screenProjectView:
		return m.viewProject()
	case screenHelp:
		return m.viewHelp()
	default:
		return m.viewBoard()
	}
}

func (m Model) updateBoard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case KeyLeft:
		m.laneIdx = max(0, m.laneIdx-1)
	case KeyRight:
		m.laneIdx = min(len(domain.LaneOrder)-1, m.laneIdx+1)
	case KeyUp:
		lane := domain.LaneOrder[m.laneIdx]
		m.laneSelection[lane] = max(0, m.laneSelection[lane]-1)
	case KeyDown:
		lane := domain.LaneOrder[m.laneIdx]
		maxIdx := len(m.board[lane]) - 1
		m.laneSelection[lane] = min(maxIdx, m.laneSelection[lane]+1)
	case "enter":
		task, ok := m.currentTask()
		if !ok {
			m.status = "No task selected"
			return m, nil
		}
		events, err := m.taskSvc.RecentEvents(m.ctx, task.ID, 8)
		if err != nil {
			m.status = fmt.Sprintf("load events failed: %v", err)
			return m, nil
		}
		m.recentEvents = events
		m.screen = screenTaskDetail
	case "a":
		m.resetQuickAdd()
		m.quickInput.Focus()
		m.screen = screenQuickAdd
	case "g":
		task, ok := m.currentTask()
		if !ok {
			m.status = "No task selected"
			return m, nil
		}
		m.projectViewID = task.ProjectID
		m.screen = screenProjectView
	case "?":
		m.screen = screenHelp
	case "p":
		return m.applyTaskAction("park/unpark", func(id string) error { return m.taskSvc.ToggleParking(m.ctx, id) })
	case "t":
		return m.applyTaskAction("touch", func(id string) error { return m.taskSvc.Touch(m.ctx, id) })
	case "d":
		return m.applyTaskAction("done", func(id string) error { return m.taskSvc.MarkDone(m.ctx, id) })
	case "b":
		return m.applyTaskAction("backlog", func(id string) error { return m.taskSvc.SendBacklog(m.ctx, id) })
	case "x":
		return m.applyTaskAction("archive", func(id string) error { return m.taskSvc.Archive(m.ctx, id) })
	case "m":
		task, ok := m.currentTask()
		if !ok {
			m.status = "No task selected"
			return m, nil
		}
		nextLane := domain.LaneOrder[(laneIndex(task.Lane)+1)%len(domain.LaneOrder)]
		if err := m.taskSvc.MoveLane(m.ctx, task.ID, nextLane); err != nil {
			m.status = fmt.Sprintf("move failed: %v", err)
			return m, nil
		}
		if err := m.reload(); err != nil {
			m.status = fmt.Sprintf("reload failed: %v", err)
			return m, nil
		}
		m.status = "Task moved"
	case "e", "/":
		m.status = "Not implemented in v1"
	}
	return m, nil
}

func (m Model) updateTaskDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.screen = screenBoard
	case "p":
		return m.applyTaskAction("park/unpark", func(id string) error { return m.taskSvc.ToggleParking(m.ctx, id) })
	case "t":
		return m.applyTaskAction("touch", func(id string) error { return m.taskSvc.Touch(m.ctx, id) })
	case "d":
		return m.applyTaskAction("done", func(id string) error { return m.taskSvc.MarkDone(m.ctx, id) })
	case "b":
		return m.applyTaskAction("backlog", func(id string) error { return m.taskSvc.SendBacklog(m.ctx, id) })
	case "x":
		return m.applyTaskAction("archive", func(id string) error { return m.taskSvc.Archive(m.ctx, id) })
	}
	return m, nil
}

func (m Model) updateQuickAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenBoard
		return m, nil
	case "enter":
		raw := m.quickInput.Value()
		parsed := parser.ParseTaskInput(raw)
		in := service.CreateTaskInput{
			Title:       parsed.Title,
			Initiative:  parsed.Initiative,
			ProjectName: parsed.Project,
			Type:        domain.TaskTypeChore,
			Loop:        domain.LoopProduct,
			EnergyType:  domain.EnergyProduces,
			Nature:      domain.NatureMushin,
			Lane:        domain.LaneBacklog,
		}
		task, err := m.taskSvc.Create(m.ctx, in)
		if err != nil {
			m.setQuickError(err)
			return m, nil
		}
		m.screen = screenBoard
		m.status = "Task added: " + task.Title
		if err := m.reload(); err != nil {
			m.status = fmt.Sprintf("reload failed: %v", err)
			return m, nil
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.quickInput, cmd = m.quickInput.Update(msg)
	return m, cmd
}

func (m *Model) reload() error {
	board, projects, err := m.boardSvc.Board(m.ctx)
	if err != nil {
		return err
	}
	m.board = board
	m.projects = projects
	for _, lane := range domain.LaneOrder {
		maxIdx := len(m.board[lane]) - 1
		if maxIdx < 0 {
			m.laneSelection[lane] = 0
			continue
		}
		if m.laneSelection[lane] > maxIdx {
			m.laneSelection[lane] = maxIdx
		}
	}
	return nil
}

func (m Model) selectedCardIndex(lane domain.Lane) int {
	idx := m.laneSelection[lane]
	if idx < 0 {
		return 0
	}
	return idx
}

func (m Model) currentTask() (domain.Task, bool) {
	if len(domain.LaneOrder) == 0 {
		return domain.Task{}, false
	}
	lane := domain.LaneOrder[m.laneIdx]
	tasks := m.board[lane]
	if len(tasks) == 0 {
		return domain.Task{}, false
	}
	idx := m.selectedCardIndex(lane)
	if idx >= len(tasks) {
		idx = len(tasks) - 1
	}
	return tasks[idx], true
}

func (m Model) applyTaskAction(name string, action func(taskID string) error) (tea.Model, tea.Cmd) {
	task, ok := m.currentTask()
	if !ok {
		m.status = "No task selected"
		return m, nil
	}
	if err := action(task.ID); err != nil {
		m.status = fmt.Sprintf("%s failed: %v", name, err)
		return m, nil
	}
	if err := m.reload(); err != nil {
		m.status = fmt.Sprintf("reload failed: %v", err)
		return m, nil
	}
	events, err := m.taskSvc.RecentEvents(m.ctx, task.ID, 8)
	if err == nil {
		m.recentEvents = events
	}
	m.status = "Task updated"
	return m, nil
}

func laneIndex(lane domain.Lane) int {
	return slices.Index(domain.LaneOrder, lane)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
