package components

import (
	"github.com/OrionVesper/tfplan-pretty-output/tui/styles"
)

func RenderHints(hints string, width int) string {
	_ = width
	return "  " + styles.Hints.Render(hints) + "\n"
}
