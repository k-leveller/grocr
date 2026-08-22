// Package keybind holds the user-configurable key bindings.
//
// Bindings live in a small flat text file (one "action = key" line per binding)
// so they are quick to edit by hand. The defaults below are the authoritative
// copy: if the file is missing, unreadable, or partly corrupted, every action
// the file fails to supply falls back to its default and the app keeps working.
package keybind

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Action names a rebindable command. The string value is the name used in the
// config file.
type Action string

const (
	Quit          Action = "quit"
	Mode          Action = "mode"
	Search        Action = "search"
	NewProduct    Action = "new"
	Export        Action = "export"
	MealPlanToday Action = "meal_plan_today"
	MealPlan      Action = "meal_plan"
	Recipes       Action = "recipes"
	EditName      Action = "edit_name"
	Notes         Action = "notes"
	PriceHistory  Action = "price_history"
	Transfer      Action = "transfer"
	Help          Action = "help"
	Up            Action = "up"
	Down          Action = "down"
	Left          Action = "left"
	Right         Action = "right"
	Consume       Action = "consume"
	Spoil         Action = "spoil"
	Refresh       Action = "refresh"
	Create        Action = "create"
	Link          Action = "link"
	Yes           Action = "yes"
)

// Order lists every action in the order they are written to the config file.
var Order = []Action{
	Quit, Mode, Search, NewProduct, Export,
	MealPlanToday, MealPlan, Recipes,
	EditName, Notes, PriceHistory, Transfer,
	Help,
	Up, Down, Left, Right,
	Consume, Spoil, Refresh,
	Create, Link, Yes,
}

// defaults is the defensive in-code copy of the stock bindings. It is never
// mutated; Defaults returns a copy.
var defaults = map[Action]string{
	Quit:          "q",
	Mode:          "m",
	Search:        "/",
	NewProduct:    "n",
	Export:        "x",
	MealPlanToday: "t",
	MealPlan:      "P",
	Recipes:       "r",
	EditName:      "e",
	Notes:         "n",
	PriceHistory:  "p",
	Transfer:      "t",
	Help:          "?",
	Up:            "k",
	Down:          "j",
	Left:          "h",
	Right:         "l",
	Consume:       "c",
	Spoil:         "d",
	Refresh:       "r",
	Create:        "c",
	Link:          "l",
	Yes:           "y",
}

// comments describes each action in the generated config file.
var comments = map[Action]string{
	Quit:          "quit the app / leave a panel",
	Mode:          "cycle Add / Consume / Lookup mode",
	Search:        "search products by name (also filters the recipe list)",
	NewProduct:    "create a product with no barcode",
	Export:        "export stock to CSV",
	MealPlanToday: "today's meal plan",
	MealPlan:      "meal plan for the next 7 days",
	Recipes:       "recipe list",
	EditName:      "edit the product name",
	Notes:         "edit product notes (lookup view)",
	PriceHistory:  "price history (lookup view)",
	Transfer:      "transfer stock between locations (lookup view)",
	Help:          "toggle the help overlay",
	Up:            "move up / older UPC history entry",
	Down:          "move down / newer UPC history entry",
	Left:          "go back",
	Right:         "open / drill in",
	Consume:       "mark as consumed (expiring-soon panel)",
	Spoil:         "mark as spoiled (expiring-soon panel)",
	Refresh:       "reload the current panel",
	Create:        "create a new product (unknown barcode)",
	Link:          "link to an existing product (unknown barcode)",
	Yes:           "confirm a yes/no prompt",
}

// Map holds the resolved binding for every action.
type Map map[Action]string

// Active is the map used at runtime. It defaults to the built-in bindings and
// is replaced by Load.
var Active = Defaults()

// Defaults returns a fresh copy of the built-in bindings.
func Defaults() Map {
	m := make(Map, len(defaults))
	for a, k := range defaults {
		m[a] = k
	}
	return m
}

// Default returns the built-in binding for an action.
func Default(a Action) string { return defaults[a] }

// Key returns the key bound to an action, falling back to the built-in default
// when the action is unbound.
func (m Map) Key(a Action) string {
	if k, ok := m[a]; ok && k != "" {
		return k
	}
	return defaults[a]
}

// Is reports whether key triggers the given action.
func (m Map) Is(a Action, key string) bool {
	return key != "" && key == m.Key(a)
}

// Key returns the key bound to an action in the active map.
func Key(a Action) string { return Active.Key(a) }

// Is reports whether key triggers the given action in the active map.
func Is(a Action, key string) bool { return Active.Is(a, key) }

// Sym returns the display symbol for an action's binding in the active map,
// e.g. "↑" for "up" or "Ctrl+E" for "ctrl+e".
func Sym(a Action) string { return Display(Active.Key(a)) }

// displayNames maps special bubbletea key names to the symbol shown in the UI.
var displayNames = map[string]string{
	" ":         "Space",
	"up":        "↑",
	"down":      "↓",
	"left":      "←",
	"right":     "→",
	"enter":     "Enter",
	"esc":       "Esc",
	"tab":       "Tab",
	"shift+tab": "S-Tab",
	"space":     "Space",
	"backspace": "Bksp",
	"home":      "Home",
	"end":       "End",
	"pgup":      "PgUp",
	"pgdown":    "PgDn",
	"delete":    "Del",
	"insert":    "Ins",
}

// Display renders a bubbletea key name the way it should be shown to the user.
func Display(key string) string {
	if key == "" {
		return ""
	}
	if s, ok := displayNames[key]; ok {
		return s
	}
	if rest, ok := strings.CutPrefix(key, "ctrl+"); ok {
		return "Ctrl+" + strings.ToUpper(rest)
	}
	if rest, ok := strings.CutPrefix(key, "alt+"); ok {
		return "Alt+" + strings.ToUpper(rest)
	}
	if rest, ok := strings.CutPrefix(key, "shift+"); ok {
		return "Shift+" + strings.ToUpper(rest)
	}
	if isFunctionKey(key) {
		return strings.ToUpper(key)
	}
	return key
}

// isFunctionKey reports whether key is an "f<n>" function key name.
func isFunctionKey(key string) bool {
	rest, ok := strings.CutPrefix(strings.ToLower(key), "f")
	if !ok || rest == "" {
		return false
	}
	for _, r := range rest {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Path returns the location of the keybind config file.
func Path() string {
	return filepath.Join(os.Getenv("HOME"), ".config", "grocr", "keybinds.conf")
}

// Load reads the keybind file at Path and installs the result as Active. A
// missing file is written out with the defaults (best effort). Any binding the
// file does not supply keeps its built-in default, so a corrupted file can
// never leave an action unbound.
func Load() Map {
	Active = LoadFrom(Path())
	return Active
}

// LoadFrom reads bindings from path, filling in defaults for anything absent
// or malformed. It writes the default file when path does not exist.
func LoadFrom(path string) Map {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			_ = WriteDefaults(path)
		}
		return Defaults()
	}
	return Parse(string(data))
}

// Parse reads "action = key" lines, ignoring blank lines, comments (# or ;),
// unknown actions and malformed lines. Every action absent from the input keeps
// its built-in default.
func Parse(src string) Map {
	m := Defaults()

	valid := make(map[Action]bool, len(Order))
	for _, a := range Order {
		valid[a] = true
	}

	sc := bufio.NewScanner(strings.NewReader(src))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		name, key, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		action := Action(strings.ToLower(strings.TrimSpace(name)))
		key = strings.TrimSpace(stripTrailingComment(key))
		if key == "" || !valid[action] {
			continue
		}
		if key = normalize(key); !isValidKey(key) || reserved[key] {
			continue
		}
		m[action] = key
	}

	return m
}

// stripTrailingComment removes an end-of-line comment from a binding value. The
// comment marker must be preceded by whitespace so that "#" and ";" stay
// bindable keys in their own right.
func stripTrailingComment(val string) string {
	val = strings.TrimSpace(val)
	for i := 1; i < len(val); i++ {
		if val[i] != '#' && val[i] != ';' {
			continue
		}
		if prev := val[i-1]; prev == ' ' || prev == '\t' {
			return strings.TrimSpace(val[:i])
		}
	}
	return val
}

// knownKeys are the multi-character key names bubbletea reports, spelled the
// way it spells them.
var knownKeys = map[string]bool{
	"up": true, "down": true, "left": true, "right": true,
	"enter": true, "esc": true, "tab": true, "backspace": true,
	"delete": true, "insert": true, "home": true, "end": true,
	"pgup": true, "pgdown": true,
	"shift+tab": true,
	"shift+up":  true, "shift+down": true, "shift+left": true, "shift+right": true,
	"shift+home": true, "shift+end": true,
	"ctrl+up": true, "ctrl+down": true, "ctrl+left": true, "ctrl+right": true,
	"ctrl+home": true, "ctrl+end": true, "ctrl+pgup": true, "ctrl+pgdown": true,
	"ctrl+shift+up": true, "ctrl+shift+down": true,
	"ctrl+shift+left": true, "ctrl+shift+right": true,
	"ctrl+shift+home": true, "ctrl+shift+end": true,
}

// reserved keys drive the parts of the UI that are not rebindable — form
// navigation, submitting and cancelling. Binding an action to one of them would
// shadow that behaviour, so such lines are ignored.
var reserved = map[string]bool{
	"enter":     true,
	"esc":       true,
	"tab":       true,
	"shift+tab": true,
	"ctrl+c":    true,
}

// isValidKey reports whether key is a key name bubbletea can actually report.
// Anything else — a typo such as "ctlr+q" — would be permanently unreachable.
func isValidKey(key string) bool {
	if len([]rune(key)) == 1 {
		return true
	}
	if knownKeys[key] || isFunctionKey(key) {
		return true
	}
	for _, prefix := range []string{"ctrl+", "alt+"} {
		rest, ok := strings.CutPrefix(key, prefix)
		if !ok {
			continue
		}
		return len([]rune(rest)) == 1 || knownKeys[rest] || isFunctionKey(rest)
	}
	return false
}

// keyAliases maps friendly spellings onto the name bubbletea reports.
var keyAliases = map[string]string{
	"escape":   "esc",
	"return":   "enter",
	"pagedown": "pgdown",
	"pgdn":     "pgdown",
	"pageup":   "pgup",
	"del":      "delete",
	"ins":      "insert",
	"space":    " ",
	"spacebar": " ",
}

// normalize converts human spellings of special keys into the names bubbletea
// reports. Single characters are left alone so bindings stay case-sensitive.
func normalize(key string) string {
	if len([]rune(key)) == 1 {
		return key
	}

	lower := strings.ToLower(key)
	if alias, ok := keyAliases[lower]; ok {
		return alias
	}
	if knownKeys[lower] || isFunctionKey(lower) {
		return lower
	}
	// bubbletea always reports ctrl combinations in lower case.
	if strings.HasPrefix(lower, "ctrl+") {
		return lower
	}
	// Alt keeps the modified character's case: alt+A is a different key event
	// from alt+a.
	if strings.HasPrefix(lower, "alt+") {
		return "alt+" + key[len("alt+"):]
	}
	// Shifted letters arrive as the upper-case rune, not as "shift+x".
	if rest, ok := strings.CutPrefix(lower, "shift+"); ok {
		if len([]rune(rest)) == 1 {
			return strings.ToUpper(rest)
		}
		return lower
	}
	return key
}

// Render returns the contents of a keybind file holding the default bindings.
func Render() string {
	var b strings.Builder
	b.WriteString("# grocr keybinds — one binding per line, written as: action = key\n")
	b.WriteString("# Lines starting with # or ; are ignored. Delete this file to restore defaults.\n")
	b.WriteString("#\n")
	b.WriteString("# Keys are case-sensitive: \"P\" means Shift+P. Special keys can be spelled\n")
	b.WriteString("# out: up, down, left, right, enter, esc, tab, space, backspace, pgup,\n")
	b.WriteString("# pgdown, home, end, delete, or ctrl+<key> / alt+<key>.\n")
	b.WriteString("#\n")
	b.WriteString("# A \" #\" or \" ;\" after a binding starts a trailing comment.\n")
	b.WriteString("#\n")
	b.WriteString("# Enter, Esc, Tab, Shift+Tab and Ctrl+C are reserved by the UI and cannot be\n")
	b.WriteString("# bound; lines that try to are ignored, as are keys that are not real key\n")
	b.WriteString("# names.\n")
	b.WriteString("#\n")
	b.WriteString("# Unknown actions and unparseable lines are ignored, and any action left out\n")
	b.WriteString("# keeps its default, so a broken file never breaks the app.\n\n")

	width := 0
	for _, a := range Order {
		if len(a) > width {
			width = len(a)
		}
	}
	for _, a := range Order {
		b.WriteString("# " + comments[a] + "\n")
		b.WriteString(string(a) + strings.Repeat(" ", width-len(a)) + " = " + defaults[a] + "\n\n")
	}
	return b.String()
}

// WriteDefaults writes the default keybind file to path, creating its parent
// directory if needed.
func WriteDefaults(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(Render()), 0o644)
}
