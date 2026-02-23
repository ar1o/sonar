package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/ar1o/sonar/internal/render"
)

var (
	// Card styles
	selectedCardBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("15"))

	dimmedCardBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8"))

	// Column header styles
	columnHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Align(lipgloss.Center)

	// Help bar
	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	helpKeyStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Bold(true)

	// Status bar
	statusBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("11"))

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("9"))

	// Title bar
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15"))

	dimStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))
)

// statusColor returns the lipgloss color for a status using the render package.
func statusColor(colorName string) lipgloss.Color {
	return render.ColorFromName(colorName)
}
