package tui

import (
	"context"
	"fmt"
	"slices"
	"strings"

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

	filter       service.TaskFilter
	filterLabel  string
	archiveTasks []domain.Task
	overview     service.ProjectOverview

	quickInput  textinput.Model
	searchInput textinput.Model
}

func NewModel(ctx context.Context, taskSvc *service.TaskService, projectSvc *service.ProjectService, boardSvc *service.BoardService) (Model, error) {
	addInput := textinput.New()
	addInput.Placeholder = "task title or structured input"
	addInput.Focus()
	addInput.CharLimit = 220
	addInput.Width = 72

	searchInput := textinput.New()
	searchInput.Placeholder = "text search or tokens"
	searchInput.CharLimit = 220
	searchInput.Width = 72

	m := Model{
		ctx:           ctx,
		taskSvc:       taskSvc,
		projectSvc:    projectSvc,
		boardSvc:      boardSvc,
		screen:        screenBoard,
		styles:        defaultStyles(),
		status:        "Ready",
		laneSelection: make(map[domain.Lane]int, len(domain.LaneOrder)),
		quickInput:    addInput,
		searchInput:   searchInput,
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
		case screenSearch:
			return m.updateSearch(msg)
		case screenArchive:
			return m.updateArchive(msg)
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
	case screenSearch:
		return m.viewSearch()
	case screenArchive:
		return m.viewArchive()
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
		events, err := m.taskSvc.RecentEvents(m.ctx, task.ID, 16)
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
		if err := m.loadProjectOverview(task.ProjectID); err != nil {
			m.status = fmt.Sprintf("load project overview failed: %v", err)
			return m, nil
		}
		m.screen = screenProjectView
	case "?":
		m.screen = screenHelp
	case "/":
		m.searchInput.SetValue(filterToInput(m.filter, m.projects))
		m.searchInput.Focus()
		m.screen = screenSearch
	case "A":
		if err := m.reloadArchive(); err != nil {
			m.status = fmt.Sprintf("load archive failed: %v", err)
			return m, nil
		}
		m.screen = screenArchive
	case "c":
		m.filter = service.TaskFilter{}
		if err := m.reload(); err != nil {
			m.status = fmt.Sprintf("reload failed: %v", err)
			return m, nil
		}
		m.status = "Filters cleared"
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
	case "K":
		return m.reorderCurrentTask(-1)
	case "J":
		return m.reorderCurrentTask(1)
	case "e":
		m.status = "Not implemented in v2"
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
			Type:        parsed.TaskType,
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

func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenBoard
		return m, nil
	case "enter":
		filter := parseFilterInput(m.searchInput.Value(), m.projects)
		m.filter = filter
		if err := m.reload(); err != nil {
			m.status = fmt.Sprintf("reload failed: %v", err)
			return m, nil
		}
		m.status = "Filters updated"
		m.screen = screenBoard
		return m, nil
	case "c":
		m.searchInput.SetValue("")
		return m, nil
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

func (m Model) updateArchive(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.screen = screenBoard
	case "/":
		m.searchInput.SetValue(filterToInput(m.filter, m.projects))
		m.searchInput.Focus()
		m.screen = screenSearch
	case "c":
		m.filter = service.TaskFilter{}
		if err := m.reload(); err != nil {
			m.status = fmt.Sprintf("reload failed: %v", err)
			return m, nil
		}
		if err := m.reloadArchive(); err != nil {
			m.status = fmt.Sprintf("archive reload failed: %v", err)
			return m, nil
		}
		m.status = "Filters cleared"
	}
	return m, nil
}

func (m *Model) reload() error {
	board, projects, err := m.boardSvc.BoardFiltered(m.ctx, m.filter, false)
	if err != nil {
		return err
	}
	m.board = board
	m.projects = projects
	m.filterLabel = formatFilterLabel(m.filter, projects)
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
	return m.reloadArchive()
}

func (m *Model) reloadArchive() error {
	archived, _, err := m.taskSvc.ListArchivedFiltered(m.ctx, m.filter)
	if err != nil {
		return err
	}
	m.archiveTasks = archived
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
	events, err := m.taskSvc.RecentEvents(m.ctx, task.ID, 16)
	if err == nil {
		m.recentEvents = events
	}
	if m.projectViewID != "" {
		_ = m.loadProjectOverview(m.projectViewID)
	}
	m.status = "Task updated"
	return m, nil
}

func (m Model) reorderCurrentTask(direction int) (tea.Model, tea.Cmd) {
	task, ok := m.currentTask()
	if !ok {
		m.status = "No task selected"
		return m, nil
	}
	if err := m.taskSvc.ReorderInLane(m.ctx, task.ID, direction); err != nil {
		m.status = fmt.Sprintf("reorder failed: %v", err)
		return m, nil
	}
	if err := m.reload(); err != nil {
		m.status = fmt.Sprintf("reload failed: %v", err)
		return m, nil
	}
	m.status = "Task order updated"
	return m, nil
}

func (m *Model) loadProjectOverview(projectID string) error {
	overview, err := m.boardSvc.ProjectOverview(m.ctx, projectID, 12)
	if err != nil {
		return err
	}
	m.overview = overview
	return nil
}

func laneIndex(lane domain.Lane) int {
	return slices.Index(domain.LaneOrder, lane)
}

func parseFilterInput(raw string, projects map[string]domain.Project) service.TaskFilter {
	parts := splitFilterArgs(strings.TrimSpace(raw))
	filter := service.TaskFilter{}
	textParts := make([]string, 0)
	for _, part := range parts {
		switch {
		case strings.HasPrefix(part, "project:"):
			name := strings.TrimSpace(strings.TrimPrefix(part, "project:"))
			filter.ProjectID = findProjectIDByName(name, projects)
		case strings.HasPrefix(part, "initiative:"):
			v := domain.Initiative(strings.TrimSpace(strings.TrimPrefix(part, "initiative:")))
			if domain.ValidInitiative(v) {
				filter.Initiative = v
			}
		case strings.HasPrefix(part, "energy:"):
			v := domain.EnergyType(strings.TrimSpace(strings.TrimPrefix(part, "energy:")))
			if domain.ValidEnergyType(v) {
				filter.EnergyType = v
			}
		case strings.HasPrefix(part, "type:"):
			v := domain.TaskType(strings.TrimSpace(strings.TrimPrefix(part, "type:")))
			if domain.ValidTaskType(v) {
				filter.TaskType = v
			}
		default:
			textParts = append(textParts, part)
		}
	}
	filter.Query = strings.TrimSpace(strings.Join(textParts, " "))
	return filter
}

func splitFilterArgs(raw string) []string {
	if raw == "" {
		return nil
	}
	var out []string
	var b strings.Builder
	inQuotes := false
	for _, r := range raw {
		switch r {
		case '"':
			inQuotes = !inQuotes
		case ' ':
			if inQuotes {
				b.WriteRune(r)
			} else if b.Len() > 0 {
				out = append(out, b.String())
				b.Reset()
			}
		default:
			b.WriteRune(r)
		}
	}
	if b.Len() > 0 {
		out = append(out, b.String())
	}
	return out
}

func findProjectIDByName(name string, projects map[string]domain.Project) string {
	n := strings.TrimSpace(strings.ToLower(name))
	for id, p := range projects {
		if strings.ToLower(p.Name) == n {
			return id
		}
	}
	return ""
}

func formatFilterLabel(filter service.TaskFilter, projects map[string]domain.Project) string {
	parts := make([]string, 0)
	if filter.ProjectID != "" {
		project := filter.ProjectID
		if p, ok := projects[filter.ProjectID]; ok {
			project = p.Name
		}
		parts = append(parts, "project="+project)
	}
	if filter.Initiative != "" {
		parts = append(parts, "initiative="+string(filter.Initiative))
	}
	if filter.EnergyType != "" {
		parts = append(parts, "energy="+string(filter.EnergyType))
	}
	if filter.TaskType != "" {
		parts = append(parts, "type="+string(filter.TaskType))
	}
	if filter.Query != "" {
		parts = append(parts, "q="+filter.Query)
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, " ")
}

func filterToInput(filter service.TaskFilter, projects map[string]domain.Project) string {
	parts := make([]string, 0)
	if filter.ProjectID != "" {
		if p, ok := projects[filter.ProjectID]; ok {
			parts = append(parts, "project:"+p.Name)
		}
	}
	if filter.Initiative != "" {
		parts = append(parts, "initiative:"+string(filter.Initiative))
	}
	if filter.EnergyType != "" {
		parts = append(parts, "energy:"+string(filter.EnergyType))
	}
	if filter.TaskType != "" {
		parts = append(parts, "type:"+string(filter.TaskType))
	}
	if filter.Query != "" {
		parts = append(parts, filter.Query)
	}
	return strings.Join(parts, " ")
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
