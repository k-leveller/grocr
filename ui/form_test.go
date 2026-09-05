package ui

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestOptionIndex(t *testing.T) {
	tests := []struct {
		name  string
		val   string
		count int
		want  int
		ok    bool
	}{
		{name: "first option", val: "1", count: 3, want: 1, ok: true},
		{name: "last option", val: "3", count: 3, want: 3, ok: true},
		{name: "two digits", val: "12", count: 12, want: 12, ok: true},
		{name: "single digit with many options", val: "1", count: 12, want: 1, ok: true},
		{name: "no options", val: "1", count: 0},
		{name: "empty value", val: "", count: 3},
		{name: "out of range", val: "4", count: 3},
		{name: "zero", val: "0", count: 3},
		{name: "leading zero", val: "01", count: 3},
		{name: "not a number", val: "oz", count: 3},
		{name: "mixed", val: "1x", count: 3},
		{name: "signed", val: "+1", count: 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := optionIndex(tc.val, tc.count)
			if ok != tc.ok || got != tc.want {
				t.Errorf("optionIndex(%q, %d) = %d, %v; want %d, %v",
					tc.val, tc.count, got, ok, tc.want, tc.ok)
			}
		})
	}
}

// typeRunes feeds s to the form one keystroke at a time.
func typeRunes(f *Form, s string) {
	for _, r := range s {
		if r == ' ' {
			f.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
			continue
		}
		f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}

func unitForm() Form {
	return NewForm([]FormField{
		{Label: "Unit", Options: []string{"oz", "lb", "kg"}},
		{Label: "Quantity", Default: "1"},
	})
}

func TestFormExpandsTypedOptionNumber(t *testing.T) {
	f := unitForm()
	typeRunes(&f, "2")

	if got := f.Fields[0].Input.Value(); got != "lb" {
		t.Errorf("input value = %q, want %q", got, "lb")
	}
	if got := f.Value(0); got != "lb" {
		t.Errorf("Value(0) = %q, want %q", got, "lb")
	}
}

func TestFormLeavesNonOptionInputAlone(t *testing.T) {
	f := unitForm()

	typeRunes(&f, "9")
	if got := f.Fields[0].Input.Value(); got != "9" {
		t.Errorf("out-of-range number = %q, want %q", got, "9")
	}

	f.Update(tea.KeyMsg{Type: tea.KeyTab})
	typeRunes(&f, "2")
	if got := f.Fields[1].Input.Value(); got != "2" {
		t.Errorf("option-less field = %q, want %q", got, "2")
	}
}

// Typing past an expansion must build on the digits, not on the option name,
// so a free-text store whose name starts with an option number stays typeable.
func TestFormTypingPastExpansionRestoresNumber(t *testing.T) {
	stores := make([]string, 8)
	for i := range stores {
		stores[i] = fmt.Sprintf("Store%d", i+1)
	}
	f := NewForm([]FormField{{Label: "Store", Options: stores}})

	typeRunes(&f, "7")
	if got := f.Fields[0].Input.Value(); got != "Store7" {
		t.Fatalf("after %q, value = %q, want %q", "7", got, "Store7")
	}

	typeRunes(&f, "-Eleven")
	if got := f.Fields[0].Input.Value(); got != "7-Eleven" {
		t.Errorf("value = %q, want %q", got, "7-Eleven")
	}
}

// A second digit re-selects rather than appending to the first expansion.
func TestFormSecondDigitReselects(t *testing.T) {
	opts := make([]string, 12)
	for i := range opts {
		opts[i] = fmt.Sprintf("u%d", i+1)
	}
	f := NewForm([]FormField{{Label: "Unit", Options: opts}})

	typeRunes(&f, "1")
	if got := f.Fields[0].Input.Value(); got != "u1" {
		t.Fatalf("after %q, value = %q, want %q", "1", got, "u1")
	}

	typeRunes(&f, "2")
	if got := f.Fields[0].Input.Value(); got != "u12" {
		t.Errorf("value = %q, want %q", got, "u12")
	}
}

// One backspace undoes the whole expansion, clearing the single digit typed.
func TestFormBackspaceUndoesExpansion(t *testing.T) {
	f := unitForm()

	typeRunes(&f, "2")
	f.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	if got := f.Fields[0].Input.Value(); got != "" {
		t.Errorf("value = %q, want empty", got)
	}
}

// Leaving the field commits the expansion, so coming back and typing edits the
// option name instead of resurrecting the number.
func TestFormLeavingFieldCommitsExpansion(t *testing.T) {
	f := unitForm()

	typeRunes(&f, "2")
	f.Update(tea.KeyMsg{Type: tea.KeyTab})
	f.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	typeRunes(&f, "s")

	if got := f.Fields[0].Input.Value(); got != "lbs" {
		t.Errorf("value = %q, want %q", got, "lbs")
	}
}

// Submitting keeps the expanded name.
func TestFormSubmitKeepsExpansion(t *testing.T) {
	f := unitForm()

	typeRunes(&f, "3")
	f.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if !f.Submitted {
		t.Fatal("form not submitted")
	}
	if got := f.Value(0); got != "kg" {
		t.Errorf("Value(0) = %q, want %q", got, "kg")
	}
}
