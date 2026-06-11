package views

import (
	"strings"

	"github.com/OrionVesper/tfplan-pretty-output/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type HelpModel struct {
	Width  int
	Height int
}

func NewHelpModel() HelpModel {
	return HelpModel{}
}

func (m *HelpModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
}

func (m HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd, bool) {
	if _, ok := msg.(tea.KeyMsg); ok {
		return m, nil, true
	}
	return m, nil, false
}

func helpRow(key, desc string) string {
	pad := 14 - len(key)
	if pad < 1 {
		pad = 1
	}
	return "    " + styles.HelpKey.Render(key) + strings.Repeat(" ", pad) +
		styles.HelpDesc.Render(desc) + "\n"
}

func (m HelpModel) View() string {
	var b strings.Builder

	b.WriteString("\n  ")
	b.WriteString(styles.Header.Render("tfplan-pretty-output — keyboard reference"))
	b.WriteString("\n\n")

	b.WriteString("  ")
	b.WriteString(styles.HelpHeading.Render("Navigation"))
	b.WriteString("\n")
	b.WriteString(helpRow("↑ ↓", "Move between resources"))
	b.WriteString(helpRow("g", "Jump to top"))
	b.WriteString(helpRow("G", "Jump to bottom"))
	b.WriteString(helpRow("Enter", "Expand / collapse current row"))
	b.WriteString(helpRow("Shift + ↑ ↓", "Scroll inside expansion"))
	b.WriteString(helpRow("Esc", "Close expansion / clear search / quit"))
	b.WriteString(helpRow("q / Ctrl+C", "Quit"))
	b.WriteString("\n")

	b.WriteString("  ")
	b.WriteString(styles.HelpHeading.Render("Modes"))
	b.WriteString("\n")
	b.WriteString(helpRow("v", "Switch to verbose mode (all expanded)"))
	b.WriteString(helpRow("Esc", "Return to list view (from verbose / help)"))
	b.WriteString("\n")

	b.WriteString("  ")
	b.WriteString(styles.HelpHeading.Render("Search"))
	b.WriteString("\n")
	b.WriteString(helpRow("/<text>", "Search → opens verbose mode with matches highlighted"))
	b.WriteString(helpRow("/<number>", "Jump to resource row N and auto-expand (list view)"))
	b.WriteString(helpRow("n", "Next match (in verbose mode)"))
	b.WriteString(helpRow("N", "Previous match (in verbose mode)"))
	b.WriteString("\n")

	b.WriteString("  ")
	b.WriteString(styles.HelpHeading.Render("Symbols"))
	b.WriteString("\n")
	b.WriteString(helpRow("+", "Create"))
	b.WriteString(helpRow("~", "Update"))
	b.WriteString(helpRow("-", "Destroy"))
	b.WriteString(helpRow("±", "Replace (destroy and re-create)"))
	b.WriteString(helpRow("=", "No-op"))
	b.WriteString("\n")

	b.WriteString("  ")
	b.WriteString(styles.HelpDesc.Render("Press any key to dismiss."))
	b.WriteString("\n")

	return b.String()
}
