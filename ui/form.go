package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/k-leveller/grocr/locale"
)

type FormField struct {
	Label    string
	Default  string
	Hint     string
	Required bool
	// Options are the numbered choices offered in Hint, in the same order.
	// Typing the number of a choice replaces the input with the choice itself.
	Options []string
	Input   textinput.Model
}

type Form struct {
	Fields      []FormField
	FocusIndex  int
	Submitted   bool
	Cancelled   bool

	// pendingNumber is the option number a live expansion replaced, and
	// pendingValue the option name that replaced it. The swap is provisional:
	// typing on undoes it, so free text that merely starts with an option
	// number stays reachable.
	pendingNumber string
	pendingValue  string
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
			f.clearPending()
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
		f.undoExpansion(msg)
	}

	// Update the focused input
	if f.FocusIndex < len(f.Fields) {
		var cmd tea.Cmd
		f.Fields[f.FocusIndex].Input, cmd = f.Fields[f.FocusIndex].Input.Update(msg)
		f.expandOption()
		return cmd
	}
	return nil
}

// undoExpansion puts back the number a live expansion replaced, before the key
// that triggered it reaches the input. Typing on an expansion therefore builds
// on the digits rather than on the option name, so a store named "7-Eleven"
// stays typeable even where 7 already names a store; a second digit likewise
// re-selects, turning "1" into "12". Any key that is not plain editing simply
// commits the expansion.
func (f *Form) undoExpansion(msg tea.KeyMsg) {
	if f.pendingValue == "" || f.FocusIndex >= len(f.Fields) {
		return
	}
	switch msg.Type {
	case tea.KeyRunes, tea.KeySpace, tea.KeyBackspace, tea.KeyDelete:
	default:
		f.clearPending()
		return
	}
	in := &f.Fields[f.FocusIndex].Input
	if in.Value() == f.pendingValue {
		in.SetValue(f.pendingNumber)
		in.CursorEnd()
	}
	f.clearPending()
}

// expandOption swaps a typed option number for the option it names, so the
// field reads "oz" the moment "1" is typed.
func (f *Form) expandOption() {
	field := &f.Fields[f.FocusIndex]
	val := field.Input.Value()
	n, ok := optionIndex(val, len(field.Options))
	if !ok {
		return
	}
	name := field.Options[n-1]
	field.Input.SetValue(name)
	field.Input.CursorEnd()
	f.pendingNumber, f.pendingValue = val, name
}

func (f *Form) clearPending() {
	f.pendingNumber, f.pendingValue = "", ""
}

// optionIndex reports the 1-based option val selects, if val is a plain number
// naming one of count options.
func optionIndex(val string, count int) (int, bool) {
	if count == 0 || val == "" || (len(val) > 1 && val[0] == '0') {
		return 0, false
	}
	for _, r := range val {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(val)
	if err != nil || n < 1 || n > count {
		return 0, false
	}
	return n, true
}

func (f *Form) nextField() tea.Cmd {
	f.clearPending()
	if f.FocusIndex < len(f.Fields)-1 {
		f.Fields[f.FocusIndex].Input.Blur()
		f.FocusIndex++
		return f.Fields[f.FocusIndex].Input.Focus()
	}
	return nil
}

func (f *Form) prevField() tea.Cmd {
	f.clearPending()
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

	footer := StyleHint.Render("  " + locale.Active.FormNavHint)
	lines = append(lines, "", footer)

	return strings.Join(lines, "\n")
}
