package tui

import (
	"fmt"
	"strings"

	"github.com/OrionVesper/tfplan-pretty-output/terraform"
	"github.com/OrionVesper/tfplan-pretty-output/tui/components"
	"github.com/OrionVesper/tfplan-pretty-output/tui/views"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenList screen = iota
	screenVerbose
	screenHelp
)

type App struct {
	planID       string
	plan         *terraform.Plan
	planFilePath string
	workdir      string

	width  int
	height int

	screen   screen
	prevScreen screen
	list     views.ListModel
	verbose  views.VerboseModel
	help     views.HelpModel

	cache       *terraform.SectionCache
	planSummary string
}

func IsEmptyPlan(plan *terraform.Plan) bool {
	if plan == nil {
		return true
	}
	for _, rc := range plan.ResourceChanges {
		if rc.Change.Actions.ActionType() != terraform.ActionNoOp {
			return false
		}
	}
	return true
}

func NewApp(planID string, plan *terraform.Plan, planFilePath, workdir string) App {
	cache := terraform.NewSectionCache()
	showText, _ := terraform.ShowPlanText(workdir, planFilePath)
	if showText != "" {
		cache.Set(planFilePath, showText)
	}

	creates, changes, destroys := terraform.ExtractPlanSummary(showText)
	planSummary := fmt.Sprintf("Plan: %d to add · %d to change · %d to destroy",
		creates, changes, destroys)

	outputSection := terraform.ExtractOutputChangesSection(showText)
	outputCount := terraform.CountOutputChanges(outputSection)
	outputSymbol := terraform.DominantOutputSymbol(outputSection)

	vr := views.VirtualRowsConfig{
		HasOutputs:    outputSection != "" && outputCount > 0,
		OutputCount:   outputCount,
		OutputSymbol:  outputSymbol,
		OutputSection: outputSection,
		HasWarnings:   false,
		HasNotes:      false,
	}

	return App{
		planID:       planID,
		plan:         plan,
		planFilePath: planFilePath,
		workdir:      workdir,
		screen:       screenList,
		list:         views.NewListModel(buildResourceItems(plan), vr),
		help:         views.NewHelpModel(),
		cache:        cache,
		planSummary:  planSummary,
	}
}

func buildResourceItems(plan *terraform.Plan) []views.ResourceItem {
	if plan == nil {
		return nil
	}
	out := make([]views.ResourceItem, 0, len(plan.ResourceChanges))
	for i := range plan.ResourceChanges {
		rc := &plan.ResourceChanges[i]
		at := rc.Change.Actions.ActionType()
		out = append(out, views.ResourceItem{
			Address:     rc.Address,
			ActionType:  string(at),
			Symbol:      at.Symbol(),
			ActionLabel: terraform.ActionLabelFor(rc),
		})
	}
	return out
}

func (a App) Init() tea.Cmd { return nil }

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.list.SetSize(msg.Width, msg.Height)
		a.verbose.SetSize(msg.Width, msg.Height)
		a.help.SetSize(msg.Width, msg.Height)
		return a, nil

	case tea.KeyMsg:
		key := msg.String()

		if key == "ctrl+c" || key == "q" {
			return a, tea.Quit
		}

		if key == "?" && a.screen != screenHelp {
			a.prevScreen = a.screen
			a.screen = screenHelp
			return a, nil
		}

		switch a.screen {
		case screenHelp:
			newHelp, cmd, dismiss := a.help.Update(msg)
			a.help = newHelp
			if dismiss {
				a.screen = a.prevScreen
			}
			return a, cmd

		case screenVerbose:
			newV, cmd, dismiss := a.verbose.Update(msg)
			a.verbose = newV
			if dismiss {
				a.screen = screenList
			}
			return a, cmd

		case screenList:
			if key == "esc" {
				if a.list.SearchMode {
					nl, _ := a.list.Update(msg)
					a.list = nl
					return a, nil
				}
				if a.list.IsExpanded() {
					a.list.Collapse()
					return a, nil
				}
				return a, tea.Quit
			}
			if key == "enter" {
				return a.handleEnterInList(msg)
			}
			nl, cmd := a.list.Update(msg)
			a.list = nl

			if a.list.PendingAutoExpand {
				a.list.PendingAutoExpand = false
				return a.doExpandAtCursor(), cmd
			}
			if a.list.WantVerbose {
				a.list.WantVerbose = false
				a = a.enterVerbose("")
				return a, cmd
			}
			if a.list.PendingVerboseSearch != "" {
				q := a.list.PendingVerboseSearch
				a.list.PendingVerboseSearch = ""
				a = a.enterVerbose(q)
				return a, cmd
			}
			return a, cmd
		}
	}
	return a, nil
}

func (a App) handleEnterInList(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.list.SearchMode {
		nl, cmd := a.list.Update(msg)
		a.list = nl
		if a.list.PendingAutoExpand {
			a.list.PendingAutoExpand = false
			return a.doExpandAtCursor(), cmd
		}
		if a.list.PendingVerboseSearch != "" {
			q := a.list.PendingVerboseSearch
			a.list.PendingVerboseSearch = ""
			a = a.enterVerbose(q)
			return a, cmd
		}
		return a, cmd
	}
	cur := a.list.Cursor
	if a.list.IsExpanded() && a.list.ExpandedIdx == cur {
		a.list.Collapse()
		return a, nil
	}
	return a.doExpandAtCursor(), nil
}

func (a App) doExpandAtCursor() App {
	cur := a.list.Cursor
	kind := a.list.VirtualRowKindAt(cur)
	if kind != "" {
		section := a.list.VirtualSectionFor(kind)
		a.list.Expand(cur, section)
		return a
	}
	if cur < 0 || cur >= len(a.plan.ResourceChanges) {
		return a
	}
	address := a.plan.ResourceChanges[cur].Address
	showText, err := a.cache.GetOrLoad(a.workdir, a.planFilePath)
	if err != nil {
		a.list.Expand(cur, "(could not load detail: "+err.Error()+")\n")
		return a
	}
	section := terraform.ExtractResourceSection(showText, address)
	section = stripFirstHeading(section)
	a.list.Expand(cur, section)
	return a
}

func (a App) enterVerbose(query string) App {
	resources := a.buildVerboseResources(query)
	a.verbose = views.NewVerboseModel(resources, query)
	a.verbose.SetSize(a.width, a.height)
	a.screen = screenVerbose
	return a
}

func (a App) buildVerboseResources(filterQuery string) []views.VerboseResource {
	showText, _ := a.cache.GetOrLoad(a.workdir, a.planFilePath)
	q := strings.ToLower(strings.TrimSpace(filterQuery))

	var out []views.VerboseResource
	for i := range a.plan.ResourceChanges {
		rc := &a.plan.ResourceChanges[i]
		at := rc.Change.Actions.ActionType()
		body := stripFirstHeading(terraform.ExtractResourceSection(showText, rc.Address))

		if q != "" {
			haystack := strings.ToLower(rc.Address + "\n" + body)
			if !strings.Contains(haystack, q) {
				continue
			}
		}

		out = append(out, views.VerboseResource{
			Address:     rc.Address,
			ActionType:  string(at),
			Symbol:      at.Symbol(),
			ActionLabel: terraform.ActionLabelFor(rc),
			Body:        body,
			SerialNum:   i + 1,
		})
	}
	return out
}

func (a App) View() string {
	if a.screen == screenHelp {
		return a.help.View()
	}

	var b strings.Builder
	header := components.RenderHeader(a.planID, a.planSummary, a.width)

	var body, hintText string
	switch a.screen {
	case screenVerbose:
		body = a.verbose.View()
		switch {
		case a.verbose.IsSearching():
			hintText = "Enter apply · Esc cancel"
		case a.verbose.HasActiveSearch() && a.verbose.HasMatches():
			hintText = "↑ ↓ scroll · n next · N prev · / new search · Esc clear · q quit"
		case a.verbose.HasActiveSearch():
			hintText = "↑ ↓ scroll · / new search · Esc clear · q quit"
		default:
			hintText = "↑ ↓ scroll · g/G top/bottom · / search · Esc list · ? help · q quit"
		}
	default:
		body = a.list.View()
		switch {
		case a.list.SearchMode:
			hintText = "Enter apply · Esc cancel"
		case a.list.IsExpanded():
			hintText = "↑ ↓ navigate · Enter close · Shift + ↑ ↓ scroll detail · q quit"
		default:
			hintText = "↑ ↓ navigate · Enter expand · / search · v verbose · g/G top/bottom · ? help · q quit"
		}
	}
	hints := components.RenderHints(hintText, a.width)

	headerLines := strings.Count(header, "\n")
	bodyLines := strings.Count(body, "\n")
	hintsLines := strings.Count(hints, "\n")
	used := headerLines + bodyLines + hintsLines + 2
	padding := a.height - used
	if padding < 1 {
		padding = 1
	}

	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(body)
	b.WriteString(strings.Repeat("\n", padding))
	b.WriteString(hints)
	return b.String()
}

func Run(planID string, plan *terraform.Plan, planFilePath, workdir string) error {
	app := NewApp(planID, plan, planFilePath, workdir)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func stripFirstHeading(section string) string {
	if section == "" {
		return section
	}
	nl := strings.IndexByte(section, '\n')
	if nl < 0 {
		return ""
	}
	first := strings.TrimSpace(section[:nl])
	if strings.HasPrefix(first, "# ") && (strings.Contains(first, " will be ") || strings.Contains(first, " must be ")) {
		return section[nl+1:]
	}
	return section
}
