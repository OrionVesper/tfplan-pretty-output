package styles

import "github.com/charmbracelet/lipgloss"

var (
	SymbolCreate  = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	SymbolUpdate  = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	SymbolDelete  = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	SymbolReplace = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true)
	SymbolNoOp    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	SymbolWarning = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	SymbolNote    = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
)

var (
	Header           = lipgloss.NewStyle().Bold(true)
	PlanID           = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	Selector         = lipgloss.NewStyle().Foreground(lipgloss.Color("210")).Bold(true)
	Hints            = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	Resource         = lipgloss.NewStyle()
	SelectedResource = lipgloss.NewStyle().Bold(true)
	Serial           = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	SearchInput      = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

var (
	ScrollHint = lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Italic(true)
	AttrLine   = lipgloss.NewStyle()
)

var (
	MatchHighlight = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)

	MatchCurrent = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true).Underline(true)

	HelpHeading = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("245"))
	HelpKey     = lipgloss.NewStyle().Foreground(lipgloss.Color("210")).Bold(true)
	HelpDesc    = lipgloss.NewStyle().Foreground(lipgloss.Color("253"))

	VerboseLineNum = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func ColorSymbol(actionType string) lipgloss.Style {
	switch actionType {
	case "create":
		return SymbolCreate
	case "update":
		return SymbolUpdate
	case "delete":
		return SymbolDelete
	case "replace":
		return SymbolReplace
	default:
		return SymbolNoOp
	}
}

func ColorVirtualSymbol(symbol string) lipgloss.Style {
	switch symbol {
	case "⚠":
		return SymbolWarning
	case "ℹ":
		return SymbolNote
	case "+":
		return SymbolCreate
	case "~":
		return SymbolUpdate
	case "-":
		return SymbolDelete
	}
	return SymbolNoOp
}
