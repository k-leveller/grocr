package ui

import (
	"fmt"
	"strings"
	"time"
)

type LogEntry struct {
	ProductName string
	Quantity    int
	Location    string
	Expiry      string
	Action      string // "add" or "consume"
	Success     bool
	Time        time.Time
}

func RenderLog(entries []LogEntry, maxLines int) string {
	if len(entries) == 0 {
		return ""
	}

	var lines []string
	count := maxLines
	if len(entries) < count {
		count = len(entries)
	}

	for i := 0; i < count; i++ {
		e := entries[i]
		timeStr := e.Time.Format("15:04")

		var icon string
		if e.Success {
			icon = StyleSuccess.Render("✓")
		} else {
			icon = StyleError.Render("✗")
		}

		var detail string
		if e.Action == "consume" {
			detail = fmt.Sprintf("%s %s x%d consumed", icon, e.ProductName, e.Quantity)
		} else {
			parts := []string{fmt.Sprintf("%s %s x%d", icon, e.ProductName, e.Quantity)}
			if e.Location != "" {
				parts = append(parts, "→ "+e.Location)
			}
			if e.Expiry != "" {
				parts = append(parts, "("+e.Expiry+")")
			}
			detail = strings.Join(parts, " ")
		}

		line := fmt.Sprintf(" %s  %s", detail, StyleHint.Render(timeStr))
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
