package tui

import (
	"github.com/ar1o/sonar/internal/model"
	"github.com/ar1o/sonar/internal/render"
)

// viewState tracks which view the TUI is currently displaying.
type viewState int

const (
	viewBoard  viewState = iota
	viewDetail
)

// dataLoadedMsg carries refreshed board data from the database.
type dataLoadedMsg struct {
	issues   []*model.Issue
	progress map[int]render.SubIssueProgress
	err      error
}

// issueMovedMsg signals that an issue's status was changed.
type issueMovedMsg struct {
	issueID   int
	newStatus model.Status
	err       error
}

// detailLoadedMsg carries the full detail data for a single issue.
type detailLoadedMsg struct {
	issue     *model.Issue
	subs      []*model.Issue
	relations []model.Relation
	comments  []*model.Comment
	activity  []model.Activity
	err       error
}

// tickMsg triggers a periodic data refresh in watch mode.
type tickMsg struct{}

// errMsg wraps a generic error for display.
type errMsg struct {
	err error
}
