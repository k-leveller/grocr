package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func RenderHeader(mode string, width int) string {
	title := StyleHeader.Render("GROCY SCANNER")

	var modeBadge string
	switch mode {
	case "add":
		modeBadge = StyleModeAdd.Render("[ADD]")
	case "consume":
		modeBadge = StyleModeConsume.Render("[EAT]")
	default:
		modeBadge = StyleModeLookup.Render("[LOOK]")
	}

	hints := StyleHint.Render("q:quit m:mode /:search ^N:new ?:help")

	left := fmt.Sprintf(" %s %s", title, modeBadge)
	right := fmt.Sprintf("%s ", hints)

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + right
}
