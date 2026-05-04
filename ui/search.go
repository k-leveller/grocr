package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevin/grocy-scanner/api"
)

type Search struct {
	Input      textinput.Model
	Products   []api.Product
	Filtered   []api.Product
	Cursor     int
	Selected   *api.Product
	Cancelled  bool
	MaxResults int
}

func NewSearch(products []api.Product) Search {
	ti := textinput.New()
	ti.Placeholder = "type to search..."
	ti.Focus()
	ti.CharLimit = 80

	return Search{
		Input:      ti,
		Products:   products,
		MaxResults: 10,
	}
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
	if query == "" {
		s.Filtered = nil
		s.Cursor = 0
		return
	}

	var results []api.Product
	for _, p := range s.Products {
		if strings.Contains(strings.ToLower(p.Name), query) {
			results = append(results, p)
			if len(results) >= s.MaxResults {
				break
			}
		}
	}
	s.Filtered = results
	if s.Cursor >= len(s.Filtered) {
		s.Cursor = 0
	}
}

func (s *Search) View() string {
	var lines []string
	lines = append(lines, " "+StyleBold.Render("Search:")+" "+s.Input.View())
	lines = append(lines, "")

	if len(s.Filtered) == 0 {
		if s.Input.Value() != "" {
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
	lines = append(lines, " "+StyleHint.Render("↑/↓ to navigate, Enter to select, Esc to cancel"))

	return strings.Join(lines, "\n")
}
