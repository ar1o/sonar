package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ar1o/sonar/internal/render"
)

// renderDetailView renders the full-screen detail view for the selected issue.
func renderDetailView(m Model) string {
	if m.detailIssue == nil {
		return dimStyle.Render("Loading issue details...")
	}

	// Use the existing render.RenderDetail to produce the detail content.
	content := render.RenderDetail(
		m.detailIssue,
		m.detailSubs,
		m.detailRels,
		m.detailComments,
		m.detailActivity,
	)

	// Apply scroll offset.
	lines := strings.Split(content, "\n")

	// Clamp scroll.
	maxScroll := len(lines) - (m.height - 3) // reserve lines for status + help
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.detailScroll > maxScroll {
		m.detailScroll = maxScroll
	}

	// Slice visible region.
	start := m.detailScroll
	if start > len(lines) {
		start = len(lines)
	}
	visibleLines := lines[start:]

	maxVisible := m.height - 3
	if maxVisible < 1 {
		maxVisible = 1
	}
	if len(visibleLines) > maxVisible {
		visibleLines = visibleLines[:maxVisible]
	}

	// Add scroll indicator if content overflows.
	result := strings.Join(visibleLines, "\n")
	if len(lines) > maxVisible {
		indicator := dimStyle.Render(scrollIndicator(m.detailScroll, maxScroll))
		// Place indicator at the right edge of the top line.
		result = lipgloss.JoinHorizontal(lipgloss.Top,
			result,
			"\n"+indicator,
		)
	}

	return result
}

// scrollIndicator returns a scroll position hint like "[3/15]".
func scrollIndicator(current, max int) string {
	if max <= 0 {
		return ""
	}
	pct := (current * 100) / max
	if pct > 100 {
		pct = 100
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render(strings.Repeat(" ", 2) + "[" + strings.Repeat("=", pct/10) + strings.Repeat(" ", 10-pct/10) + "]")
}
