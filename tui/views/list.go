package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/OrionVesper/tfplan-pretty-output/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ResourceItem struct {
	Address     string
	ActionType  string
	Symbol      string
	ActionLabel string
}

type VirtualRowsConfig struct {
	HasOutputs     bool
	OutputCount    int
	OutputSymbol   string
	OutputSection  string
	HasWarnings    bool
	WarningCount   int
	WarningSection string
	HasNotes       bool
	NoteCount      int
	NoteSection    string
}

type ListModel struct {
	Resources []ResourceItem
	Cursor    int
	Width     int
	Height    int

	VR VirtualRowsConfig

	ExpandedIdx        int
	DetailLines        []string
	DetailScrollOffset int
	DetailMaxVisible   int

	SearchMode        bool
	SearchInput       string
	PendingAutoExpand bool

	WantVerbose          bool
	PendingVerboseSearch string
	ViewportStart int
}

func NewListModel(resources []ResourceItem, vr VirtualRowsConfig) ListModel {
	return ListModel{
		Resources:        resources,
		VR:               vr,
		ExpandedIdx:      -1,
		DetailMaxVisible: 10,
	}
}

func (m *ListModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	v := h - 8
	if v > 20 {
		v = 20
	}
	if v < 6 {
		v = 6
	}
	m.DetailMaxVisible = v
	m.clampDetailScroll()
	m.adjustViewport()
}

func (m ListModel) IsExpanded() bool { return m.ExpandedIdx >= 0 }

func (m *ListModel) Expand(idx int, detailText string) {
	m.ExpandedIdx = idx
	m.DetailLines = splitDetailLines(detailText)
	m.DetailScrollOffset = 0
}

func (m *ListModel) Collapse() {
	m.ExpandedIdx = -1
	m.DetailLines = nil
	m.DetailScrollOffset = 0
}

func (m *ListModel) ScrollDetailUp() {
	if m.DetailScrollOffset > 0 {
		m.DetailScrollOffset--
	}
}

func (m *ListModel) ScrollDetailDown() {
	if m.DetailScrollOffset+m.DetailMaxVisible < len(m.DetailLines) {
		m.DetailScrollOffset++
	}
}

func (m *ListModel) clampDetailScroll() {
	if !m.IsExpanded() {
		return
	}
	max := len(m.DetailLines) - m.DetailMaxVisible
	if max < 0 {
		max = 0
	}
	if m.DetailScrollOffset > max {
		m.DetailScrollOffset = max
	}
}

func (m *ListModel) adjustViewport() {
	if m.Height <= 0 {
		return
	}
	visibleHeight := m.Height - 6
	if visibleHeight < 5 {
		visibleHeight = 5
	}
	total := m.totalRowCount()
	if total == 0 {
		m.ViewportStart = 0
		return
	}
	rowHeights := computeRowHeights(*m, total)

	if m.Cursor < m.ViewportStart {
		m.ViewportStart = m.Cursor
		return
	}
	end := m.ViewportStart
	used := 0
	for end < total {
		used += rowHeights[end]
		if used > visibleHeight {
			break
		}
		end++
	}
	if m.Cursor >= end {
		h := rowHeights[m.Cursor]
		start := m.Cursor
		for start > 0 {
			prev := start - 1
			if h+rowHeights[prev] > visibleHeight {
				break
			}
			h += rowHeights[prev]
			start = prev
		}
		m.ViewportStart = start
	}
	if m.ViewportStart < 0 {
		m.ViewportStart = 0
	}
	if m.ViewportStart > total-1 {
		m.ViewportStart = total - 1
	}
}

func (m ListModel) virtualKindAtOrdered(v int) string {
	order := m.virtualOrder()
	if v < 0 || v >= len(order) {
		return ""
	}
	return order[v]
}

func (m ListModel) virtualOrder() []string {
	var order []string
	if m.VR.HasOutputs {
		order = append(order, "outputs")
	}
	if m.VR.HasWarnings {
		order = append(order, "warnings")
	}
	if m.VR.HasNotes {
		order = append(order, "notes")
	}
	return order
}

func (m ListModel) totalRowCount() int {
	return len(m.Resources) + len(m.virtualOrder())
}

func (m ListModel) VirtualRowKindAt(idx int) string {
	if idx < len(m.Resources) {
		return ""
	}
	v := idx - len(m.Resources)
	return m.virtualKindAtOrdered(v)
}

func (m ListModel) VirtualSectionFor(kind string) string {
	switch kind {
	case "outputs":
		return m.VR.OutputSection
	case "warnings":
		return m.VR.WarningSection
	case "notes":
		return m.VR.NoteSection
	}
	return ""
}

func (m ListModel) virtualLabel(kind string) string {
	switch kind {
	case "outputs":
		s := "change"
		if m.VR.OutputCount != 1 {
			s = "changes"
		}
		return fmt.Sprintf("# Changes to Outputs (%d %s)", m.VR.OutputCount, s)
	case "warnings":
		s := "warning"
		if m.VR.WarningCount != 1 {
			s = "warnings"
		}
		return fmt.Sprintf("# Warnings (%d %s)", m.VR.WarningCount, s)
	case "notes":
		s := "note"
		if m.VR.NoteCount != 1 {
			s = "notes"
		}
		return fmt.Sprintf("# Notes (%d %s)", m.VR.NoteCount, s)
	}
	return ""
}

func (m ListModel) virtualSymbol(kind string) string {
	switch kind {
	case "outputs":
		if m.VR.OutputSymbol == "" {
			return "+"
		}
		return m.VR.OutputSymbol
	case "warnings":
		return "⚠"
	case "notes":
		return "ℹ"
	}
	return "="
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func (m ListModel) Update(msg tea.Msg) (ListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		if m.SearchMode {
			switch key {
			case "esc":
				m.SearchMode = false
				m.SearchInput = ""
				return m, nil
			case "enter":
				if m.SearchInput == "" {
					m.SearchMode = false
					return m, nil
				}
				if isAllDigits(m.SearchInput) {
					if n, err := strconv.Atoi(m.SearchInput); err == nil {
						target := n - 1
						if target >= 0 && target < m.totalRowCount() {
							if m.IsExpanded() {
								m.Collapse()
							}
							m.Cursor = target
							m.PendingAutoExpand = true
							m.adjustViewport()
						}
					}
				} else {

					m.PendingVerboseSearch = m.SearchInput
				}
				m.SearchMode = false
				m.SearchInput = ""
				return m, nil
			case "backspace":
				if len(m.SearchInput) > 0 {
					m.SearchInput = m.SearchInput[:len(m.SearchInput)-1]
				}
				return m, nil
			default:
				if len(key) == 1 && key[0] >= 0x20 && key[0] < 0x7f {
					m.SearchInput += key
				}
				return m, nil
			}
		}

		switch key {
		case "g":
			if m.IsExpanded() {
				m.Collapse()
			}
			m.Cursor = 0
			m.adjustViewport()
		case "G":
			if m.IsExpanded() {
				m.Collapse()
			}
			if total := m.totalRowCount(); total > 0 {
				m.Cursor = total - 1
			}
			m.adjustViewport()
		case "/":
			if m.IsExpanded() {
				m.Collapse()
			}
			m.SearchMode = true
			m.SearchInput = ""
		case "v":
			if m.IsExpanded() {
				m.Collapse()
			}
			m.WantVerbose = true
		case "up":
			if m.Cursor > 0 {
				m.Cursor--
			}
			if m.IsExpanded() && m.Cursor != m.ExpandedIdx {
				m.Collapse()
			}
			m.adjustViewport()
		case "down":
			if m.Cursor < m.totalRowCount()-1 {
				m.Cursor++
			}
			if m.IsExpanded() && m.Cursor != m.ExpandedIdx {
				m.Collapse()
			}
			m.adjustViewport()
		case "shift+up":
			m.ScrollDetailUp()
		case "shift+down":
			m.ScrollDetailDown()
		}
	}
	return m, nil
}

func (m ListModel) isFirstVirtualRow(i int) bool {
	return i == len(m.Resources) && len(m.virtualOrder()) > 0
}

func computeRowHeights(m ListModel, total int) []int {
	width := m.Width
	if width <= 0 {
		width = 100
	}
	digits := len(strconv.Itoa(total))
	h := make([]int, total)
	for i := 0; i < total; i++ {
		h[i] = m.wrappedRowHeight(i, digits, width)
		if m.isFirstVirtualRow(i) {
			h[i]++
		}
		if i == m.ExpandedIdx {
			windowSize := m.DetailMaxVisible
			if windowSize > len(m.DetailLines) {
				windowSize = len(m.DetailLines)
			}
			h[i] += windowSize
			if m.DetailScrollOffset+m.DetailMaxVisible < len(m.DetailLines) {
				h[i]++
			}
		}
	}
	return h
}

func (m ListModel) wrappedRowHeight(i, digits, width int) int {
	addressLabel := m.rowAddressLabel(i)
	overhead := 2 + digits + 2 + 2 + 1 + 1
	available := width - overhead
	if available < 10 {
		available = 10
	}
	lines := wrapToWidth(addressLabel, available)
	if len(lines) == 0 {
		return 1
	}
	return len(lines)
}

func (m ListModel) rowAddressLabel(i int) string {
	if i >= len(m.Resources) {
		return m.virtualLabel(m.VirtualRowKindAt(i))
	}
	r := m.Resources[i]
	if r.ActionLabel != "" {
		return "# " + r.Address + " " + r.ActionLabel
	}
	return "# " + r.Address
}

func wrapToWidth(s string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{s}
	}
	if lipgloss.Width(s) <= maxWidth {
		return []string{s}
	}
	var lines []string
	remaining := s
	for lipgloss.Width(remaining) > maxWidth {
		breakAt := -1
		for j := maxWidth; j > maxWidth/2; j-- {
			if j >= len(remaining) {
				continue
			}
			c := remaining[j]
			if c == '.' || c == '/' || c == ' ' || c == ',' || c == '[' || c == ']' {
				breakAt = j
				break
			}
		}
		if breakAt < 0 {
			breakAt = maxWidth
		}
		if breakAt > len(remaining) {
			breakAt = len(remaining)
		}
		lines = append(lines, remaining[:breakAt])
		remaining = remaining[breakAt:]
		remaining = strings.TrimLeft(remaining, " ")
	}
	if remaining != "" {
		lines = append(lines, remaining)
	}
	return lines
}

func (m ListModel) View() string {
	if len(m.Resources) == 0 && len(m.virtualOrder()) == 0 {
		return "  (No resource changes — nothing to view)\n"
	}
	width := m.Width
	if width <= 0 {
		width = 100
	}
	visibleHeight := m.Height - 6
	if visibleHeight < 5 {
		visibleHeight = 5
	}
	total := m.totalRowCount()
	digits := len(strconv.Itoa(total))
	rowHeights := computeRowHeights(m, total)

	startIdx := m.ViewportStart
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= total {
		startIdx = total - 1
	}

	var b strings.Builder
	used := 0
	for i := startIdx; i < total; i++ {
		if used+rowHeights[i] > visibleHeight && i != m.Cursor {
			break
		}
		if m.isFirstVirtualRow(i) {
			b.WriteString("\n")
		}
		b.WriteString(m.renderRow(i, digits, width))
		if i == m.ExpandedIdx {
			b.WriteString(m.renderExpansion())
		}
		used += rowHeights[i]
	}

	if m.SearchMode {
		b.WriteString("\n  ")
		b.WriteString(styles.SearchInput.Render("Search: /" + m.SearchInput + "_"))
		b.WriteString("\n")
	}
	return b.String()
}

func (m ListModel) renderRow(i, digits, width int) string {
	serial := fmt.Sprintf("%*d", digits, i+1)
	serialRendered := styles.Serial.Render(serial)
	serialBlank := strings.Repeat(" ", lipgloss.Width(serialRendered))

	isVirtual := i >= len(m.Resources)
	var addressLabel, symbol, actionType string
	if isVirtual {
		kind := m.VirtualRowKindAt(i)
		addressLabel = m.virtualLabel(kind)
		symbol = m.virtualSymbol(kind)
	} else {
		r := m.Resources[i]
		actionType = r.ActionType
		symbol = r.Symbol
		if r.ActionLabel != "" {
			addressLabel = "# " + r.Address + " " + r.ActionLabel
		} else {
			addressLabel = "# " + r.Address
		}
	}

	var coloredSym string
	if isVirtual {
		coloredSym = styles.ColorVirtualSymbol(symbol).Render(symbol)
	} else {
		coloredSym = styles.ColorSymbol(actionType).Render(symbol)
	}

	overhead := 2 + lipgloss.Width(serialRendered) + 2 + 2 + 1 + lipgloss.Width(coloredSym)
	available := width - overhead
	if available < 10 {
		available = 10
	}
	wrapped := wrapToWidth(addressLabel, available)
	if len(wrapped) == 0 {
		wrapped = []string{""}
	}

	selectorPrefix := "  "
	if i == m.Cursor {
		selectorPrefix = styles.Selector.Render("> ")
	}
	continuationPrefix := "  " + serialBlank + "  " + "  "

	var b strings.Builder
	for lineIdx, lineText := range wrapped {
		var styledLine string
		if i == m.Cursor {
			styledLine = styles.SelectedResource.Render(lineText)
		} else {
			styledLine = styles.Resource.Render(lineText)
		}
		if lineIdx == 0 {
			b.WriteString("  ")
			b.WriteString(serialRendered)
			b.WriteString("  ")
			b.WriteString(selectorPrefix)
			b.WriteString(styledLine)
		} else {
			b.WriteString(continuationPrefix)
			b.WriteString(styledLine)
		}
		if lineIdx == len(wrapped)-1 {
			lineFootprint := lipgloss.Width(styledLine)
			var prefixFootprint int
			if lineIdx == 0 {
				prefixFootprint = 2 + lipgloss.Width(serialRendered) + 2 + lipgloss.Width(selectorPrefix)
			} else {
				prefixFootprint = lipgloss.Width(continuationPrefix)
			}
			gap := width - prefixFootprint - lineFootprint - lipgloss.Width(coloredSym)
			if gap < 1 {
				gap = 1
			}
			b.WriteString(strings.Repeat(" ", gap))
			b.WriteString(coloredSym)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m ListModel) renderExpansion() string {
	if !m.IsExpanded() || len(m.DetailLines) == 0 {
		return ""
	}
	var b strings.Builder
	end := m.DetailScrollOffset + m.DetailMaxVisible
	if end > len(m.DetailLines) {
		end = len(m.DetailLines)
	}
	for i := m.DetailScrollOffset; i < end; i++ {
		b.WriteString("      ")
		b.WriteString(styles.AttrLine.Render(m.DetailLines[i]))
		b.WriteString("\n")
	}
	remaining := len(m.DetailLines) - end
	if remaining > 0 {
		b.WriteString("      ")
		hint := fmt.Sprintf("↓ %d more lines (Shift + ↑ ↓ to scroll)", remaining)
		b.WriteString(styles.ScrollHint.Render(hint))
		b.WriteString("\n")
	}
	return b.String()
}

func splitDetailLines(s string) []string {
	return strings.Split(strings.TrimRight(s, "\n"), "\n")
}
