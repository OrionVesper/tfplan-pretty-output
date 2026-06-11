package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/OrionVesper/tfplan-pretty-output/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

type VerboseResource struct {
	Address     string
	ActionType  string
	Symbol      string
	ActionLabel string
	Body        string
	SerialNum   int
}

type matchPos struct {
	Line int
	Col  int
}

type VerboseModel struct {
	Width  int
	Height int

	Resources []VerboseResource

	ScrollOffset int

	SearchMode  bool
	SearchInput string

	SearchQuery     string
	MatchPositions  []matchPos
	CurrentMatchIdx int

	renderedLines []string
	styledLines   []string
}

func NewVerboseModel(resources []VerboseResource, initialQuery string) VerboseModel {
	m := VerboseModel{Resources: resources}
	m.rebuildContent()
	if initialQuery != "" {
		m.SearchQuery = initialQuery
		m.applySearch()
	}
	return m
}

func (m *VerboseModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.rebuildContent()
	if m.SearchQuery != "" {
		m.applySearch()
	}
}

func (m *VerboseModel) rebuildContent() {
	width := m.Width
	if width <= 0 {
		width = 100
	}
	maxSerial := 0
	for _, r := range m.Resources {
		if r.SerialNum > maxSerial {
			maxSerial = r.SerialNum
		}
	}
	digits := len(strconv.Itoa(maxSerial))
	if digits < 1 {
		digits = 1
	}

	var plain []string
	var styled []string

	for i, r := range m.Resources {
		if i > 0 {
			plain = append(plain, "")
			styled = append(styled, "")
		}

		serial := fmt.Sprintf("%*d", digits, r.SerialNum)
		label := "# " + r.Address
		if r.ActionLabel != "" {
			label += " " + r.ActionLabel
		}
		header := "  " + serial + "  " + label
		visibleSym := r.Symbol
		gap := width - len(header) - len(visibleSym) - 2
		if gap < 1 {
			gap = 1
		}
		plainHeader := header + strings.Repeat(" ", gap) + visibleSym
		styledHeader := "  " +
			styles.VerboseLineNum.Render(serial) + "  " +
			styles.SelectedResource.Render(label) +
			strings.Repeat(" ", gap) +
			styles.ColorSymbol(r.ActionType).Render(visibleSym)

		plain = append(plain, plainHeader)
		styled = append(styled, styledHeader)

		for _, line := range strings.Split(strings.TrimRight(r.Body, "\n"), "\n") {
			indented := "      " + line
			plain = append(plain, indented)
			styled = append(styled, styles.AttrLine.Render(indented))
		}
	}

	m.renderedLines = plain
	m.styledLines = styled
}

func (m *VerboseModel) applySearch() {
	q := strings.ToLower(strings.TrimSpace(m.SearchQuery))
	if q == "" {
		m.MatchPositions = nil
		m.CurrentMatchIdx = 0
		// Restore plain styling.
		for i, plain := range m.renderedLines {
			m.styledLines[i] = styles.AttrLine.Render(plain)
		}
		return
	}

	var positions []matchPos
	for li, line := range m.renderedLines {
		lower := strings.ToLower(line)
		start := 0
		for {
			idx := strings.Index(lower[start:], q)
			if idx < 0 {
				break
			}
			positions = append(positions, matchPos{Line: li, Col: start + idx})
			start = start + idx + len(q)
		}
	}
	m.MatchPositions = positions
	m.CurrentMatchIdx = 0

	qLen := len(q)
	linesWithMatches := map[int]bool{}
	for _, p := range positions {
		linesWithMatches[p.Line] = true
	}
	for i, plain := range m.renderedLines {
		if !linesWithMatches[i] {
			m.styledLines[i] = styles.AttrLine.Render(plain)
			continue
		}
		lower := strings.ToLower(plain)
		var b strings.Builder
		cursor := 0
		for cursor < len(plain) {
			idx := strings.Index(lower[cursor:], q)
			if idx < 0 {
				b.WriteString(plain[cursor:])
				break
			}
			absIdx := cursor + idx
			b.WriteString(plain[cursor:absIdx])
			b.WriteString(styles.MatchHighlight.Render(plain[absIdx : absIdx+qLen]))
			cursor = absIdx + qLen
		}
		m.styledLines[i] = b.String()
	}

	if len(positions) > 0 {
		m.scrollToLine(positions[0].Line)
	}
}

func (m *VerboseModel) scrollToLine(target int) {
	if target < 0 {
		target = 0
	}
	visible := m.viewportHeight()
	margin := visible / 3
	desired := target - margin
	if desired < 0 {
		desired = 0
	}
	maxOffset := len(m.renderedLines) - visible
	if maxOffset < 0 {
		maxOffset = 0
	}
	if desired > maxOffset {
		desired = maxOffset
	}
	m.ScrollOffset = desired
}

func (m *VerboseModel) viewportHeight() int {
	h := m.Height - 6
	if h < 5 {
		h = 5
	}
	return h
}

func (m VerboseModel) Update(msg tea.Msg) (VerboseModel, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()

		if m.SearchMode {
			switch key {
			case "esc":
				m.SearchMode = false
				m.SearchInput = ""
				return m, nil, false
			case "enter":
				m.SearchMode = false
				m.SearchQuery = strings.TrimSpace(m.SearchInput)
				m.SearchInput = ""
				m.applySearch()
				return m, nil, false
			case "backspace":
				if len(m.SearchInput) > 0 {
					m.SearchInput = m.SearchInput[:len(m.SearchInput)-1]
				}
				return m, nil, false
			default:
				if len(key) == 1 && key[0] >= 0x20 && key[0] < 0x7f {
					m.SearchInput += key
				}
				return m, nil, false
			}
		}

		switch key {
		case "up":
			if m.ScrollOffset > 0 {
				m.ScrollOffset--
			}
		case "down":
			max := len(m.renderedLines) - m.viewportHeight()
			if max < 0 {
				max = 0
			}
			if m.ScrollOffset < max {
				m.ScrollOffset++
			}
		case "g":
			m.ScrollOffset = 0
		case "G":
			m.ScrollOffset = len(m.renderedLines) - m.viewportHeight()
			if m.ScrollOffset < 0 {
				m.ScrollOffset = 0
			}
		case "/":
			m.SearchMode = true
			m.SearchInput = ""
		case "n":
			if len(m.MatchPositions) > 0 {
				m.CurrentMatchIdx = (m.CurrentMatchIdx + 1) % len(m.MatchPositions)
				m.scrollToLine(m.MatchPositions[m.CurrentMatchIdx].Line)
			}
		case "N":
			if len(m.MatchPositions) > 0 {
				m.CurrentMatchIdx = (m.CurrentMatchIdx - 1 + len(m.MatchPositions)) % len(m.MatchPositions)
				m.scrollToLine(m.MatchPositions[m.CurrentMatchIdx].Line)
			}
		case "esc":
			if m.SearchQuery != "" || len(m.MatchPositions) > 0 {
				m.SearchQuery = ""
				m.MatchPositions = nil
				m.CurrentMatchIdx = 0
				m.applySearch()
				return m, nil, false
			}
			return m, nil, true
		}
	}
	return m, nil, false
}

func (m VerboseModel) SearchStatusLine() string {
	if m.SearchMode {
		return "  " + styles.SearchInput.Render("Search: /"+m.SearchInput+"_")
	}
	if m.SearchQuery == "" {
		return ""
	}
	if len(m.MatchPositions) == 0 {
		return "  " + styles.Hints.Render(fmt.Sprintf("search: %s  (no matches)", m.SearchQuery))
	}
	return "  " + styles.Hints.Render(fmt.Sprintf(
		"search: %s  (match %d of %d)",
		m.SearchQuery, m.CurrentMatchIdx+1, len(m.MatchPositions),
	))
}

func (m VerboseModel) View() string {
	if len(m.renderedLines) == 0 {
		return "  (no content to display)\n"
	}
	height := m.viewportHeight()
	start := m.ScrollOffset
	if start < 0 {
		start = 0
	}
	end := start + height
	if end > len(m.styledLines) {
		end = len(m.styledLines)
	}

	var b strings.Builder
	for i := start; i < end; i++ {
		b.WriteString(m.styledLines[i])
		b.WriteString("\n")
	}

	if status := m.SearchStatusLine(); status != "" {
		b.WriteString("\n")
		b.WriteString(status)
		b.WriteString("\n")
	}
	return b.String()
}

func (m VerboseModel) HasMatches() bool { return len(m.MatchPositions) > 0 }

func (m VerboseModel) IsSearching() bool { return m.SearchMode }

func (m VerboseModel) HasActiveSearch() bool { return m.SearchQuery != "" }
