package keybind

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultsCoverEveryAction(t *testing.T) {
	d := Defaults()
	for _, a := range Order {
		if d[a] == "" {
			t.Errorf("action %q has no default binding", a)
		}
		if comments[a] == "" {
			t.Errorf("action %q has no config-file comment", a)
		}
	}
	if len(d) != len(Order) {
		t.Errorf("Order lists %d actions, defaults has %d", len(Order), len(d))
	}
}

func TestDefaultsIsACopy(t *testing.T) {
	d := Defaults()
	d[Quit] = "Z"
	if Defaults()[Quit] != "q" {
		t.Fatal("mutating a returned map changed the built-in defaults")
	}
	if Default(Quit) != "q" {
		t.Fatalf("Default(Quit) = %q, want q", Default(Quit))
	}
}

func TestParseOverridesAndFallsBack(t *testing.T) {
	m := Parse("quit = Q\nsearch=s\n")

	if got := m.Key(Quit); got != "Q" {
		t.Errorf("Key(Quit) = %q, want Q", got)
	}
	if got := m.Key(Search); got != "s" {
		t.Errorf("Key(Search) = %q, want s", got)
	}
	// Untouched actions keep their defaults.
	if got := m.Key(Mode); got != "m" {
		t.Errorf("Key(Mode) = %q, want m", got)
	}
}

func TestParseIgnoresJunk(t *testing.T) {
	src := `
# a comment
; another comment

quit = Q
not_an_action = z
this line has no equals sign
mode =
   search   =   s
{"json": "nope"}
`
	m := Parse(src)

	want := map[Action]string{
		Quit:   "Q",
		Mode:   "m", // blank value ignored, default kept
		Search: "s",
		Help:   "?",
	}
	for a, k := range want {
		if got := m.Key(a); got != k {
			t.Errorf("Key(%q) = %q, want %q", a, got, k)
		}
	}
	if _, ok := m[Action("not_an_action")]; ok {
		t.Error("unknown action was stored in the map")
	}
	// Every action still resolves.
	for _, a := range Order {
		if m.Key(a) == "" {
			t.Errorf("action %q unbound after parsing junk", a)
		}
	}
}

func TestParseGarbageKeepsAllDefaults(t *testing.T) {
	m := Parse("\x00\x01 binary nonsense \xff\n<<<>>>\n")
	for _, a := range Order {
		if m.Key(a) != Default(a) {
			t.Errorf("action %q = %q, want default %q", a, m.Key(a), Default(a))
		}
	}
}

func TestParseNormalizesSpecialKeys(t *testing.T) {
	m := Parse("quit = Escape\nmode = RETURN\nsearch = CTRL+F\nrefresh = pgdn\nnew = Up\n")

	cases := map[Action]string{
		Quit:       "esc",
		Mode:       "enter",
		Search:     "ctrl+f",
		Refresh:    "pgdown",
		NewProduct: "up",
	}
	for a, want := range cases {
		if got := m.Key(a); got != want {
			t.Errorf("Key(%q) = %q, want %q", a, got, want)
		}
	}
}

// Bubbletea reports the space bar as " " and a shifted letter as the upper-case
// rune, so the friendly spellings have to be translated or the binding is dead.
func TestParseNormalizesSpaceAndShift(t *testing.T) {
	m := Parse("quit = space\nmode = Shift+P\nsearch = SPACEBAR\nrecipes = F5\n")

	cases := map[Action]string{
		Quit:    " ",
		Mode:    "P",
		Search:  " ",
		Recipes: "f5",
	}
	for a, want := range cases {
		if got := m.Key(a); got != want {
			t.Errorf("Key(%q) = %q, want %q", a, got, want)
		}
	}
	if got := Display(" "); got != "Space" {
		t.Errorf("Display(\" \") = %q, want Space", got)
	}
	if got := Display("f5"); got != "F5" {
		t.Errorf("Display(\"f5\") = %q, want F5", got)
	}
}

func TestParseNormalizesAltKeepsCase(t *testing.T) {
	m := Parse("quit = Alt+A\nmode = alt+b\nsearch = shift+tab\n")

	if got := m.Key(Quit); got != "alt+A" {
		t.Errorf("Key(Quit) = %q, want alt+A", got)
	}
	if got := m.Key(Mode); got != "alt+b" {
		t.Errorf("Key(Mode) = %q, want alt+b", got)
	}
	if got := m.Key(Search); got != "shift+tab" {
		t.Errorf("Key(Search) = %q, want shift+tab", got)
	}
}

func TestParseStripsTrailingComments(t *testing.T) {
	m := Parse("quit = Q  # leave the app\nmode = M\t; cycle modes\nsearch = # \nnew = ;\n")

	if got := m.Key(Quit); got != "Q" {
		t.Errorf("Key(Quit) = %q, want Q", got)
	}
	if got := m.Key(Mode); got != "M" {
		t.Errorf("Key(Mode) = %q, want M", got)
	}
	// A marker in the key position is still a bindable key.
	if got := m.Key(Search); got != "#" {
		t.Errorf("Key(Search) = %q, want #", got)
	}
	if got := m.Key(NewProduct); got != ";" {
		t.Errorf("Key(NewProduct) = %q, want ;", got)
	}
}

func TestParseKeepsSingleCharCase(t *testing.T) {
	m := Parse("meal_plan = P\nquit = Q\n")
	if got := m.Key(MealPlan); got != "P" {
		t.Errorf("Key(MealPlan) = %q, want P", got)
	}
	if got := m.Key(Quit); got != "Q" {
		t.Errorf("Key(Quit) = %q, want Q", got)
	}
}

func TestParseActionNameIsCaseInsensitive(t *testing.T) {
	m := Parse("  QUIT  =  Q  \nMeal_Plan = z\n")
	if got := m.Key(Quit); got != "Q" {
		t.Errorf("Key(Quit) = %q, want Q", got)
	}
	if got := m.Key(MealPlan); got != "z" {
		t.Errorf("Key(MealPlan) = %q, want z", got)
	}
}

func TestIs(t *testing.T) {
	m := Parse("quit = Q\n")
	if !m.Is(Quit, "Q") {
		t.Error("Is(Quit, \"Q\") = false, want true")
	}
	if m.Is(Quit, "q") {
		t.Error("Is is not case-sensitive")
	}
	if m.Is(Quit, "") {
		t.Error("an empty key must not match any action")
	}
}

func TestIsWithUnboundActionUsesDefault(t *testing.T) {
	m := Map{}
	if !m.Is(Quit, "q") {
		t.Error("an empty map must still resolve to the default binding")
	}
}

func TestDisplay(t *testing.T) {
	cases := map[string]string{
		"q":         "q",
		"P":         "P",
		"/":         "/",
		"up":        "↑",
		"down":      "↓",
		"left":      "←",
		"right":     "→",
		"enter":     "Enter",
		"esc":       "Esc",
		"shift+tab": "S-Tab",
		"ctrl+e":    "Ctrl+E",
		"alt+x":     "Alt+X",
		"":          "",
	}
	for in, want := range cases {
		if got := Display(in); got != want {
			t.Errorf("Display(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSymUsesActiveMap(t *testing.T) {
	orig := Active
	t.Cleanup(func() { Active = orig })

	Active = Parse("quit = ctrl+q\n")
	if got := Sym(Quit); got != "Ctrl+Q" {
		t.Errorf("Sym(Quit) = %q, want Ctrl+Q", got)
	}
	if !Is(Quit, "ctrl+q") {
		t.Error("Is did not consult the active map")
	}
	if got := Key(Mode); got != "m" {
		t.Errorf("Key(Mode) = %q, want m", got)
	}
}

func TestRenderRoundTrips(t *testing.T) {
	m := Parse(Render())
	for _, a := range Order {
		if got := m.Key(a); got != Default(a) {
			t.Errorf("round-tripped %q = %q, want %q", a, got, Default(a))
		}
	}
}

func TestLoadFromWritesDefaultsWhenMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "keybinds.conf")

	m := LoadFrom(path)
	for _, a := range Order {
		if m.Key(a) != Default(a) {
			t.Fatalf("action %q = %q, want default", a, m.Key(a))
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("default file was not written: %v", err)
	}
	if again := Parse(string(data)); again.Key(Quit) != "q" {
		t.Errorf("written file does not parse back to the defaults")
	}
}

func TestLoadFromReadsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keybinds.conf")
	if err := os.WriteFile(path, []byte("quit = Q\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	m := LoadFrom(path)
	if got := m.Key(Quit); got != "Q" {
		t.Errorf("Key(Quit) = %q, want Q", got)
	}
	if got := m.Key(Mode); got != "m" {
		t.Errorf("Key(Mode) = %q, want m", got)
	}
}

func TestLoadFromUnreadablePathUsesDefaults(t *testing.T) {
	dir := t.TempDir() // a directory is not a readable config file
	m := LoadFrom(dir)
	if got := m.Key(Quit); got != "q" {
		t.Errorf("Key(Quit) = %q, want q", got)
	}
}
