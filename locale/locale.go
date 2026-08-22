// Package locale holds all user-facing strings for the application.
// To add a new language, define a new Strings value and register it in Load.
package locale

import "github.com/k-leveller/grocr/keybind"

// Strings holds every user-facing string in the application.
// Fields named FmtXxx are Go format strings; the comment on each field lists the
// positional arguments expected by fmt.Sprintf.
type Strings struct {
	// ── Input bar ──────────────────────────────────────────────────────────────
	InputPlaceholder  string // idle UPC input field placeholder
	SearchPlaceholder string // product search input field placeholder

	// ── General status ─────────────────────────────────────────────────────────
	Loading        string // "Loading..."  — panel loading indicator (capital L)
	HintLoading    string // "loading..."  — inline / input-line loading hint (lowercase)
	Cancelled      string
	Processing     string
	HintFormActive string // shown in input line while a form is active

	// ── Error messages  (format strings; arg 1 = error string) ─────────────────
	ErrLoadDefaults           string // %s = error
	ErrLookup                 string // %s = error
	ErrAction                 string // %s = error
	ErrShoppingList           string // %s = error
	ErrExport                 string // %s = error
	ErrMealPlan               string // %s = error
	ErrLoadRecipes            string // %s = error
	ErrLoadRecipeDetail       string // %s = error
	ErrLinkBarcode            string // %s = error
	ErrUpdateName             string // %s = error
	ErrUpdateNotes            string // %s = error
	ErrPriceHistory           string // %s = error
	ErrProductNotFound        string
	ErrProductNotFoundConsume string
	ErrInvalidUPC             string // %s = raw input value
	ErrCannotConsumeTestMode  string
	ErrNoStockFor             string // %s = product name
	ErrSameLocation           string

	// ── Success / info messages ─────────────────────────────────────────────────
	NameUpdated  string
	NotesUpdated string

	// ── Formatted messages ──────────────────────────────────────────────────────
	FmtSpoiled             string // %s = product name
	FmtAddedToShoppingList string // %s = product name
	FmtExportedTo          string // %s = file path
	FmtBarcodeLinkSuccess  string // %s = product name
	FmtLookingUpUPC        string // %s = UPC string

	// ── Header bar ─────────────────────────────────────────────────────────────
	ModeAdd     string // badge shown in add mode
	ModeConsume string // badge shown in consume mode
	ModeLookup  string // badge shown in lookup mode
	HeaderHints string // %s = quit, mode, search, new and help keys

	// ── Input line hints ───────────────────────────────────────────────────────
	// Hints that mention rebindable keys are format strings; the arguments are
	// key symbols supplied by ui.Hint* (see ui/hints.go).
	HintLookupView     string // %s = notes, price history, transfer keys
	HintTransfer       string
	HintPriceHistory   string // %s = up, down, price history keys
	HintMealPlan       string // %s = up, down, refresh, quit keys
	HintTodayMealPlan  string // %s = refresh, quit keys
	HintRecipeList     string // %s = up, down, search, right (with arrow), refresh, quit keys
	HintRecipeDetail   string // %s = up, down, left (with arrow), quit keys
	HintEditNotes      string
	HintYesNo          string // %s = yes key
	HintUnknownBarcode string // %s = create, link keys
	HintExpiringDetail string // %s = left (with arrow), down, up keys

	// ── Form ───────────────────────────────────────────────────────────────────
	FormNavHint         string // navigation hint shown below every form
	FieldName           string
	FieldDisplayName    string
	FieldQtyUnit        string
	FieldExpires        string
	FieldExpiresHint    string // suffix appended to expiry hint
	FieldLocation       string
	FieldStore          string
	FieldQuantity       string
	FieldPrice          string
	FieldPriceHint      string
	FieldFrom           string
	FieldTo             string
	FmtShelfFromSource  string // %d = days, %s = source label ("OFF", "name", "category")
	FmtShelfFromProduct string // %d = days

	// ── Search ─────────────────────────────────────────────────────────────────
	SearchHeader  string
	AllLocations  string
	NoMatches     string
	SearchNavHint string

	// ── Expiring soon panel ────────────────────────────────────────────────────
	ExpiringSoonHeader   string
	NoneExpiringSoon     string
	ExpiringDetailHeader string

	// ── Product info labels ────────────────────────────────────────────────────
	LabelUPC         string
	LabelProduct     string
	LabelInStock     string
	LabelShelfLife   string
	LabelLocation    string
	LabelExpires     string
	LabelStore       string
	LabelLastPrice   string
	LabelNotes       string
	LabelOFF         string // short label used in product-info panel
	LabelOFFName     string // label used in lookup-view panel
	LabelCategories  string
	NewProductBadge  string
	FoundInGrocy     string
	OpenFoodFacts    string // parenthetical source tag
	ShelfLifeNever   string
	FmtDays          string // %d = number of days
	FmtProductIDHint string // %d = product ID

	// ── Lookup view ───────────────────────────────────────────────────────────
	ProductOverview string

	// ── Edit views ────────────────────────────────────────────────────────────
	EditNameLabel string
	EditNameHint  string
	NotesLabel    string

	// ── Consume ───────────────────────────────────────────────────────────────
	ConsumingLabel     string
	ConsumeQtyLabel    string
	ConsumeQtyHint     string
	FmtInStock         string // %g = stock amount
	ErrInvalidQty      string // %s = entered value
	FmtQtyExceedsStock string // %g = stock amount

	// ── Transfer ──────────────────────────────────────────────────────────────
	TransferStockHeader string

	// ── Shopping list prompt ───────────────────────────────────────────────────
	ConsumedLabel      string
	StockNowZero       string
	ShoppingListPrompt string

	// ── Unknown barcode ───────────────────────────────────────────────────────
	UnknownBarcodePrompt string
	UnknownBarcodeCreate string
	UnknownBarcodeLink   string

	// ── Price history ─────────────────────────────────────────────────────────
	PriceHistoryHeader string
	NoPriceHistory     string
	NoPriceData        string

	// ── Meal plan ─────────────────────────────────────────────────────────────
	MealPlanHeader      string
	NoMealsThisWeek     string
	Today               string // parenthetical label "(Today)"
	Tomorrow            string // parenthetical label "(Tomorrow)"
	TodayMealPlanHeader string
	NoMealsToday        string
	ThisWeek            string
	NoMoreMealsThisWeek string
	FmtMoreItems        string // %d = extra item count
	FmtTomorrowLabel    string // %s = date string (e.g. "Jan 02")

	// ── Recipes ───────────────────────────────────────────────────────────────
	RecipesHeader         string
	NoRecipes             string
	NoRecipeDescription   string
	FmtMissingIngredients string // %d = missing ingredient count
	FmtRecipeID           string // %d = recipe ID (fallback when name is unknown)
	FmtServings           string // %.0f = serving count
	FmtProductID          string // %d = product ID (fallback when name is unknown)
	FmtProductAmount      string // %.0f = product amount

	// ── Activity log ─────────────────────────────────────────────────────────
	// Args: %s = styled icon, %s = product name, %g = quantity
	FmtLogConsumed string
	FmtLogSpoiled  string
	// Args: %s = styled icon, %s = product name, %g = quantity, %s = location
	FmtLogTransfer string
	// Args: %s = styled icon, %s = product name, %g = quantity
	FmtLogAdd    string
	LogAddTo     string // %s = location name
	LogAddExpiry string // %s = date string

	// ── Help overlay ──────────────────────────────────────────────────────────
	HelpTitle string
	// HelpDescs describes each rebindable action; the key column is rendered
	// from the active keybinds. Missing entries are skipped.
	HelpDescs map[keybind.Action]string
	// HelpStaticRows lists the shortcuts that cannot be rebound, one per line,
	// as "key\tdescription". The key column is padded to match the bound rows.
	HelpStaticRows          string
	HelpUnknownBarcodeTitle string
	HelpUnknownBarcodeBody  string // %s = create key, %s = link key
}

// English is the built-in English locale.
var English = Strings{
	// Input bar
	InputPlaceholder:  "Scan or enter UPC...",
	SearchPlaceholder: "type to search...",

	// General status
	Loading:        "Loading...",
	HintLoading:    "loading...",
	Cancelled:      "Cancelled",
	Processing:     "Processing...",
	HintFormActive: "(form active)",

	// Error messages
	ErrLoadDefaults:           "Failed to load Grocy defaults: %s",
	ErrLookup:                 "Lookup error: %s",
	ErrAction:                 "Error: %s",
	ErrShoppingList:           "Shopping list error: %s",
	ErrExport:                 "Export failed: %s",
	ErrMealPlan:               "Meal plan unavailable: %s",
	ErrLoadRecipes:            "Failed to load recipes: %s",
	ErrLoadRecipeDetail:       "Failed to load recipe: %s",
	ErrLinkBarcode:            "Link failed: %s",
	ErrUpdateName:             "Error updating name: %s",
	ErrUpdateNotes:            "Error updating notes: %s",
	ErrPriceHistory:           "Price history unavailable: %s",
	ErrProductNotFound:        "Product not found in Grocy",
	ErrProductNotFoundConsume: "Product not found in Grocy — cannot consume",
	ErrInvalidUPC:             "Invalid UPC: %s",
	ErrCannotConsumeTestMode:  "cannot consume in test mode",
	ErrNoStockFor:             "no stock on hand for %s",
	ErrSameLocation:           "from and to locations must differ",

	// Success / info
	NameUpdated:  "Name updated",
	NotesUpdated: "Notes updated",

	// Formatted messages
	FmtSpoiled:             "Spoiled: %s",
	FmtAddedToShoppingList: "%s added to shopping list",
	FmtExportedTo:          "Exported to %s",
	FmtBarcodeLinkSuccess:  "Barcode linked to %s",
	FmtLookingUpUPC:        "Looking up UPC %s...",

	// Header
	ModeAdd:     "[ADD]",
	ModeConsume: "[EAT]",
	ModeLookup:  "[LOOK]",
	HeaderHints: "%s:quit %s:mode %s:search %s:new %s:help",

	// Input line hints
	HintLookupView:     "%s = notes  •  %s = price history  •  %s = transfer  •  Esc/Enter = dismiss",
	HintTransfer:       "Tab/↓ next field  •  Enter submit  •  Esc cancel",
	HintPriceHistory:   "↑/↓/%s/%s = navigate  •  Esc/%s = back",
	HintMealPlan:       "↑/↓/%s/%s = scroll  •  %s = refresh  •  Esc/%s = back",
	HintTodayMealPlan:  "%s = refresh  •  Esc/%s = back",
	HintRecipeList:     "↑/↓/%s/%s = navigate  •  %s = search  •  Enter/%s = view  •  %s = refresh  •  Esc/%s = back",
	HintRecipeDetail:   "↑/↓/%s/%s = scroll  •  %s/%s/Esc = back",
	HintEditNotes:      "Enter to save  •  Esc to cancel",
	HintYesNo:          "%s = yes, any other key = no",
	HintUnknownBarcode: "%s/Enter = create new  •  %s = link existing  •  Esc = cancel",
	HintExpiringDetail: "%s / Esc = back  •  %s/%s = navigate",

	// Form
	FormNavHint:         "Tab/↓ to navigate fields, Enter to submit, Esc to cancel",
	FieldName:           "Name",
	FieldDisplayName:    "Display name",
	FieldQtyUnit:        "Qty unit",
	FieldExpires:        "Expires",
	FieldExpiresHint:    "(YYYY-MM-DD, days, or blank=never)",
	FieldLocation:       "Location",
	FieldStore:          "Store",
	FieldQuantity:       "Quantity",
	FieldPrice:          "Price",
	FieldPriceHint:      "total price paid",
	FieldFrom:           "From",
	FieldTo:             "To",
	FmtShelfFromSource:  "~%dd from %s",
	FmtShelfFromProduct: "~%dd from product",

	// Search
	SearchHeader:  "Search:",
	AllLocations:  "all locations",
	NoMatches:     "No matches",
	SearchNavHint: "↑/↓ navigate · Tab filter location · Enter select · %s new · Esc cancel",

	// Expiring soon panel
	ExpiringSoonHeader:   "Expiring Soon",
	NoneExpiringSoon:     "None expiring soon",
	ExpiringDetailHeader: "Item Details",

	// Product info labels
	LabelUPC:         "UPC:",
	LabelProduct:     "Product:",
	LabelInStock:     "In stock:",
	LabelShelfLife:   "Shelf life:",
	LabelLocation:    "Location:",
	LabelExpires:     "Expires:",
	LabelStore:       "Store:",
	LabelLastPrice:   "Last price:",
	LabelNotes:       "Notes:",
	LabelOFF:         "OFF:",
	LabelOFFName:     "OFF name:",
	LabelCategories:  "Categories:",
	NewProductBadge:  "NEW PRODUCT",
	FoundInGrocy:     "✓ Found in Grocy",
	OpenFoodFacts:    "(Open Food Facts)",
	ShelfLifeNever:   "never",
	FmtDays:          "%d days",
	FmtProductIDHint: "[id:%d]",

	// Lookup view
	ProductOverview: "Product Overview",

	// Edit views
	EditNameLabel: "Edit name: ",
	EditNameHint:  "Enter to save, Esc to cancel",
	NotesLabel:    "Notes: ",

	// Consume
	ConsumingLabel:     "Consuming:",
	ConsumeQtyLabel:    "Quantity to consume: ",
	ConsumeQtyHint:     "Enter to confirm, Esc to cancel",
	FmtInStock:         "%g in stock",
	ErrInvalidQty:      "invalid quantity: %s",
	FmtQtyExceedsStock: "only %g in stock",

	// Transfer
	TransferStockHeader: "Transfer stock",

	// Shopping list prompt
	ConsumedLabel:      "Consumed:",
	StockNowZero:       "Stock is now zero.",
	ShoppingListPrompt: "Add to shopping list? [%s/N]",

	// Unknown barcode
	UnknownBarcodePrompt: "Unknown barcode — what would you like to do?",
	UnknownBarcodeCreate: "  [%s] / Enter  Create a new product",
	UnknownBarcodeLink:   "  [%s]          Link to an existing product",

	// Price history
	PriceHistoryHeader: "Price History:",
	NoPriceHistory:     "No purchase history found.",
	NoPriceData:        "(no price)",

	// Meal plan
	MealPlanHeader:      "Meal Plan — Next 7 Days",
	NoMealsThisWeek:     "No meals planned for the next 7 days.",
	Today:               "(Today)",
	Tomorrow:            "(Tomorrow)",
	TodayMealPlanHeader: "Today's Meal Plan",
	NoMealsToday:        "Nothing planned for today.",
	ThisWeek:            "This Week",
	NoMoreMealsThisWeek: "No more meals planned this week.",
	FmtMoreItems:        "+%d more",
	FmtTomorrowLabel:    "Tomorrow %s",

	// Recipes
	RecipesHeader:         "Recipes",
	NoRecipes:             "No recipes found.",
	NoRecipeDescription:   "No description available.",
	FmtMissingIngredients: "(%d missing)",
	FmtRecipeID:           "Recipe #%d",
	FmtServings:           "(%.0f srv)",
	FmtProductID:          "Product #%d",
	FmtProductAmount:      "×%.0f",

	// Activity log
	FmtLogConsumed: "%s %s x%g consumed",
	FmtLogSpoiled:  "%s %s x%g spoiled",
	FmtLogTransfer: "%s %s x%g %s",
	FmtLogAdd:      "%s %s x%g",
	LogAddTo:       "→ %s",
	LogAddExpiry:   "(%s)",

	// Help overlay
	HelpTitle: "Keyboard Shortcuts",
	HelpDescs: map[keybind.Action]string{
		keybind.Quit:          "Quit",
		keybind.Mode:          "Cycle Add / Consume / Lookup mode",
		keybind.Search:        "Search products by name",
		keybind.NewProduct:    "New product (no barcode)",
		keybind.Export:        "Export stock to CSV",
		keybind.MealPlanToday: "Today's meal plan (compact)",
		keybind.MealPlan:      "Meal plan (next 7 days, scrollable)",
		keybind.Recipes:       "Recipe list (fulfillability), Enter/→ to view details",
		keybind.EditName:      "Edit product name",
		keybind.Notes:         "Edit notes (lookup view)",
		keybind.PriceHistory:  "Price history (lookup view)",
		keybind.Transfer:      "Transfer stock between locations (lookup view)",
		keybind.Help:          "Toggle this help",
		keybind.Up:            "Older UPC history entry (or expiry panel up when no history)",
		keybind.Down:          "Newer UPC history entry (or expiry panel down)",
		keybind.Consume:       "Mark as consumed (expiring-soon panel)",
		keybind.Spoil:         "Mark as spoiled (expiring-soon panel)",
		keybind.Refresh:       "Reload the current panel",
		keybind.Right:         "Open item details (expiring-soon panel)",
		keybind.Left:          "Go back",
	},
	HelpStaticRows: "" +
		"Tab/↓\tNext field (form)\n" +
		"S-Tab/↑\tPrevious field (form)\n" +
		"Enter\tSubmit form\n" +
		"Esc\tCancel current scan\n" +
		"Ctrl+C\tQuit\n",
	HelpUnknownBarcodeTitle: "Unknown Barcode",
	HelpUnknownBarcodeBody: "" +
		"  %s/Enter  Create a new product for this barcode\n" +
		"  %s        Link barcode to an existing product (rebrand/repackage)\n",
}

// Active is the locale used at runtime. It is set by Load and defaults to English.
var Active = &English

// Load sets the active locale based on the given language tag (e.g. "en").
// Unrecognised tags fall back to English.
func Load(lang string) {
	switch lang {
	case "en", "":
		Active = &English
	default:
		Active = &English
	}
}
