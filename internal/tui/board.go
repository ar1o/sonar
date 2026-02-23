package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/ar1o/sonar/internal/model"
	"github.com/ar1o/sonar/internal/render"
)

const (
	maxCardsPerColumn = 10
	minColumnWidth    = 20
	cardPadding       = 2
)

// renderBoardView renders the full board with columns and selection highlighting.
func renderBoardView(m Model) string {
	if len(m.issues) == 0 {
		return render.EmptyState("No issues on the board.", "Create one with: sonar issue create", false)
	}

	if len(m.activeStatuses) == 0 {
		return ""
	}

	tw := m.width
	if tw <= 0 {
		tw = 100
	}

	// Calculate column widths.
	gaps := len(m.activeStatuses) - 1
	colWidth := (tw - gaps) / len(m.activeStatuses)
	if colWidth < minColumnWidth {
		colWidth = minColumnWidth
	}
	cardContentWidth := max(colWidth-cardPadding-2, 5)

	var columns []string
	for i, status := range m.activeStatuses {
		isSelectedCol := i == m.colIdx
		col := renderTUIColumn(m, status, colWidth, cardContentWidth, isSelectedCol)
		columns = append(columns, col)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, columns...)
}

// renderTUIColumn renders a single status column with selection awareness.
func renderTUIColumn(m Model, status model.Status, colWidth, contentWidth int, isSelectedCol bool) string {
	issues := m.columns[status]

	// Column header
	headerColor := statusColor(status.Color())
	header := columnHeaderStyle.
		Width(colWidth).
		Foreground(headerColor).
		Render(fmt.Sprintf("%s %s (%d)", status.Icon(), strings.ToUpper(string(status)), len(issues)))

	// Render cards up to the maximum.
	visible := issues
	overflow := 0
	if len(issues) > maxCardsPerColumn {
		visible = issues[:maxCardsPerColumn]
		overflow = len(issues) - maxCardsPerColumn
	}

	cards := make([]string, 0, len(visible)+2)
	cards = append(cards, header)

	for i, issue := range visible {
		isSelected := isSelectedCol && i == m.cardIdx
		card := renderTUICard(issue, colWidth, contentWidth, isSelected, m.progress)
		cards = append(cards, card)
	}

	if overflow > 0 {
		moreStyle := lipgloss.NewStyle().
			Width(colWidth).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("8"))
		cards = append(cards, moreStyle.Render(fmt.Sprintf("+%d more", overflow)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, cards...)
}

// renderTUICard renders a single card with optional selection highlighting.
func renderTUICard(issue *model.Issue, colWidth, contentWidth int, selected bool, progress map[int]render.SubIssueProgress) string {
	lines := render.CardLines(issue, contentWidth, progress)
	body := strings.Join(lines, "\n")

	var cardStyle lipgloss.Style
	if selected {
		cardStyle = selectedCardBorder.
			Width(colWidth - 2).
			Padding(0, 1)
	} else {
		cardStyle = dimmedCardBorder.
			Width(colWidth - 2).
			Padding(0, 1).
			BorderForeground(lipgloss.Color("8"))
	}

	return cardStyle.Render(body)
}
