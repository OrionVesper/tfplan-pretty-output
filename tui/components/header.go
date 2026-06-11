package components

import (
	"strings"

	"github.com/OrionVesper/tfplan-pretty-output/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

func RenderHeader(planID, summary string, width int) string {
	if width <= 0 {
		width = 100
	}
	left := styles.Header.Render("tfplan-pretty-output")
	right := styles.PlanID.Render(planID)
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	line1 := left + strings.Repeat(" ", gap) + right

	if strings.TrimSpace(summary) == "" {
		return line1 + "\n"
	}
	line3 := "  " + styles.Hints.Render(summary)
	return line1 + "\n\n" + line3 + "\n"
}
