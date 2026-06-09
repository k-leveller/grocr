package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/k-leveller/grocr/locale"
)

type LogEntry struct {
	ProductName  string
	Quantity     float64
	Location     string
	Expiry       string
	Action       string // "add", "consume", "spoiled", or "transfer"
	Success      bool
	Time         time.Time
	FromLocation string // set for "transfer" action
	ToLocation   string // set for "transfer" action
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
			detail = fmt.Sprintf(locale.Active.FmtLogConsumed, icon, e.ProductName, e.Quantity)
		} else if e.Action == "spoiled" {
			detail = fmt.Sprintf(locale.Active.FmtLogSpoiled, icon, e.ProductName, e.Quantity)
		} else if e.Action == "transfer" {
			detail = fmt.Sprintf(locale.Active.FmtLogTransfer, icon, e.ProductName, e.Quantity, e.Location)
		} else {
			parts := []string{fmt.Sprintf(locale.Active.FmtLogAdd, icon, e.ProductName, e.Quantity)}
			if e.Location != "" {
				parts = append(parts, fmt.Sprintf(locale.Active.LogAddTo, e.Location))
			}
			if e.Expiry != "" {
				parts = append(parts, fmt.Sprintf(locale.Active.LogAddExpiry, e.Expiry))
			}
			detail = strings.Join(parts, " ")
		}

		line := fmt.Sprintf(" %s  %s", detail, StyleHint.Render(timeStr))
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
