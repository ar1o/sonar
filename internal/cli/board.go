package cli

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/ar1o/sonar/internal/db"
	"github.com/ar1o/sonar/internal/model"
	"github.com/ar1o/sonar/internal/output"
	"github.com/ar1o/sonar/internal/render"
	"github.com/ar1o/sonar/internal/tui"
)

// boardColumn represents a single status column in the board JSON output.
type boardColumn struct {
	Status string         `json:"status"`
	Count  int            `json:"count"`
	Issues []*model.Issue `json:"issues"`
}

// boardResult is the JSON output structure for the board command.
type boardResult struct {
	Columns []boardColumn `json:"columns"`
}

var boardCmd = &cobra.Command{
	Use:   "board",
	Short: "Show Kanban board",
	RunE: func(cmd *cobra.Command, args []string) error {
		w := getWriter(cmd)
		conn := getDB(cmd)

		labels, _ := cmd.Flags().GetStringSlice("label")
		priorities, _ := cmd.Flags().GetStringSlice("priority")
		assignee, _ := cmd.Flags().GetString("assignee")
		expand, _ := cmd.Flags().GetBool("expand")
		interactive, _ := cmd.Flags().GetBool("interactive")
		watch, _ := cmd.Flags().GetBool("watch")
		interval, _ := cmd.Flags().GetDuration("interval")

		// Validate filter enum values.
		for _, p := range priorities {
			if err := model.ValidatePriority(model.Priority(p)); err != nil {
				return cmdErr(err, output.ErrValidation)
			}
		}

		// Launch interactive TUI only when explicitly requested, not JSON, and stdout is a TTY.
		interactive = interactive && !w.JSONMode && term.IsTerminal(int(os.Stdout.Fd()))

		if interactive {
			cfg := tui.BoardConfig{
				Labels:     labels,
				Priorities: priorities,
				Assignee:   assignee,
				Expand:     expand,
			}
			m := tui.NewModel(conn, cfg, watch, interval)
			p := tea.NewProgram(m, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				return cmdErr(fmt.Errorf("TUI error: %w", err), output.ErrGeneral)
			}
			return nil
		}

		// --- Static board (existing behavior) ---

		opts := db.ListOptions{
			Priorities:  priorities,
			Labels:      labels,
			Assignee:    assignee,
			IncludeDone: true,
		}

		issues, _, err := db.ListIssues(conn, opts)
		if err != nil {
			return cmdErr(fmt.Errorf("listing issues: %w", err), output.ErrGeneral)
		}

		// By default, roll up sub-issues into their parent (exclude issues that
		// have a parent). When --expand is set, show all issues individually.
		if !expand {
			var roots []*model.Issue
			for _, issue := range issues {
				if issue.ParentID == nil {
					roots = append(roots, issue)
				}
			}
			issues = roots
		}

		if w.JSONMode {
			// Group issues by status for structured output.
			groups := make(map[model.Status][]*model.Issue)
			for _, issue := range issues {
				groups[issue.Status] = append(groups[issue.Status], issue)
			}

			var columns []boardColumn
			for _, status := range render.StatusOrder {
				col := groups[status]
				if col == nil {
					col = []*model.Issue{}
				}
				columns = append(columns, boardColumn{
					Status: string(status),
					Count:  len(col),
					Issues: col,
				})
			}

			w.Success(boardResult{Columns: columns}, "")
			return nil
		}

		// Build sub-issue progress map for parent issues in a single query.
		parentIDs := make([]int, len(issues))
		for i, issue := range issues {
			parentIDs[i] = issue.ID
		}
		batchProgress, err := db.GetBatchSubIssueProgress(conn, parentIDs)
		if err != nil {
			return cmdErr(fmt.Errorf("fetching sub-issue progress: %w", err), output.ErrGeneral)
		}
		progress := make(map[int]render.SubIssueProgress, len(batchProgress))
		for id, counts := range batchProgress {
			if counts[1] > 0 {
				progress[id] = render.SubIssueProgress{Done: counts[0], Total: counts[1]}
			}
		}

		boardOpts := render.BoardOptions{
			Expand:   expand,
			Progress: progress,
		}
		message := render.RenderBoard(issues, boardOpts)
		w.Success(nil, message)

		return nil
	},
}

func init() {
	boardCmd.Flags().StringSliceP("label", "l", nil, "Filter by label (repeatable)")
	boardCmd.Flags().StringSliceP("priority", "p", nil, "Filter by priority (repeatable)")
	boardCmd.Flags().StringP("assignee", "a", "", "Filter by assignee")
	boardCmd.Flags().Bool("expand", false, "Show sub-issues individually instead of rolling up")
	boardCmd.Flags().BoolP("interactive", "i", false, "Launch interactive TUI board")
	boardCmd.Flags().BoolP("watch", "w", false, "Auto-refresh the board periodically")
	boardCmd.Flags().Duration("interval", 2*time.Second, "Watch mode refresh interval")
	rootCmd.AddCommand(boardCmd)
}
