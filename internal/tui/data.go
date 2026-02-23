package tui

import (
	"database/sql"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ar1o/sonar/internal/config"
	"github.com/ar1o/sonar/internal/db"
	"github.com/ar1o/sonar/internal/model"
	"github.com/ar1o/sonar/internal/render"
)

// BoardConfig holds filtering options passed to the TUI from CLI flags.
type BoardConfig struct {
	Priorities []string
	Labels     []string
	Assignee   string
	Expand     bool
}

// loadBoardData fetches all issues and sub-issue progress from the database.
func loadBoardData(conn *sql.DB, cfg BoardConfig) tea.Cmd {
	return func() tea.Msg {
		opts := db.ListOptions{
			Priorities:  cfg.Priorities,
			Labels:      cfg.Labels,
			Assignee:    cfg.Assignee,
			IncludeDone: true,
		}

		issues, _, err := db.ListIssues(conn, opts)
		if err != nil {
			return dataLoadedMsg{err: err}
		}

		// Hydrate labels and files.
		if err := db.HydrateLabels(conn, issues); err != nil {
			return dataLoadedMsg{err: err}
		}
		if err := db.HydrateFiles(conn, issues); err != nil {
			return dataLoadedMsg{err: err}
		}

		// Filter to roots if not expanded.
		if !cfg.Expand {
			var roots []*model.Issue
			for _, issue := range issues {
				if issue.ParentID == nil {
					roots = append(roots, issue)
				}
			}
			issues = roots
		}

		// Build progress map.
		parentIDs := make([]int, len(issues))
		for i, issue := range issues {
			parentIDs[i] = issue.ID
		}

		batchProgress, err := db.GetBatchSubIssueProgress(conn, parentIDs)
		if err != nil {
			return dataLoadedMsg{err: err}
		}

		progress := make(map[int]render.SubIssueProgress, len(batchProgress))
		for id, counts := range batchProgress {
			if counts[1] > 0 {
				progress[id] = render.SubIssueProgress{Done: counts[0], Total: counts[1]}
			}
		}

		return dataLoadedMsg{
			issues:   issues,
			progress: progress,
		}
	}
}

// moveIssue updates an issue's status in the database.
func moveIssue(conn *sql.DB, issueID int, newStatus model.Status) tea.Cmd {
	return func() tea.Msg {
		err := db.UpdateIssue(conn, issueID, map[string]interface{}{
			"status": string(newStatus),
		}, config.DefaultAuthor())

		return issueMovedMsg{
			issueID:   issueID,
			newStatus: newStatus,
			err:       err,
		}
	}
}

// loadIssueDetail fetches full detail data for a single issue.
func loadIssueDetail(conn *sql.DB, issueID int) tea.Cmd {
	return func() tea.Msg {
		issue, err := db.GetIssue(conn, issueID)
		if err != nil {
			return detailLoadedMsg{err: err}
		}

		// Hydrate the issue.
		if err := db.HydrateLabels(conn, []*model.Issue{issue}); err != nil {
			return detailLoadedMsg{err: err}
		}
		if err := db.HydrateFiles(conn, []*model.Issue{issue}); err != nil {
			return detailLoadedMsg{err: err}
		}

		subs, err := db.GetSubIssues(conn, issueID)
		if err != nil {
			return detailLoadedMsg{err: err}
		}

		relations, err := db.GetIssueRelations(conn, issueID)
		if err != nil {
			return detailLoadedMsg{err: err}
		}

		comments, err := db.ListComments(conn, issueID)
		if err != nil {
			return detailLoadedMsg{err: err}
		}

		activity, err := db.GetActivity(conn, issueID, 20)
		if err != nil {
			return detailLoadedMsg{err: err}
		}

		return detailLoadedMsg{
			issue:     issue,
			subs:      subs,
			relations: relations,
			comments:  comments,
			activity:  activity,
		}
	}
}

// tickCmd returns a tea.Cmd that sends a tickMsg after the given duration.
func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}
