package cli

import (
	"errors"
	"fmt"

	"github.com/ar1o/sonar/internal/db"
	"github.com/ar1o/sonar/internal/output"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

type cleanResult struct {
	Cleared bool `json:"cleared"`
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Delete all issues, comments, labels, and relations",
	RunE: func(cmd *cobra.Command, args []string) error {
		w := getWriter(cmd)
		conn := getDB(cmd)

		force, _ := cmd.Flags().GetBool("force")

		// Count issues so we can report what was cleared.
		total, err := db.CountIssues(conn)
		if err != nil {
			return cmdErr(fmt.Errorf("counting issues: %w", err), output.ErrGeneral)
		}

		if total == 0 {
			w.Success(cleanResult{Cleared: false}, "Nothing to clean — database is already empty.")
			return nil
		}

		if !force {
			if w.JSONMode {
				return cmdErr(fmt.Errorf("this will delete all %d issue(s) and related data; pass --force to confirm", total), output.ErrValidation)
			}

			var confirmed bool
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Delete all %d issue(s) and related data?", total)).
						Affirmative("Yes, delete everything").
						Negative("Cancel").
						Value(&confirmed),
				),
			)

			if err := form.Run(); err != nil {
				if errors.Is(err, huh.ErrUserAborted) {
					w.Info("Cancelled.")
					return nil
				}
				return cmdErr(fmt.Errorf("interactive form failed: %w", err), output.ErrGeneral)
			}

			if !confirmed {
				w.Info("Cancelled.")
				return nil
			}
		}

		if err := db.ClearAllData(conn); err != nil {
			return cmdErr(fmt.Errorf("clearing database: %w", err), output.ErrGeneral)
		}

		w.Success(cleanResult{Cleared: true}, fmt.Sprintf("Cleared all data (%d issue(s) removed).", total))
		return nil
	},
}

func init() {
	cleanCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	rootCmd.AddCommand(cleanCmd)
}
