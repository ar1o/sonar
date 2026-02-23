package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderHelpBar renders the context-aware help bar at the bottom of the screen.
func renderHelpBar(v viewState, width int) string {
	var bindings []string

	switch v {
	case viewBoard:
		bindings = []string{
			helpBinding("hjkl", "navigate"),
			helpBinding("H/L", "move card"),
			helpBinding("enter", "details"),
			helpBinding("r", "refresh"),
			helpBinding("?", "help"),
			helpBinding("q", "quit"),
		}
	case viewDetail:
		bindings = []string{
			helpBinding("jk", "scroll"),
			helpBinding("H/L", "move status"),
			helpBinding("esc", "back"),
			helpBinding("q", "quit"),
		}
	}

	line := strings.Join(bindings, helpStyle.Render("  |  "))

	return lipgloss.NewStyle().
		Width(width).
		Align(lipgloss.Center).
		Render(line)
}

func helpBinding(key, desc string) string {
	return fmt.Sprintf("%s %s", helpKeyStyle.Render(key), helpStyle.Render(desc))
}
