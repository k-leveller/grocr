package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type FormField struct {
	Label    string
	Default  string
	Hint     string
	Required bool
	Input    textinput.Model
}

type Form struct {
	Fields      []FormField
	FocusIndex  int
	Submitted   bool
	Cancelled   bool
}

func NewForm(fields []FormField) Form {
	for i := range fields {
		ti := textinput.New()
		ti.Placeholder = fields[i].Default
		ti.CharLimit = 100
		if i == 0 {
			ti.Focus()
		}
		fields[i].Input = ti
	}
	return Form{Fields: fields, FocusIndex: 0}
}

func (f *Form) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			return f.nextField()
		case "shift+tab", "up":
			return f.prevField()
		case "enter":
			if f.FocusIndex >= len(f.Fields)-1 {
				if i, ok := f.firstMissingRequired(); ok {
					f.Fields[f.FocusIndex].Input.Blur()
					f.FocusIndex = i
					return f.Fields[f.FocusIndex].Input.Focus()
				}
				f.Submitted = true
				return nil
			}
			return f.nextField()
		case "right":
			if i, ok := f.firstMissingRequired(); ok {
				f.Fields[f.FocusIndex].Input.Blur()
				f.FocusIndex = i
				return f.Fields[f.FocusIndex].Input.Focus()
			}
			f.Submitted = true
			return nil
		case "esc":
			f.Cancelled = true
			return nil
		}
	}

	// Update the focused input
	if f.FocusIndex < len(f.Fields) {
		var cmd tea.Cmd
		f.Fields[f.FocusIndex].Input, cmd = f.Fields[f.FocusIndex].Input.Update(msg)
		return cmd
	}
	return nil
}

func (f *Form) nextField() tea.Cmd {
	if f.FocusIndex < len(f.Fields)-1 {
		f.Fields[f.FocusIndex].Input.Blur()
		f.FocusIndex++
		return f.Fields[f.FocusIndex].Input.Focus()
	}
	return nil
}

func (f *Form) prevField() tea.Cmd {
	if f.FocusIndex > 0 {
		f.Fields[f.FocusIndex].Input.Blur()
		f.FocusIndex--
		return f.Fields[f.FocusIndex].Input.Focus()
	}
	return nil
}

func (f *Form) firstMissingRequired() (int, bool) {
	for i, field := range f.Fields {
		if field.Required && field.Input.Value() == "" && field.Default == "" {
			return i, true
		}
	}
	return 0, false
}

func (f *Form) Value(index int) string {
	if index >= len(f.Fields) {
		return ""
	}
	val := f.Fields[index].Input.Value()
	if val == "" {
		return f.Fields[index].Default
	}
	return val
}

func (f *Form) View() string {
	var lines []string
	for i, field := range f.Fields {
		labelText := field.Label
		if field.Required {
			labelText += StyleRequired.Render("*")
		}
		label := StyleLabel.Render(labelText)
		defaultStr := ""
		if field.Default != "" {
			defaultStr = fmt.Sprintf("[%s]", field.Default)
		}

		var inputView string
		if i == f.FocusIndex {
			inputView = field.Input.View()
		} else {
			val := field.Input.Value()
			if val == "" {
				inputView = StyleHint.Render(defaultStr)
			} else {
				inputView = val
			}
		}

		hint := ""
		if field.Hint != "" && i == f.FocusIndex {
			hint = " " + StyleHint.Render(field.Hint)
		}

		line := fmt.Sprintf(" %s %s %s%s", label, defaultStr, inputView, hint)
		lines = append(lines, line)
	}

	footer := StyleHint.Render("  Tab/Shift-Tab to navigate, Enter to advance/submit, Esc to cancel")
	lines = append(lines, "", footer)

	return strings.Join(lines, "\n")
}
