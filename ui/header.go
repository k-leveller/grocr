package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/k-leveller/grocr/locale"
)

func RenderHeader(mode string, width int) string {
	title := StyleHeader.Render("GROCR")

	var modeBadge string
	switch mode {
	case "add":
		modeBadge = StyleModeAdd.Render(locale.Active.ModeAdd)
	case "consume":
		modeBadge = StyleModeConsume.Render(locale.Active.ModeConsume)
	default:
		modeBadge = StyleModeLookup.Render(locale.Active.ModeLookup)
	}

	hints := StyleHint.Render(HeaderHints())

	left := fmt.Sprintf(" %s %s", title, modeBadge)
	right := fmt.Sprintf("%s ", hints)

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + right
}
