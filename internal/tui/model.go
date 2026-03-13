package tui

import (
	"context"
	"fmt"
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

	filter           service.TaskFilter
	filterLabel      string
	boardEnergy      service.BoardEnergySummary
	archiveTasks     []domain.Task
	archiveIdx       int
	overview         service.ProjectOverview
	constellation    service.ProjectConstellation
	constellationIdx int

	quickInput    textinput.Model
	quickHistory  quickAddHistory
	searchInput   textinput.Model
	editInput     textinput.Model
	editingTaskID string
	editOriginal  service.EditTaskInput
	editDraft     service.EditTaskInput
	editFieldIdx  int
	editExitArmed bool
	pendingTaskID string
	pendingAction *domain.TaskAction
	helpScroll    int
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

	editInput := textinput.New()
	editInput.Placeholder = "edit value"
	editInput.CharLimit = 220
	editInput.Width = 72

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
		editInput:     editInput,
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
		case screenEdit:
			return m.updateEdit(msg)
		case screenTaskDetail:
			return m.updateTaskDetail(msg)
		case screenProjectView:
			if msg.String() == "esc" {
				m.screen = screenBoard
			}
			return m, nil
		case screenConstellation:
			return m.updateConstellation(msg)
		case screenHelp:
			return m.updateHelp(msg)
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
	case screenConstellation:
		return m.viewConstellation()
	case screenHelp:
		return m.viewHelp()
	case screenSearch:
		return m.viewSearch()
	case screenArchive:
		return m.viewArchive()
	case screenEdit:
		return m.viewEdit()
	default:
		return m.viewBoard()
	}
}

func (m Model) updateBoard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.pendingAction != nil {
		switch msg.String() {
		case "y":
			return m.confirmPendingAction()
		case "esc":
			m.pendingAction = nil
			m.pendingTaskID = ""
			m.status = "Action canceled"
			return m, nil
		default:
			return m, nil
		}
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q":
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
		m.helpScroll = 0
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
	case "C":
		if err := m.loadConstellation(); err != nil {
			m.status = fmt.Sprintf("load constellation failed: %v", err)
			return m, nil
		}
		m.screen = screenConstellation
	case "c":
		m.filter = service.TaskFilter{}
		if err := m.reload(); err != nil {
			m.status = fmt.Sprintf("reload failed: %v", err)
			return m, nil
		}
		m.status = "Filters cleared"
	case "t":
		return m.applyTaskAction("touch", func(id string) error { return m.taskSvc.Touch(m.ctx, id) })
	case "K":
		return m.reorderCurrentTask(-1)
	case "J":
		return m.reorderCurrentTask(1)
	default:
		if isSingleKey(msg.String()) {
			return m.applyLaneActionByKey(msg.String())
		}
	}
	return m, nil
}

func (m Model) updateTaskDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.pendingAction != nil {
		switch msg.String() {
		case "y":
			return m.confirmPendingAction()
		case "esc":
			m.pendingAction = nil
			m.pendingTaskID = ""
			m.status = "Action canceled"
			return m, nil
		default:
			return m, nil
		}
	}

	switch msg.String() {
	case "esc":
		m.screen = screenBoard
	case "t":
		return m.applyTaskAction("touch", func(id string) error { return m.taskSvc.Touch(m.ctx, id) })
	default:
		if isSingleKey(msg.String()) {
			return m.applyLaneActionByKey(msg.String())
		}
	}
	return m, nil
}

func (m Model) updateQuickAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenBoard
		return m, nil
	case "up":
		m.quickInput.SetValue(m.quickHistory.Prev(m.quickInput.Value()))
		m.quickInput.CursorEnd()
		return m, nil
	case "down":
		m.quickInput.SetValue(m.quickHistory.Next(m.quickInput.Value()))
		m.quickInput.CursorEnd()
		return m, nil
	case "ctrl+n":
		updated, ok := m.autocompleteQuickAdd()
		if ok {
			m.quickInput.SetValue(updated)
			m.quickInput.CursorEnd()
			m.quickErr = ""
		} else {
			m.status = "No autocomplete suggestion"
		}
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
		m.quickHistory.Add(raw)
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

func (m Model) updateEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editExitArmed {
		switch msg.String() {
		case "esc":
			m.screen = screenBoard
			m.editingTaskID = ""
			m.editOriginal = service.EditTaskInput{}
			m.editDraft = service.EditTaskInput{}
			m.editExitArmed = false
			m.status = "Edit canceled"
			return m, nil
		default:
			m.editExitArmed = false
		}
	}

	switch msg.String() {
	case "esc":
		if m.editFormDirty() {
			m.editExitArmed = true
			m.status = "Unsaved edits will be lost. Press Esc again to discard."
			return m, nil
		}
		m.screen = screenBoard
		m.editingTaskID = ""
		m.editOriginal = service.EditTaskInput{}
		m.editDraft = service.EditTaskInput{}
		m.editExitArmed = false
		return m, nil
	case "shift+tab":
		m.applyEditFieldValue(m.editFieldKey(), m.editInput.Value())
		m.editFieldIdx = max(0, m.editFieldIdx-1)
		m.editInput.SetValue(m.currentEditFieldValue())
		m.editInput.CursorEnd()
		return m, nil
	case "tab":
		m.applyEditFieldValue(m.editFieldKey(), m.editInput.Value())
		m.editFieldIdx = min(len(editFieldKeys())-1, m.editFieldIdx+1)
		m.editInput.SetValue(m.currentEditFieldValue())
		m.editInput.CursorEnd()
		return m, nil
	case "ctrl+n":
		m.applyEditFieldValue(m.editFieldKey(), m.editInput.Value())
		value, ok := m.autocompleteEditField()
		if ok {
			m.applyEditFieldValue(m.editFieldKey(), value)
			m.editInput.SetValue(value)
			m.editInput.CursorEnd()
		} else {
			m.status = "No autocomplete suggestion"
		}
		return m, nil
	case "enter":
		if m.editingTaskID == "" {
			m.status = "No task selected for edit"
			m.screen = screenBoard
			return m, nil
		}
		m.applyEditFieldValue(m.editFieldKey(), m.editInput.Value())
		if err := m.taskSvc.Edit(m.ctx, m.editingTaskID, m.editDraft); err != nil {
			m.status = fmt.Sprintf("edit failed: %v", err)
			return m, nil
		}
		m.screen = screenBoard
		m.editingTaskID = ""
		m.editOriginal = service.EditTaskInput{}
		m.editDraft = service.EditTaskInput{}
		m.editExitArmed = false
		if err := m.reload(); err != nil {
			m.status = fmt.Sprintf("reload failed: %v", err)
			return m, nil
		}
		m.status = "Task edited"
		return m, nil
	}
	var cmd tea.Cmd
	m.editInput, cmd = m.editInput.Update(msg)
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
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

func (m Model) updateHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenBoard
	case "j":
		m.helpScroll++
	case "k":
		m.helpScroll = max(0, m.helpScroll-1)
	}
	return m, nil
}

func (m Model) updateConstellation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = screenBoard
	case "j":
		total := len(m.constellationProjects())
		if total > 0 {
			m.constellationIdx = min(total-1, m.constellationIdx+1)
		}
	case "k":
		m.constellationIdx = max(0, m.constellationIdx-1)
	case "enter":
		project, ok := m.currentConstellationProject()
		if !ok {
			m.status = "No project selected"
			return m, nil
		}
		m.projectViewID = project.Project.ID
		if err := m.loadProjectOverview(project.Project.ID); err != nil {
			m.status = fmt.Sprintf("load project overview failed: %v", err)
			return m, nil
		}
		m.screen = screenProjectView
	}
	return m, nil
}

func (m Model) updateArchive(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.pendingAction != nil {
		switch msg.String() {
		case "y":
			return m.confirmPendingAction()
		case "esc":
			m.pendingAction = nil
			m.pendingTaskID = ""
			m.status = "Action canceled"
			return m, nil
		default:
			return m, nil
		}
	}

	switch msg.String() {
	case "esc":
		m.screen = screenBoard
	case "j":
		if len(m.archiveTasks) > 0 {
			m.archiveIdx = min(len(m.archiveTasks)-1, m.archiveIdx+1)
		}
	case "k":
		m.archiveIdx = max(0, m.archiveIdx-1)
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
	default:
		if isSingleKey(msg.String()) {
			return m.applyArchiveActionByKey(msg.String())
		}
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
	m.boardEnergy = service.BuildBoardEnergySummary(board)
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
	if len(m.archiveTasks) == 0 {
		m.archiveIdx = 0
	} else if m.archiveIdx >= len(m.archiveTasks) {
		m.archiveIdx = len(m.archiveTasks) - 1
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

func (m Model) currentArchiveTask() (domain.Task, bool) {
	if len(m.archiveTasks) == 0 {
		return domain.Task{}, false
	}
	idx := m.archiveIdx
	if idx < 0 {
		idx = 0
	}
	if idx >= len(m.archiveTasks) {
		idx = len(m.archiveTasks) - 1
	}
	return m.archiveTasks[idx], true
}

func (m Model) currentTaskActions() []domain.TaskAction {
	task, ok := m.currentTask()
	if !ok {
		return nil
	}
	return m.taskSvc.AllowedActionsForTask(task)
}

func (m Model) taskActions(task domain.Task) []domain.TaskAction {
	return m.taskSvc.AllowedActionsForTask(task)
}

func (m Model) applyLaneActionByKey(key string) (tea.Model, tea.Cmd) {
	task, ok := m.currentTask()
	if !ok {
		m.status = "No task selected"
		return m, nil
	}
	actions := m.taskActions(task)
	return m.applyActionByKey(task, actions, key)
}

func (m Model) applyArchiveActionByKey(key string) (tea.Model, tea.Cmd) {
	task, ok := m.currentArchiveTask()
	if !ok {
		m.status = "No archived task selected"
		return m, nil
	}
	actions := m.taskActions(task)
	return m.applyActionByKey(task, actions, key)
}

func (m Model) applyActionByKey(task domain.Task, actions []domain.TaskAction, key string) (tea.Model, tea.Cmd) {
	if len(actions) == 0 {
		m.status = "No valid actions for selected task"
		return m, nil
	}
	for _, action := range actions {
		if action.Shortcut == key {
			return m.runTaskAction(task, action, false)
		}
	}
	return m, nil
}

func (m Model) confirmPendingAction() (tea.Model, tea.Cmd) {
	if m.pendingAction == nil || m.pendingTaskID == "" {
		m.pendingAction = nil
		m.pendingTaskID = ""
		return m, nil
	}
	task, ok, err := m.taskSvc.Get(m.ctx, m.pendingTaskID)
	if err != nil {
		m.status = fmt.Sprintf("confirm failed: %v", err)
		m.pendingAction = nil
		m.pendingTaskID = ""
		return m, nil
	}
	if !ok {
		m.status = "Task no longer exists"
		m.pendingAction = nil
		m.pendingTaskID = ""
		return m, nil
	}
	action := *m.pendingAction
	return m.runTaskAction(task, action, true)
}

func (m Model) runTaskAction(task domain.Task, action domain.TaskAction, confirmed bool) (tea.Model, tea.Cmd) {
	if action.RequiresConfirmation && !confirmed {
		act := action
		m.pendingAction = &act
		m.pendingTaskID = task.ID
		m.status = fmt.Sprintf("Confirm '%s' on '%s' with y, Esc to cancel", action.Label, task.Title)
		return m, nil
	}

	var err error
	switch action.Kind {
	case domain.ActionMove:
		err = m.taskSvc.Transition(m.ctx, task.ID, action.To, confirmed)
	case domain.ActionEdit:
		m.startEdit(task)
		m.pendingAction = nil
		m.pendingTaskID = ""
		return m, nil
	case domain.ActionDelete:
		err = m.taskSvc.Delete(m.ctx, task.ID, confirmed)
	default:
		err = fmt.Errorf("unsupported action: %s", action.Kind)
	}
	if err != nil {
		m.pendingAction = nil
		m.pendingTaskID = ""
		m.status = fmt.Sprintf("action failed: %v", err)
		return m, nil
	}

	m.pendingAction = nil
	m.pendingTaskID = ""
	if err := m.reload(); err != nil {
		m.status = fmt.Sprintf("reload failed: %v", err)
		return m, nil
	}
	events, evtErr := m.taskSvc.RecentEvents(m.ctx, task.ID, 16)
	if evtErr == nil {
		m.recentEvents = events
	}
	if m.projectViewID != "" {
		_ = m.loadProjectOverview(m.projectViewID)
	}
	m.status = "Action applied"
	return m, nil
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

func (m *Model) loadConstellation() error {
	constellation, err := m.boardSvc.ProjectConstellation(m.ctx)
	if err != nil {
		return err
	}
	m.constellation = constellation
	total := len(m.constellationProjects())
	if total == 0 {
		m.constellationIdx = 0
		return nil
	}
	if m.constellationIdx >= total {
		m.constellationIdx = total - 1
	}
	if m.constellationIdx < 0 {
		m.constellationIdx = 0
	}
	return nil
}

func parseFilterInput(raw string, projects map[string]domain.Project) service.TaskFilter {
	parts := splitFilterArgs(strings.TrimSpace(raw))
	filter := service.TaskFilter{}
	textParts := make([]string, 0)
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		key := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(part)), "/")

		if strings.Contains(key, ":") {
			kv := strings.SplitN(key, ":", 2)
			applyFilterKeyValue(&filter, kv[0], kv[1], projects)
			continue
		}

		switch key {
		case "project", "initiative", "energy", "type":
			if i+1 < len(parts) {
				applyFilterKeyValue(&filter, key, parts[i+1], projects)
				i++
			}
		default:
			textParts = append(textParts, part)
		}
	}
	filter.Query = strings.TrimSpace(strings.Join(textParts, " "))
	return filter
}

func applyFilterKeyValue(filter *service.TaskFilter, key string, value string, projects map[string]domain.Project) {
	v := strings.TrimSpace(value)
	switch key {
	case "project":
		filter.ProjectID = findProjectIDByName(v, projects)
	case "initiative":
		initiative := domain.Initiative(strings.ToLower(v))
		if domain.ValidInitiative(initiative) {
			filter.Initiative = initiative
		}
	case "energy":
		energy := domain.EnergyType(strings.ToLower(v))
		if domain.ValidEnergyType(energy) {
			filter.EnergyType = energy
		}
	case "type":
		taskType := domain.TaskType(strings.ToLower(v))
		if domain.ValidTaskType(taskType) {
			filter.TaskType = taskType
		}
	}
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

func (m Model) actionHintForActions(actions []domain.TaskAction) string {
	if len(actions) == 0 {
		return "actions: none"
	}
	parts := make([]string, 0, len(actions))
	for _, a := range actions {
		label := a.Label
		if a.RequiresConfirmation {
			label += " (confirm)"
		}
		parts = append(parts, fmt.Sprintf("%s %s", a.Shortcut, label))
	}
	return "actions " + strings.Join(parts, "  ")
}

func isSingleKey(s string) bool {
	return len(s) == 1
}

func (m Model) hasUnsavedInput() bool {
	switch m.screen {
	case screenQuickAdd:
		return strings.TrimSpace(m.quickInput.Value()) != ""
	case screenSearch:
		return strings.TrimSpace(m.searchInput.Value()) != ""
	case screenEdit:
		return m.editFormDirty()
	default:
		return false
	}
}

func (m Model) actionHint() string {
	return m.actionHintForActions(m.currentTaskActions())
}

func (m Model) constellationProjects() []service.ConstellationProject {
	out := make([]service.ConstellationProject, 0)
	for _, group := range m.constellation.Groups {
		out = append(out, group.Projects...)
	}
	return out
}

func (m Model) currentConstellationProject() (service.ConstellationProject, bool) {
	projects := m.constellationProjects()
	if len(projects) == 0 {
		return service.ConstellationProject{}, false
	}
	idx := m.constellationIdx
	if idx < 0 {
		idx = 0
	}
	if idx >= len(projects) {
		idx = len(projects) - 1
	}
	return projects[idx], true
}

func (m *Model) startEdit(task domain.Task) {
	projectName := "general"
	if p, ok := m.projects[task.ProjectID]; ok {
		projectName = p.Name
	}
	m.editingTaskID = task.ID
	m.editOriginal = service.EditTaskInput{
		Title:       task.Title,
		Description: task.Description,
		ProjectName: projectName,
		Initiative:  task.Initiative,
		Type:        task.Type,
		Loop:        task.Loop,
		EnergyType:  task.EnergyType,
		Nature:      task.Nature,
	}
	m.editDraft = m.editOriginal
	m.editFieldIdx = 0
	m.editExitArmed = false
	m.editInput.SetValue(m.currentEditFieldValue())
	m.editInput.Focus()
	m.editInput.CursorEnd()
	m.screen = screenEdit
}

func editFieldKeys() []string {
	return []string{
		"title",
		"description",
		"project",
		"initiative",
		"type",
		"loop",
		"energy",
		"nature",
	}
}

func editFieldLabel(key string) string {
	switch key {
	case "title":
		return "title"
	case "description":
		return "description"
	case "project":
		return "project"
	case "initiative":
		return "initiative"
	case "type":
		return "type"
	case "loop":
		return "loop"
	case "energy":
		return "energy_type"
	case "nature":
		return "nature"
	default:
		return key
	}
}

func (m Model) editFieldKey() string {
	keys := editFieldKeys()
	if m.editFieldIdx < 0 {
		return keys[0]
	}
	if m.editFieldIdx >= len(keys) {
		return keys[len(keys)-1]
	}
	return keys[m.editFieldIdx]
}

func (m Model) currentEditFieldValue() string {
	switch m.editFieldKey() {
	case "title":
		return m.editDraft.Title
	case "description":
		return m.editDraft.Description
	case "project":
		return m.editDraft.ProjectName
	case "initiative":
		return string(m.editDraft.Initiative)
	case "type":
		return string(m.editDraft.Type)
	case "loop":
		return string(m.editDraft.Loop)
	case "energy":
		return string(m.editDraft.EnergyType)
	case "nature":
		return string(m.editDraft.Nature)
	default:
		return ""
	}
}

func (m *Model) applyEditFieldValue(key, value string) {
	v := strings.TrimSpace(value)
	switch key {
	case "title":
		m.editDraft.Title = v
	case "description":
		m.editDraft.Description = v
	case "project":
		m.editDraft.ProjectName = v
	case "initiative":
		m.editDraft.Initiative = domain.Initiative(strings.ToLower(v))
	case "type":
		m.editDraft.Type = domain.TaskType(strings.ToLower(v))
	case "loop":
		m.editDraft.Loop = domain.Loop(strings.ToLower(v))
	case "energy":
		m.editDraft.EnergyType = domain.EnergyType(strings.ToLower(v))
	case "nature":
		m.editDraft.Nature = domain.Nature(strings.ToLower(v))
	}
}

func (m Model) editFormDirty() bool {
	if strings.TrimSpace(m.editInput.Value()) != strings.TrimSpace(m.currentEditFieldValue()) {
		return true
	}
	return strings.TrimSpace(m.editDraft.Title) != strings.TrimSpace(m.editOriginal.Title) ||
		strings.TrimSpace(m.editDraft.Description) != strings.TrimSpace(m.editOriginal.Description) ||
		strings.TrimSpace(m.editDraft.ProjectName) != strings.TrimSpace(m.editOriginal.ProjectName) ||
		m.editDraft.Initiative != m.editOriginal.Initiative ||
		m.editDraft.Type != m.editOriginal.Type ||
		m.editDraft.Loop != m.editOriginal.Loop ||
		m.editDraft.EnergyType != m.editOriginal.EnergyType ||
		m.editDraft.Nature != m.editOriginal.Nature
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
