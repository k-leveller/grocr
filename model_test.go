package main

import (
	"testing"
	"time"

	"github.com/k-leveller/grocr/api"
)

// ---- resolveExpiry ----

func TestResolveExpiry(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty becomes never", input: "", want: "2999-12-31"},
		{name: "zero becomes never", input: "0", want: "2999-12-31"},
		{name: "negative becomes never", input: "-10", want: "2999-12-31"},
		{name: "date passthrough", input: "2026-06-15", want: "2026-06-15"},
		{name: "non-numeric passthrough", input: "next-week", want: "next-week"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveExpiry(tc.input)
			if got != tc.want {
				t.Errorf("resolveExpiry(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}

	t.Run("positive days offset from today", func(t *testing.T) {
		expected := time.Now().AddDate(0, 0, 30).Format("2006-01-02")
		got := resolveExpiry("30")
		if got != expected {
			t.Errorf("resolveExpiry(%q) = %q, want %q", "30", got, expected)
		}
	})
}

// ---- evalArith ----

func TestEvalArith(t *testing.T) {
	tests := []struct {
		name  string
		expr  string
		want  float64
		valid bool
	}{
		{name: "simple number", expr: "3.99", want: 3.99, valid: true},
		{name: "addition", expr: "1+2", want: 3, valid: true},
		{name: "subtraction", expr: "5-1.5", want: 3.5, valid: true},
		{name: "multiplication", expr: "2*3.5", want: 7, valid: true},
		{name: "division", expr: "10/4", want: 2.5, valid: true},
		{name: "mixed precedence", expr: "3*2+1.5", want: 7.5, valid: true},
		{name: "scientific notation", expr: "1e3", want: 1000, valid: true},
		{name: "scientific neg exponent", expr: "1e-3", want: 0.001, valid: true},
		{name: "unary minus after operator", expr: "6+-1.5", want: 4.5, valid: true},
		{name: "invalid non-numeric", expr: "abc", want: 0, valid: false},
		{name: "division by zero", expr: "5/0", want: 0, valid: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := evalArith(tc.expr)
			if ok != tc.valid {
				t.Errorf("evalArith(%q) valid = %v, want %v", tc.expr, ok, tc.valid)
			}
			if tc.valid && got != tc.want {
				t.Errorf("evalArith(%q) = %v, want %v", tc.expr, got, tc.want)
			}
		})
	}
}

// ---- prependUPCHistory ----

func TestPrependUPCHistory(t *testing.T) {
	t.Run("prepend to empty", func(t *testing.T) {
		got := prependUPCHistory(nil, "111")
		if len(got) != 1 || got[0] != "111" {
			t.Errorf("unexpected result: %v", got)
		}
	})

	t.Run("deduplicates existing entry", func(t *testing.T) {
		got := prependUPCHistory([]string{"111", "222", "333"}, "222")
		if got[0] != "222" {
			t.Errorf("expected 222 at front, got %v", got)
		}
		for _, h := range got[1:] {
			if h == "222" {
				t.Error("222 should appear only once")
			}
		}
	})

	t.Run("prepends new entry", func(t *testing.T) {
		got := prependUPCHistory([]string{"111", "222"}, "333")
		if got[0] != "333" || got[1] != "111" || got[2] != "222" {
			t.Errorf("unexpected order: %v", got)
		}
	})

	t.Run("caps at max history 50", func(t *testing.T) {
		var h []string
		for i := 0; i < 50; i++ {
			h = append(h, "000")
		}
		got := prependUPCHistory(h, "999")
		if len(got) > 50 {
			t.Errorf("len = %d, want ≤ 50", len(got))
		}
	})
}

// ---- recipeScore ----

func TestRecipeScore(t *testing.T) {
	tests := []struct {
		name    string
		f       api.RecipeFulfillment
		hasData bool
		want    int
	}{
		{name: "no data", f: api.RecipeFulfillment{}, hasData: false, want: 2},
		{name: "fully fulfilled", f: api.RecipeFulfillment{NeedFulfilled: true}, hasData: true, want: 0},
		{name: "with shopping list", f: api.RecipeFulfillment{NeedFulfilledWithShoppingList: true}, hasData: true, want: 1},
		{name: "not fulfilled", f: api.RecipeFulfillment{}, hasData: true, want: 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := recipeScore(tc.f, tc.hasData)
			if got != tc.want {
				t.Errorf("recipeScore() = %d, want %d", got, tc.want)
			}
		})
	}
}

// ---- Model method helpers ----

func newTestModel() Model {
	return Model{
		testMode: true,
		defaults: &api.Defaults{
			LocationID: 1,
			QuID:       1,
			Locations: []api.Location{
				{ID: 1, Name: "Fridge"},
				{ID: 2, Name: "Freezer"},
				{ID: 3, Name: "Pantry"},
			},
			QuantityUnits: []api.QuantityUnit{
				{ID: 1, Name: "Piece"},
				{ID: 2, Name: "Oz"},
				{ID: 3, Name: "Kg"},
			},
			Stores: []api.Store{
				{ID: 10, Name: "Whole Foods"},
				{ID: 11, Name: "Costco"},
			},
		},
	}
}

func TestParseQuantity(t *testing.T) {
	m := newTestModel()
	tests := []struct {
		input string
		want  float64
	}{
		{"1", 1},
		{"2.5", 2.5},
		{"", 1},
		{"0", 1},    // zero → defaults to 1
		{"-1", 1},   // negative → defaults to 1
		{"abc", 1},  // invalid → defaults to 1
		{"  3  ", 3},
	}
	for _, tc := range tests {
		got := m.parseQuantity(tc.input)
		if got != tc.want {
			t.Errorf("parseQuantity(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestParsePrice(t *testing.T) {
	m := newTestModel()
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		{name: "simple price", input: "3.99", want: 3.99},
		{name: "strips dollar sign", input: "$3.99", want: 3.99},
		{name: "empty returns zero", input: "", want: 0},
		{name: "invalid returns zero", input: "abc", want: 0},
		{name: "negative returns zero", input: "-1", want: 0},
		{name: "arithmetic", input: "2*3.99", want: 7.98},
		{name: "addition", input: "2+1.99", want: 3.99},
		{name: "division by zero returns zero", input: "1/0", want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := m.parsePrice(tc.input)
			// Use a small epsilon for floating point comparison
			diff := got - tc.want
			if diff < -0.001 || diff > 0.001 {
				t.Errorf("parsePrice(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestResolveLocation(t *testing.T) {
	m := newTestModel()
	tests := []struct {
		input string
		want  int
	}{
		{"1", 1},       // by 1-based index
		{"2", 2},       // by 1-based index
		{"3", 3},       // by 1-based index
		{"Fridge", 1},  // exact name match
		{"Fre", 2},     // prefix match (Freezer)
		{"Pa", 3},      // prefix match (Pantry)
		{"", 1},        // empty → default
		{"9", 1},       // out-of-range → default
		{"Unknown", 1}, // no match → default
	}
	for _, tc := range tests {
		got := m.resolveLocation(tc.input)
		if got != tc.want {
			t.Errorf("resolveLocation(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestResolveQuantityUnit(t *testing.T) {
	m := newTestModel()
	tests := []struct {
		input string
		want  int
	}{
		{"1", 1},      // 1-based index → Piece
		{"2", 2},      // 1-based index → Oz
		{"Piece", 1},  // exact name
		{"Oz", 2},     // exact name
		{"k", 3},      // prefix match (Kg)
		{"", 1},       // empty → default QuID
		{"99", 1},     // out of range → default QuID
		{"xyz", 1},    // no match → default QuID
	}
	for _, tc := range tests {
		got := m.resolveQuantityUnit(tc.input)
		if got != tc.want {
			t.Errorf("resolveQuantityUnit(%q) = %d, want %d", tc.input, got, tc.want)
		}
	}
}

func TestResolveOrCreateStore(t *testing.T) {
	m := newTestModel()

	t.Run("empty string returns 0", func(t *testing.T) {
		id, err := m.resolveOrCreateStore("")
		if err != nil || id != 0 {
			t.Errorf("resolveOrCreateStore(%q) = (%d, %v), want (0, nil)", "", id, err)
		}
	})

	t.Run("by 1-based index", func(t *testing.T) {
		id, err := m.resolveOrCreateStore("1")
		if err != nil || id != 10 {
			t.Errorf("resolveOrCreateStore(\"1\") = (%d, %v), want (10, nil)", id, err)
		}
	})

	t.Run("by name prefix", func(t *testing.T) {
		id, err := m.resolveOrCreateStore("Cos")
		if err != nil || id != 11 {
			t.Errorf("resolveOrCreateStore(\"Cos\") = (%d, %v), want (11, nil)", id, err)
		}
	})

	t.Run("no match in testMode returns 0", func(t *testing.T) {
		id, err := m.resolveOrCreateStore("Unknown Store")
		if err != nil || id != 0 {
			t.Errorf("resolveOrCreateStore(\"Unknown Store\") = (%d, %v), want (0, nil)", id, err)
		}
	})
}

func TestLocationName(t *testing.T) {
	m := newTestModel()
	tests := []struct {
		id   int
		want string
	}{
		{1, "Fridge"},
		{2, "Freezer"},
		{3, "Pantry"},
		{99, ""},
	}
	for _, tc := range tests {
		got := m.locationName(tc.id)
		if got != tc.want {
			t.Errorf("locationName(%d) = %q, want %q", tc.id, got, tc.want)
		}
	}
}

func TestDaysFromExpiry(t *testing.T) {
	m := newTestModel()

	t.Run("never sentinel returns -1", func(t *testing.T) {
		got := m.daysFromExpiry("2999-12-31")
		if got == nil {
			t.Fatal("expected non-nil")
		}
		if *got != -1 {
			t.Errorf("daysFromExpiry(\"2999-12-31\") = %d, want -1", *got)
		}
	})

	t.Run("empty string returns nil", func(t *testing.T) {
		got := m.daysFromExpiry("")
		if got != nil {
			t.Errorf("expected nil, got %d", *got)
		}
	})

	t.Run("invalid date returns nil", func(t *testing.T) {
		got := m.daysFromExpiry("not-a-date")
		if got != nil {
			t.Errorf("expected nil, got %d", *got)
		}
	})

	t.Run("past date returns nil", func(t *testing.T) {
		got := m.daysFromExpiry("2000-01-01")
		if got != nil {
			t.Errorf("expected nil for past date, got %d", *got)
		}
	})

	t.Run("future date returns positive days", func(t *testing.T) {
		future := time.Now().AddDate(0, 0, 30).Format("2006-01-02")
		got := m.daysFromExpiry(future)
		if got == nil {
			t.Fatal("expected non-nil for future date")
		}
		if *got < 28 || *got > 31 {
			t.Errorf("expected ~30 days, got %d", *got)
		}
	})
}
