package tui

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ar1o/sonar/internal/model"
	"github.com/ar1o/sonar/internal/render"
)

// Model is the top-level bubbletea model for the interactive board.
type Model struct {
	// Database
	conn *sql.DB
	cfg  BoardConfig

	// View state
	view viewState
	err  error

	// Board data
	issues         []*model.Issue
	columns        map[model.Status][]*model.Issue
	progress       map[int]render.SubIssueProgress
	activeStatuses []model.Status

	// Navigation
	colIdx  int
	cardIdx int

	// Detail view
	detailIssue    *model.Issue
	detailSubs     []*model.Issue
	detailRels     []model.Relation
	detailComments []*model.Comment
	detailActivity []model.Activity
	detailScroll   int

	// Watch mode
	watchMode    bool
	pollInterval time.Duration

	// Dimensions
	width  int
	height int

	// Status message (transient)
	statusMsg string
}

// NewModel creates a new TUI model.
func NewModel(conn *sql.DB, cfg BoardConfig, watchMode bool, pollInterval time.Duration) Model {
	return Model{
		conn:         conn,
		cfg:          cfg,
		view:         viewBoard,
		watchMode:    watchMode,
		pollInterval: pollInterval,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		loadBoardData(m.conn, m.cfg),
	}
	if m.watchMode {
		cmds = append(cmds, tickCmd(m.pollInterval))
	}
	return tea.Batch(cmds...)
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case dataLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		m.statusMsg = ""
		m.issues = msg.issues
		m.progress = msg.progress
		m.columns = render.GroupByStatus(msg.issues)
		m.activeStatuses = activeStatuses(m.columns)
		m.clampCursor()
		return m, nil

	case issueMovedMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
			return m, nil
		}
		m.statusMsg = fmt.Sprintf("Moved %s to %s", model.FormatID(msg.issueID), msg.newStatus)
		return m, loadBoardData(m.conn, m.cfg)

	case detailLoadedMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
			m.view = viewBoard
			return m, nil
		}
		m.detailIssue = msg.issue
		m.detailSubs = msg.subs
		m.detailRels = msg.relations
		m.detailComments = msg.comments
		m.detailActivity = msg.activity
		m.detailScroll = 0
		m.view = viewDetail
		return m, nil

	case tickMsg:
		cmds := []tea.Cmd{
			loadBoardData(m.conn, m.cfg),
		}
		if m.watchMode {
			cmds = append(cmds, tickCmd(m.pollInterval))
		}
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

// handleKey dispatches key events based on the current view.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit
	if key.Matches(msg, keys.Quit) {
		if m.view == viewDetail {
			m.view = viewBoard
			m.detailIssue = nil
			return m, nil
		}
		return m, tea.Quit
	}

	switch m.view {
	case viewBoard:
		return m.handleBoardKey(msg)
	case viewDetail:
		return m.handleDetailKey(msg)
	}

	return m, nil
}

// handleBoardKey handles key events in board view.
func (m Model) handleBoardKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Left):
		if len(m.activeStatuses) > 0 {
			m.colIdx--
			if m.colIdx < 0 {
				m.colIdx = len(m.activeStatuses) - 1
			}
			m.clampCardIdx()
		}
		return m, nil

	case key.Matches(msg, keys.Right):
		if len(m.activeStatuses) > 0 {
			m.colIdx++
			if m.colIdx >= len(m.activeStatuses) {
				m.colIdx = 0
			}
			m.clampCardIdx()
		}
		return m, nil

	case key.Matches(msg, keys.Up):
		m.cardIdx--
		if m.cardIdx < 0 {
			col := m.currentColumnIssues()
			if len(col) > 0 {
				m.cardIdx = len(col) - 1
			} else {
				m.cardIdx = 0
			}
		}
		return m, nil

	case key.Matches(msg, keys.Down):
		col := m.currentColumnIssues()
		m.cardIdx++
		if len(col) == 0 || m.cardIdx >= len(col) {
			m.cardIdx = 0
		}
		return m, nil

	case key.Matches(msg, keys.MoveLeft):
		return m.moveCard(-1)

	case key.Matches(msg, keys.MoveRight):
		return m.moveCard(1)

	case key.Matches(msg, keys.Select):
		issue := m.selectedIssue()
		if issue != nil {
			return m, loadIssueDetail(m.conn, issue.ID)
		}
		return m, nil

	case key.Matches(msg, keys.Refresh):
		m.statusMsg = "Refreshing..."
		return m, loadBoardData(m.conn, m.cfg)

	case key.Matches(msg, keys.Help):
		// Could toggle expanded help; for now just clear status
		m.statusMsg = ""
		return m, nil
	}

	return m, nil
}

// handleDetailKey handles key events in detail view.
func (m Model) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Back):
		m.view = viewBoard
		m.detailIssue = nil
		return m, nil

	case key.Matches(msg, keys.Up):
		if m.detailScroll > 0 {
			m.detailScroll--
		}
		return m, nil

	case key.Matches(msg, keys.Down):
		m.detailScroll++
		return m, nil

	case key.Matches(msg, keys.MoveLeft):
		if m.detailIssue != nil {
			return m.moveDetailCard(-1)
		}
		return m, nil

	case key.Matches(msg, keys.MoveRight):
		if m.detailIssue != nil {
			return m.moveDetailCard(1)
		}
		return m, nil
	}

	return m, nil
}

// moveCard moves the selected card to an adjacent status column.
func (m Model) moveCard(direction int) (tea.Model, tea.Cmd) {
	issue := m.selectedIssue()
	if issue == nil {
		return m, nil
	}

	newStatus := adjacentStatus(issue.Status, direction)
	if newStatus == issue.Status {
		return m, nil
	}

	return m, moveIssue(m.conn, issue.ID, newStatus)
}

// moveDetailCard moves the detail view's issue to an adjacent status.
func (m Model) moveDetailCard(direction int) (tea.Model, tea.Cmd) {
	newStatus := adjacentStatus(m.detailIssue.Status, direction)
	if newStatus == m.detailIssue.Status {
		return m, nil
	}

	return m, tea.Batch(
		moveIssue(m.conn, m.detailIssue.ID, newStatus),
		loadIssueDetail(m.conn, m.detailIssue.ID),
	)
}

// adjacentStatus returns the status that is `direction` steps away in StatusOrder.
func adjacentStatus(current model.Status, direction int) model.Status {
	for i, s := range render.StatusOrder {
		if s == current {
			next := i + direction
			if next < 0 || next >= len(render.StatusOrder) {
				return current
			}
			return render.StatusOrder[next]
		}
	}
	return current
}

// activeStatuses returns the statuses that have issues, in StatusOrder.
func activeStatuses(columns map[model.Status][]*model.Issue) []model.Status {
	var result []model.Status
	for _, s := range render.StatusOrder {
		if len(columns[s]) > 0 {
			result = append(result, s)
		}
	}
	return result
}

// currentColumnIssues returns the issues in the currently selected column.
func (m Model) currentColumnIssues() []*model.Issue {
	if len(m.activeStatuses) == 0 || m.colIdx >= len(m.activeStatuses) {
		return nil
	}
	return m.columns[m.activeStatuses[m.colIdx]]
}

// selectedIssue returns the currently selected issue, or nil.
func (m Model) selectedIssue() *model.Issue {
	col := m.currentColumnIssues()
	if len(col) == 0 || m.cardIdx >= len(col) {
		return nil
	}
	return col[m.cardIdx]
}

// clampCursor ensures cursor positions are within valid bounds.
func (m *Model) clampCursor() {
	if len(m.activeStatuses) == 0 {
		m.colIdx = 0
		m.cardIdx = 0
		return
	}
	if m.colIdx >= len(m.activeStatuses) {
		m.colIdx = len(m.activeStatuses) - 1
	}
	m.clampCardIdx()
}

// clampCardIdx ensures cardIdx is within the current column's bounds.
func (m *Model) clampCardIdx() {
	col := m.currentColumnIssues()
	if len(col) == 0 {
		m.cardIdx = 0
	} else if m.cardIdx >= len(col) {
		m.cardIdx = len(col) - 1
	}
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var content string
	switch m.view {
	case viewBoard:
		content = renderBoardView(m)
	case viewDetail:
		content = renderDetailView(m)
	}

	// Status/error bar
	var statusLine string
	if m.err != nil {
		statusLine = errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	} else if m.statusMsg != "" {
		statusLine = statusBarStyle.Render(m.statusMsg)
	}

	// Help bar
	helpBar := renderHelpBar(m.view, m.width)

	// Compose: content + status + help
	lines := strings.Split(content, "\n")

	// Reserve 2 lines for status + help (or 1 if no status)
	reserved := 1
	if statusLine != "" {
		reserved = 2
	}
	maxContentLines := m.height - reserved
	if maxContentLines < 1 {
		maxContentLines = 1
	}
	if len(lines) > maxContentLines {
		lines = lines[:maxContentLines]
	}

	result := strings.Join(lines, "\n")
	if statusLine != "" {
		result += "\n" + statusLine
	}
	result += "\n" + helpBar

	return result
}
