// Package locale holds all user-facing strings for the application.
// To add a new language, define a new Strings value and register it in Load.
package locale

// Strings holds every user-facing string in the application.
// Fields named FmtXxx are Go format strings; the comment on each field lists the
// positional arguments expected by fmt.Sprintf.
type Strings struct {
	// ── Input bar ──────────────────────────────────────────────────────────────
	InputPlaceholder  string // idle UPC input field placeholder
	SearchPlaceholder string // product search input field placeholder

	// ── General status ─────────────────────────────────────────────────────────
	Loading     string // "Loading..."  — panel loading indicator (capital L)
	HintLoading string // "loading..."  — inline / input-line loading hint (lowercase)
	Cancelled   string
	Processing  string
	HintFormActive string // shown in input line while a form is active

	// ── Error messages  (format strings; arg 1 = error string) ─────────────────
	ErrLoadDefaults         string // %s = error
	ErrLookup               string // %s = error
	ErrAction               string // %s = error
	ErrShoppingList         string // %s = error
	ErrExport               string // %s = error
	ErrMealPlan             string // %s = error
	ErrLoadRecipes          string // %s = error
	ErrLoadRecipeDetail     string // %s = error
	ErrLinkBarcode          string // %s = error
	ErrUpdateName           string // %s = error
	ErrUpdateNotes          string // %s = error
	ErrPriceHistory         string // %s = error
	ErrProductNotFound      string
	ErrProductNotFoundConsume string
	ErrInvalidUPC           string // %s = raw input value
	ErrCannotConsumeTestMode string
	ErrNoStock              string
	ErrNoStockFor           string // %s = product name
	ErrSameLocation         string

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
	HeaderHints string // key-binding hint line in header

	// ── Input line hints ───────────────────────────────────────────────────────
	HintLookupView    string
	HintTransfer      string
	HintPriceHistory  string
	HintMealPlan      string
	HintTodayMealPlan string
	HintRecipeList    string
	HintRecipeDetail  string
	HintEditNotes     string
	HintYesNo          string
	HintUnknownBarcode string
	HintExpiringDetail string

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
	LabelUPC        string
	LabelProduct    string
	LabelInStock    string
	LabelShelfLife  string
	LabelLocation   string
	LabelExpires    string
	LabelStore      string
	LabelLastPrice  string
	LabelNotes      string
	LabelOFF        string // short label used in product-info panel
	LabelOFFName    string // label used in lookup-view panel
	LabelCategories string
	NewProductBadge string
	FoundInGrocy    string
	OpenFoodFacts   string // parenthetical source tag
	ShelfLifeNever  string
	FmtDays         string // %d = number of days
	FmtProductIDHint string // %d = product ID

	// ── Lookup view ───────────────────────────────────────────────────────────
	ProductOverview string

	// ── Edit views ────────────────────────────────────────────────────────────
	EditNameLabel string
	EditNameHint  string
	NotesLabel    string

	// ── Consume ───────────────────────────────────────────────────────────────
	ConsumingLabel string

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
	FmtLogAdd     string
	LogAddTo      string // %s = location name
	LogAddExpiry  string // %s = date string

	// ── Help overlay ──────────────────────────────────────────────────────────
	HelpTitle               string
	HelpBody                string // multi-line keyboard shortcut descriptions
	HelpUnknownBarcodeTitle string
	HelpUnknownBarcodeBody  string // multi-line unknown-barcode action descriptions
}

// English is the built-in English locale.
var English = Strings{
	// Input bar
	InputPlaceholder:  "Scan UPC or type command...",
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
	ErrNoStock:                "no stock on hand",
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
	HeaderHints: "q:quit m:mode /:search n:new ?:help",

	// Input line hints
	HintLookupView:     "n = notes  •  p = price history  •  t = transfer  •  Esc/Enter = dismiss",
	HintTransfer:       "Tab/↓ next field  •  Enter submit  •  Esc cancel",
	HintPriceHistory:   "↑/↓/j/k = navigate  •  Esc/p = back",
	HintMealPlan:       "↑/↓/j/k = scroll  •  r = refresh  •  Esc/q = back",
	HintTodayMealPlan:  "r = refresh  •  Esc/q = back",
	HintRecipeList:     "↑/↓/j/k = navigate  •  Enter/→ = view  •  r = refresh  •  Esc/q = back",
	HintRecipeDetail:   "↑/↓/j/k = scroll  •  ←/h/q/Esc = back",
	HintEditNotes:      "Enter to save  •  Esc to cancel",
	HintYesNo:          "y = yes, any other key = no",
	HintUnknownBarcode: "C/Enter = create new  •  L = link existing  •  Esc = cancel",
	HintExpiringDetail: "← / Esc = back  •  j/k = navigate",

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
	SearchNavHint: "↑/↓ navigate · Tab filter location · Enter select · n new · Esc cancel",

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
	ConsumingLabel: "Consuming:",

	// Transfer
	TransferStockHeader: "Transfer stock",

	// Shopping list prompt
	ConsumedLabel:      "Consumed:",
	StockNowZero:       "Stock is now zero.",
	ShoppingListPrompt: "Add to shopping list? [y/N]",

	// Unknown barcode
	UnknownBarcodePrompt: "Unknown barcode — what would you like to do?",
	UnknownBarcodeCreate: "  [C] / Enter  Create a new product",
	UnknownBarcodeLink:   "  [L]          Link to an existing product",

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
	HelpBody: "" +
		"  q       Quit\n" +
		"  m       Toggle Add/Consume mode\n" +
		"  /       Search products by name\n" +
		"  n       New product (no barcode)\n" +
		"  x       Export stock to CSV\n" +
		"  t       Today's meal plan (compact)\n" +
		"  P       Meal plan (next 7 days, scrollable)\n" +
		"  r       Recipe list (fulfillability), Enter/→ to view details\n" +
		"  e       Edit product name\n" +
		"  p       Price history (lookup view)\n" +
		"  t       Transfer stock between locations (lookup view)\n" +
		"  ?       Toggle this help\n" +
		"  ↑/k     Older UPC history entry (or expiry panel up when no history)\n" +
		"  ↓/j     Newer UPC history entry (or expiry panel down)\n" +
		"  c       Mark as consumed (expiring-soon panel)\n" +
		"  d       Mark as spoiled (expiring-soon panel)\n" +
		"  →/l     Open item details (expiring-soon panel)\n" +
		"  Tab/↓   Next field (form)\n" +
		"  S-Tab/↑ Previous field (form)\n" +
		"  Enter   Submit form\n" +
		"  Esc     Cancel current scan\n" +
		"  Ctrl+C  Quit\n\n",
	HelpUnknownBarcodeTitle: "Unknown Barcode",
	HelpUnknownBarcodeBody: "" +
		"  C/Enter  Create a new product for this barcode\n" +
		"  L        Link barcode to an existing product (rebrand/repackage)\n",
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
