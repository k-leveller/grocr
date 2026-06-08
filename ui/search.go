package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/k-leveller/grocr/api"
)

type Search struct {
	Input      textinput.Model
	Products   []api.Product
	Locations  []api.Location
	LocFilter  int // 0 = all; otherwise a location ID
	Filtered   []api.Product
	Cursor     int
	Selected   *api.Product
	Cancelled  bool
	MaxResults int
}

func NewSearch(products []api.Product, locations []api.Location) Search {
	ti := textinput.New()
	ti.Placeholder = "type to search..."
	ti.Focus()
	ti.CharLimit = 80

	return Search{
		Input:      ti,
		Products:   products,
		Locations:  locations,
		MaxResults: 10,
	}
}

func (s *Search) cycleLocation() {
	if len(s.Locations) == 0 {
		return
	}
	if s.LocFilter == 0 {
		s.LocFilter = s.Locations[0].ID
		return
	}
	for i, loc := range s.Locations {
		if loc.ID == s.LocFilter {
			if i+1 < len(s.Locations) {
				s.LocFilter = s.Locations[i+1].ID
			} else {
				s.LocFilter = 0
			}
			return
		}
	}
	s.LocFilter = 0
}

func (s *Search) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.Cancelled = true
			return nil
		case "enter":
			if len(s.Filtered) > 0 && s.Cursor < len(s.Filtered) {
				s.Selected = &s.Filtered[s.Cursor]
			}
			return nil
		case "down":
			if s.Cursor < len(s.Filtered)-1 {
				s.Cursor++
			}
			return nil
		case "up":
			if s.Cursor > 0 {
				s.Cursor--
			}
			return nil
		case "tab":
			s.cycleLocation()
			s.filter()
			return nil
		}
	}

	var cmd tea.Cmd
	s.Input, cmd = s.Input.Update(msg)

	// Re-filter on every keystroke
	s.filter()

	return cmd
}

func (s *Search) filter() {
	query := strings.ToLower(s.Input.Value())
	if query == "" && s.LocFilter == 0 {
		s.Filtered = nil
		s.Cursor = 0
		return
	}

	var results []api.Product
	for _, p := range s.Products {
		if s.LocFilter != 0 && p.LocationID != 0 && p.LocationID != s.LocFilter {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(p.Name), query) {
			continue
		}
		results = append(results, p)
		if len(results) >= s.MaxResults {
			break
		}
	}
	s.Filtered = results
	if s.Cursor >= len(s.Filtered) {
		s.Cursor = 0
	}
}

func (s *Search) locFilterName() string {
	if s.LocFilter == 0 {
		return "all locations"
	}
	for _, loc := range s.Locations {
		if loc.ID == s.LocFilter {
			return loc.Name
		}
	}
	return "unknown"
}

func (s *Search) View() string {
	var lines []string

	header := " " + StyleBold.Render("Search:") + " " + s.Input.View()
	if len(s.Locations) > 0 {
		header += "  " + StyleHint.Render("[Tab:") + " " + StyleInfo.Render(s.locFilterName()) + StyleHint.Render("]")
	}
	lines = append(lines, header)
	lines = append(lines, "")

	noResults := len(s.Filtered) == 0
	hasQuery := s.Input.Value() != "" || s.LocFilter != 0
	if noResults {
		if hasQuery {
			lines = append(lines, " "+StyleHint.Render("No matches"))
		}
	} else {
		for i, p := range s.Filtered {
			prefix := "  "
			if i == s.Cursor {
				prefix = StyleInfo.Render("> ")
			}
			lines = append(lines, fmt.Sprintf(" %s%s  %s", prefix, p.Name, StyleHint.Render(fmt.Sprintf("[id:%d]", p.ID))))
		}
	}

	lines = append(lines, "")
	lines = append(lines, " "+StyleHint.Render("↑/↓ navigate · Tab filter location · Enter select · n new · Esc cancel"))

	return strings.Join(lines, "\n")
}
