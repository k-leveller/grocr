package ui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/k-leveller/grocr/api"
	"github.com/k-leveller/grocr/locale"
)

type Search struct {
	Input         textinput.Model
	Products      []api.Product
	Locations     []api.Location
	QuantityUnits []api.QuantityUnit
	LocFilter     int // 0 = all; otherwise a location ID
	Filtered      []api.Product
	Cursor        int
	Selected      *api.Product
	Cancelled     bool
	MaxResults    int
	StockAmounts  map[int]float64
	DimZeroStock  bool // grey out zero-stock product names (lookup and consume modes)
}

func NewSearch(products []api.Product, locations []api.Location, quantityUnits []api.QuantityUnit, stockAmounts map[int]float64, dimZeroStock bool) Search {
	ti := textinput.New()
	ti.Placeholder = locale.Active.SearchPlaceholder
	ti.Focus()
	ti.CharLimit = 80

	return Search{
		Input:         ti,
		Products:      products,
		Locations:     locations,
		QuantityUnits: quantityUnits,
		MaxResults:    10,
		StockAmounts:  stockAmounts,
		DimZeroStock:  dimZeroStock,
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

func (s *Search) UpdateFilter() { s.filter() }

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
	}

	// Zero-stock items always sort below in-stock items.
	if s.StockAmounts != nil {
		sort.SliceStable(results, func(i, j int) bool {
			return s.StockAmounts[results[i].ID] > 0 && s.StockAmounts[results[j].ID] <= 0
		})
	}

	if len(results) > s.MaxResults {
		results = results[:s.MaxResults]
	}

	s.Filtered = results
	if s.Cursor >= len(s.Filtered) {
		s.Cursor = 0
	}
}

func (s *Search) locFilterName() string {
	if s.LocFilter == 0 {
		return locale.Active.AllLocations
	}
	for _, loc := range s.Locations {
		if loc.ID == s.LocFilter {
			return loc.Name
		}
	}
	return "unknown"
}

func (s *Search) locationName(id int) string {
	for _, loc := range s.Locations {
		if loc.ID == id {
			return loc.Name
		}
	}
	return ""
}

func (s *Search) unitName(id int) string {
	for _, u := range s.QuantityUnits {
		if u.ID == id {
			return u.Name
		}
	}
	return ""
}

func formatQty(qty float64) string {
	if qty == float64(int64(qty)) {
		return strconv.FormatInt(int64(qty), 10)
	}
	return strconv.FormatFloat(qty, 'f', 2, 64)
}

func (s *Search) View() string {
	var lines []string

	header := " " + StyleBold.Render(locale.Active.SearchHeader) + " " + s.Input.View()
	if len(s.Locations) > 0 {
		header += "  " + StyleHint.Render("[Tab:") + " " + StyleInfo.Render(s.locFilterName()) + StyleHint.Render("]")
	}
	lines = append(lines, header)
	lines = append(lines, "")

	noResults := len(s.Filtered) == 0
	hasQuery := s.Input.Value() != "" || s.LocFilter != 0
	if noResults {
		if hasQuery {
			lines = append(lines, " "+StyleHint.Render(locale.Active.NoMatches))
		}
	} else {
		for i, p := range s.Filtered {
			prefix := "  "
			if i == s.Cursor {
				prefix = StyleInfo.Render("> ")
			}
			name := p.Name
			zeroStock := s.StockAmounts != nil && s.StockAmounts[p.ID] <= 0
			if s.DimZeroStock && zeroStock {
				name = StyleHint.Render(name)
			}

			var meta []string
			if s.StockAmounts != nil {
				qty := s.StockAmounts[p.ID]
				unit := s.unitName(p.QuIDStock)
				qtyStr := formatQty(qty)
				if unit != "" {
					qtyStr += " " + unit
				}
				if qty <= 0 {
					meta = append(meta, StyleHint.Render(qtyStr))
				} else {
					meta = append(meta, qtyStr)
				}
			}
			if loc := s.locationName(p.LocationID); loc != "" {
				meta = append(meta, StyleHint.Render(loc))
			}

			line := fmt.Sprintf(" %s%s", prefix, name)
			if len(meta) > 0 {
				line += "  " + strings.Join(meta, StyleHint.Render(" · "))
			}
			lines = append(lines, line)
		}
	}

	lines = append(lines, "")
	lines = append(lines, " "+StyleHint.Render(SearchNavHint()))

	return strings.Join(lines, "\n")
}
