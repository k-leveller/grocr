package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kevin/grocy-scanner/api"
	"github.com/kevin/grocy-scanner/logger"
	"github.com/kevin/grocy-scanner/scanner"
	"github.com/kevin/grocy-scanner/ui"
)

const (
	expPanelWidth    = 28
	expPanelMinWidth = 100
)

type AppState int

const (
	StateIdle AppState = iota
	StateLookup
	StateDisplay
	StateForm
	StateConsume
	StateSearch
	StateEditName
	StateLookupView
	StateShoppingListPrompt
	StatePriceHistory
	StateEditNotes
	StateTransfer
	StateMealPlan
)

type Model struct {
	state    AppState
	mode     string // "add", "consume", or "lookup"
	width    int
	height   int
	showHelp bool

	// Input bar
	input textinput.Model

	// API clients
	grocy    *api.GrocyClient
	off      *api.OFFClient
	defaults *api.Defaults
	testMode bool

	// Product cache for search
	allProducts []api.Product

	// Current scan context
	currentUPC     string
	currentProduct *api.Product
	offInfo        *api.OFFProduct
	isNewProduct   bool
	stockInfo      *api.StockInfo

	// Form
	form ui.Form

	// Search
	search ui.Search

	// Edit name
	editInput textinput.Model

	// Log
	logEntries []ui.LogEntry

	// Expiring soon panel
	expiringSoon       []api.ExpiringItem
	expiringSoonLoaded bool
	expPanelCursor     int

	// Price history panel
	priceHistory       []api.StockTransaction
	priceHistoryCursor int

	// Status message
	statusMsg string
	statusErr bool

	// Loading
	loading   bool
	lookupSeq int

	// UPC scan history for up/down navigation in idle input
	upcHistory  []string
	historyPos  int    // -1 = not navigating; 0 = most recent
	historySave string // input value saved when history nav begins

	// Meal plan
	mealPlan        []api.MealPlanItem
	mealPlanRecipes map[int]string
	mealPlanLoaded  bool
	mealPlanOffset  int
}

// Messages
type lookupResultMsg struct {
	seq     int
	product *api.Product
	offInfo *api.OFFProduct
	err     error
}

type actionResultMsg struct {
	entry       ui.LogEntry
	err         error
	zeroedStock bool
}

type shoppingListMsg struct {
	productName string
	err         error
}

type stockInfoMsg struct {
	productID int
	info      *api.StockInfo
}

type productsLoadedMsg struct {
	products []api.Product
}

type expiringSoonMsg struct {
	items []api.ExpiringItem
	err   error
}

type priceHistoryMsg struct {
	productID int
	items     []api.StockTransaction
	err       error
}

type defaultsLoadedMsg struct {
	defaults *api.Defaults
	err      error
}

type exportResultMsg struct {
	path string
	err  error
}

type mealPlanMsg struct {
	items   []api.MealPlanItem
	recipes map[int]string
	err     error
}

func NewModel(grocy *api.GrocyClient, off *api.OFFClient, testMode bool) Model {
	ti := textinput.New()
	ti.Placeholder = "Scan UPC or type command..."
	ti.Focus()
	ti.CharLimit = 20

	return Model{
		state:      StateIdle,
		mode:       "add",
		input:      ti,
		grocy:      grocy,
		off:        off,
		testMode:   testMode,
		historyPos: -1,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.loadDefaults(),
		m.loadProducts(),
		m.loadExpiringSoon(),
	)
}

func (m Model) loadExpiringSoon() tea.Cmd {
	return func() tea.Msg {
		if m.testMode {
			return expiringSoonMsg{}
		}
		items, err := m.grocy.GetExpiringSoon(7)
		return expiringSoonMsg{items: items, err: err}
	}
}

func (m Model) loadDefaults() tea.Cmd {
	return func() tea.Msg {
		if m.testMode {
			return defaultsLoadedMsg{defaults: &api.Defaults{
				LocationID: 1,
				QuID:       1,
				Locations: []api.Location{
					{ID: 1, Name: "Fridge"},
					{ID: 2, Name: "Freezer"},
					{ID: 3, Name: "Pantry"},
					{ID: 4, Name: "Bathroom"},
				},
				QuantityUnits: []api.QuantityUnit{
					{ID: 1, Name: "Piece"},
					{ID: 2, Name: "Oz"},
				},
			}}
		}
		d, err := m.grocy.GetDefaults()
		return defaultsLoadedMsg{defaults: d, err: err}
	}
}

func (m Model) loadProducts() tea.Cmd {
	return func() tea.Msg {
		if m.testMode {
			return productsLoadedMsg{products: []api.Product{
				{ID: 1, Name: "Test Product"},
			}}
		}
		products, _ := m.grocy.GetAllProducts()
		return productsLoadedMsg{products: products}
	}
}

func (m Model) lookupUPC(upc string, seq int) tea.Cmd {
	return func() tea.Msg {
		var product *api.Product
		var offInfo *api.OFFProduct

		if !m.testMode {
			p, err := m.grocy.FindProductByBarcode(upc)
			if err == nil {
				product = p
			}
		}

		off, _ := m.off.Lookup(upc)
		offInfo = off

		return lookupResultMsg{seq: seq, product: product, offInfo: offInfo}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case defaultsLoadedMsg:
		if msg.err != nil {
			logger.LogError("failed to load Grocy defaults: " + msg.err.Error())
			m.statusMsg = "Failed to load Grocy defaults: " + msg.err.Error()
			m.statusErr = true
		} else {
			m.defaults = msg.defaults
		}
		return m, nil

	case productsLoadedMsg:
		m.allProducts = msg.products
		return m, nil

	case expiringSoonMsg:
		m.expiringSoonLoaded = true
		if msg.err == nil {
			m.expiringSoon = msg.items
			if m.expPanelCursor >= len(m.expiringSoon) {
				m.expPanelCursor = max(len(m.expiringSoon)-1, 0)
			}
		}
		return m, nil

	case priceHistoryMsg:
		if m.currentProduct != nil && msg.productID == m.currentProduct.ID {
			if msg.err != nil {
				m.statusMsg = "Price history unavailable: " + msg.err.Error()
				m.statusErr = true
				m.state = StateLookupView
			} else {
				m.priceHistory = msg.items
				m.priceHistoryCursor = 0
			}
		}
		return m, nil

	case lookupResultMsg:
		if msg.seq != m.lookupSeq {
			return m, nil // stale result from a cancelled lookup
		}
		m.loading = false
		if msg.err != nil {
			logger.LogError("lookup error: " + msg.err.Error())
			m.statusMsg = "Lookup error: " + msg.err.Error()
			m.statusErr = true
			m.state = StateIdle
			return m, m.input.Focus()
		}
		return m.handleLookupResult(msg)

	case stockInfoMsg:
		if m.currentProduct != nil && msg.productID == m.currentProduct.ID {
			m.stockInfo = msg.info
			if m.state == StateDisplay {
				m.applyPriceDefault()
			}
		}
		return m, nil

	case actionResultMsg:
		m.loading = false
		if msg.err != nil {
			logger.LogError("action error: " + msg.err.Error())
			m.statusMsg = "Error: " + msg.err.Error()
			m.statusErr = true
		} else {
			m.logEntries = append([]ui.LogEntry{msg.entry}, m.logEntries...)
			switch msg.entry.Action {
			case "consume":
				logger.LogConsume(msg.entry.ProductName, msg.entry.Quantity)
				m.statusMsg = ""
			case "spoiled":
				logger.LogConsume(msg.entry.ProductName, msg.entry.Quantity)
				m.statusMsg = "Spoiled: " + msg.entry.ProductName
				m.statusErr = false
			case "transfer":
				logger.LogTransfer(msg.entry.ProductName, msg.entry.Quantity, msg.entry.FromLocation, msg.entry.ToLocation)
				m.statusMsg = ""
			default:
				logger.LogAdd(msg.entry.ProductName, msg.entry.Quantity, msg.entry.Location, msg.entry.Expiry)
				m.statusMsg = ""
			}
		}
		if msg.err == nil && msg.zeroedStock && m.currentProduct != nil {
			m.state = StateShoppingListPrompt
			return m, nil
		}
		m.state = StateIdle
		m.currentProduct = nil
		m.offInfo = nil
		m.stockInfo = nil
		m.input.SetValue("")
		cmds := []tea.Cmd{m.input.Focus()}
		if msg.err == nil {
			cmds = append(cmds, m.loadExpiringSoon())
		}
		return m, tea.Batch(cmds...)

	case shoppingListMsg:
		if msg.err != nil {
			logger.LogError("shopping list: " + msg.err.Error())
			m.statusMsg = "Shopping list error: " + msg.err.Error()
			m.statusErr = true
		} else {
			logger.LogShoppingList(msg.productName)
			m.statusMsg = msg.productName + " added to shopping list"
			m.statusErr = false
		}
		return m, nil

	case exportResultMsg:
		m.loading = false
		if msg.err != nil {
			logger.LogError("export: " + msg.err.Error())
			m.statusMsg = "Export failed: " + msg.err.Error()
			m.statusErr = true
		} else {
			m.statusMsg = "Exported to " + msg.path
			m.statusErr = false
		}
		return m, nil

	case mealPlanMsg:
		m.mealPlanLoaded = true
		if msg.err != nil {
			if m.state == StateMealPlan {
				m.statusMsg = "Meal plan unavailable: " + msg.err.Error()
				m.statusErr = true
			}
		} else {
			m.mealPlan = msg.items
			m.mealPlanRecipes = msg.recipes
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Update sub-components based on state
	switch m.state {
	case StateIdle:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	case StateForm:
		cmd := m.form.Update(msg)
		return m, cmd
	case StateSearch:
		cmd := m.search.Update(msg)
		return m, cmd
	case StateEditName, StateEditNotes:
		var cmd tea.Cmd
		m.editInput, cmd = m.editInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keys
	switch key {
	case "ctrl+c":
		return m, tea.Quit
	}

	// Help toggle works in any state
	if key == "?" && m.state == StateIdle {
		m.showHelp = !m.showHelp
		return m, nil
	}
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	switch m.state {
	case StateIdle:
		return m.handleIdleKey(msg)
	case StateForm:
		return m.handleFormKey(msg)
	case StateDisplay:
		return m.handleDisplayKey(msg)
	case StateConsume:
		return m.handleConsumeKey(msg)
	case StateSearch:
		return m.handleSearchKey(msg)
	case StateEditName:
		return m.handleEditNameKey(msg)
	case StateLookupView:
		return m.handleLookupViewKey(msg)
	case StatePriceHistory:
		return m.handlePriceHistoryKey(msg)
	case StateShoppingListPrompt:
		return m.handleShoppingListKey(msg)
	case StateEditNotes:
		return m.handleEditNotesKey(msg)
	case StateTransfer:
		return m.handleTransferFormKey(msg)
	case StateMealPlan:
		return m.handleMealPlanKey(msg)
	}

	return m, nil
}

func (m Model) handleIdleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		if m.loading {
			m.loading = false
			m.lookupSeq++ // invalidate the in-flight lookup
			m.currentUPC = ""
			m.statusMsg = "Cancelled"
			m.statusErr = false
			return m, m.input.Focus()
		}
	case "q":
		if m.input.Value() == "" {
			return m, tea.Quit
		}
	case "j":
		if m.input.Value() == "" && len(m.expiringSoon) > 0 {
			m.expPanelCursor = min(m.expPanelCursor+1, len(m.expiringSoon)-1)
			return m, nil
		}
	case "k":
		if m.input.Value() == "" && len(m.expiringSoon) > 0 {
			m.expPanelCursor = max(m.expPanelCursor-1, 0)
			return m, nil
		}
	case "up":
		if len(m.upcHistory) > 0 {
			next := m.historyPos + 1
			if next < len(m.upcHistory) {
				if m.historyPos < 0 {
					m.historySave = m.input.Value()
				}
				m.historyPos = next
				m.input.SetValue(m.upcHistory[next])
				m.input.CursorEnd()
				return m, nil
			}
		}
	case "down":
		if m.historyPos >= 0 {
			next := m.historyPos - 1
			if next < 0 {
				m.input.SetValue(m.historySave)
				m.input.CursorEnd()
				m.historyPos = -1
			} else {
				m.historyPos = next
				m.input.SetValue(m.upcHistory[next])
				m.input.CursorEnd()
			}
			return m, nil
		}
	case "d":
		if !m.loading && m.input.Value() == "" && m.expPanelCursor >= 0 && m.expPanelCursor < len(m.expiringSoon) {
			item := m.expiringSoon[m.expPanelCursor]
			var product *api.Product
			for i := range m.allProducts {
				if m.allProducts[i].ID == item.ProductID {
					product = &m.allProducts[i]
					break
				}
			}
			m.currentProduct = product
			m.loading = true
			m.statusMsg = ""
			return m, m.consumeFromPanel(item, product)
		}
	case "m":
		if m.input.Value() == "" {
			switch m.mode {
			case "add":
				m.mode = "consume"
			case "consume":
				m.mode = "lookup"
			default:
				m.mode = "add"
			}
			return m, nil
		}
	case "/":
		if m.input.Value() == "" {
			m.state = StateSearch
			var locs []api.Location
			if m.defaults != nil {
				locs = m.defaults.Locations
			}
			m.search = ui.NewSearch(m.allProducts, locs)
			m.input.Blur()
			return m, nil
		}
	case "ctrl+e":
		if m.input.Value() == "" {
			m.loading = true
			m.statusMsg = ""
			return m, m.doExport()
		}
	case "ctrl+n":
		m.input.SetValue("")
		m.historyPos = -1
		return m.startManualProductEntry()
	case "P":
		if m.input.Value() == "" {
			m.state = StateMealPlan
			m.mealPlanLoaded = false
			m.mealPlan = nil
			m.mealPlanRecipes = nil
			m.mealPlanOffset = 0
			m.input.Blur()
			return m, m.loadMealPlan()
		}
		return m, nil
	case "enter":
		val := strings.TrimSpace(m.input.Value())
		if val == "" {
			return m, nil
		}
		upc := scanner.CleanUPC(val)
		if upc == "" {
			m.statusMsg = "Invalid UPC: " + val
			m.statusErr = true
			m.input.SetValue("")
			m.historyPos = -1
			return m, nil
		}
		m.currentUPC = upc
		m.upcHistory = prependUPCHistory(m.upcHistory, upc)
		m.historyPos = -1
		m.loading = true
		m.statusMsg = ""
		m.lookupSeq++
		m.input.Blur()
		return m, m.lookupUPC(upc, m.lookupSeq)
	}

	m.historyPos = -1
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) handleLookupResult(msg lookupResultMsg) (tea.Model, tea.Cmd) {
	m.currentProduct = msg.product
	m.offInfo = msg.offInfo
	m.isNewProduct = msg.product == nil

	if m.mode == "consume" {
		if msg.product == nil {
			m.statusMsg = "Product not found in Grocy — cannot consume"
			m.statusErr = true
			m.state = StateIdle
			return m, m.input.Focus()
		}
		// Load stock info
		m.state = StateConsume
		return m, m.loadStockForConsume()
	}

	if m.mode == "lookup" {
		if msg.product == nil {
			m.statusMsg = "Product not found in Grocy"
			m.statusErr = true
			m.state = StateIdle
			return m, m.input.Focus()
		}
		m.state = StateLookupView
		m.stockInfo = nil
		return m, m.loadStock()
	}

	if m.isNewProduct {
		m.state = StateForm
		m.form = m.buildNewProductForm()
		return m, nil
	}

	m.state = StateDisplay
	m.form = m.buildExistingProductForm()
	m.stockInfo = nil
	return m, m.loadStock()
}

func (m Model) startManualProductEntry() (tea.Model, tea.Cmd) {
	m.lookupSeq++
	m.loading = false
	m.isNewProduct = true
	m.currentProduct = nil
	m.currentUPC = ""
	m.offInfo = nil
	m.statusMsg = ""
	m.statusErr = false
	m.input.Blur()
	m.state = StateForm
	m.form = m.buildNewProductForm()
	return m, nil
}

func (m Model) loadStock() tea.Cmd {
	productID := m.currentProduct.ID
	if m.testMode {
		return func() tea.Msg {
			return stockInfoMsg{productID: productID, info: &api.StockInfo{StockAmount: 3}}
		}
	}
	return func() tea.Msg {
		info, err := m.grocy.GetStock(productID)
		if err != nil {
			return stockInfoMsg{productID: productID}
		}
		return stockInfoMsg{productID: productID, info: info}
	}
}

func (m Model) loadMealPlan() tea.Cmd {
	return func() tea.Msg {
		if m.testMode {
			return mealPlanMsg{items: []api.MealPlanItem{}, recipes: map[int]string{}}
		}
		today := time.Now().Format("2006-01-02")
		end := time.Now().AddDate(0, 0, 6).Format("2006-01-02")
		items, err := m.grocy.GetMealPlan(today, end)
		if err != nil {
			return mealPlanMsg{err: err}
		}
		recipes, _ := m.grocy.GetRecipes()
		recipeMap := make(map[int]string, len(recipes))
		for _, r := range recipes {
			recipeMap[r.ID] = r.Name
		}
		return mealPlanMsg{items: items, recipes: recipeMap}
	}
}

func (m Model) loadPriceHistory() tea.Cmd {
	productID := m.currentProduct.ID
	if m.testMode {
		return func() tea.Msg {
			return priceHistoryMsg{productID: productID, items: []api.StockTransaction{}}
		}
	}
	return func() tea.Msg {
		txns, err := m.grocy.GetPurchaseHistory(productID, 20)
		return priceHistoryMsg{productID: productID, items: txns, err: err}
	}
}

func (m Model) loadStockForConsume() tea.Cmd {
	return func() tea.Msg {
		if m.testMode || m.currentProduct == nil {
			return actionResultMsg{err: fmt.Errorf("cannot consume in test mode")}
		}
		info, err := m.grocy.GetStock(m.currentProduct.ID)
		if err != nil {
			return actionResultMsg{err: err}
		}
		// We store it and let the view handle prompting for quantity
		// For now, consume 1
		if info.StockAmount <= 0 {
			return actionResultMsg{err: fmt.Errorf("no stock on hand")}
		}
		err = m.grocy.ConsumeStock(m.currentProduct.ID, 1, false)
		entry := ui.LogEntry{
			ProductName: m.currentProduct.Name,
			Quantity:    1,
			Action:      "consume",
			Success:     err == nil,
			Time:        time.Now(),
		}
		zeroedStock := err == nil && info.StockAmount-1 <= 0
		return actionResultMsg{entry: entry, err: err, zeroedStock: zeroedStock}
	}
}

func (m Model) consumeFromPanel(item api.ExpiringItem, product *api.Product) tea.Cmd {
	return func() tea.Msg {
		if m.testMode {
			return actionResultMsg{err: fmt.Errorf("cannot consume in test mode")}
		}
		info, err := m.grocy.GetStock(item.ProductID)
		if err != nil {
			return actionResultMsg{err: err}
		}
		if info.StockAmount <= 0 {
			return actionResultMsg{err: fmt.Errorf("no stock on hand for %s", item.ProductName)}
		}
		err = m.grocy.ConsumeStock(item.ProductID, 1, true)
		entry := ui.LogEntry{
			ProductName: item.ProductName,
			Quantity:    1,
			Action:      "spoiled",
			Success:     err == nil,
			Time:        time.Now(),
		}
		zeroedStock := err == nil && info.StockAmount-1 <= 0
		return actionResultMsg{entry: entry, err: err, zeroedStock: zeroedStock}
	}
}

func (m Model) buildNewProductForm() ui.Form {
	// Determine defaults
	defaultName := ""
	if m.offInfo != nil && m.offInfo.Name != "" {
		defaultName = m.offInfo.Name
	}

	shelfDays := 0
	shelfSource := ""
	if m.offInfo != nil && m.offInfo.ShelfLifeDays != nil {
		shelfDays = *m.offInfo.ShelfLifeDays
		shelfSource = "OFF"
	}
	if shelfDays == 0 && defaultName != "" {
		if d := api.EstimateShelfLife(defaultName); d > 0 {
			shelfDays = d
			shelfSource = "name"
		}
	}
	if shelfDays == 0 && m.offInfo != nil && m.offInfo.Categories != "" {
		if d := api.EstimateShelfLife(m.offInfo.Categories); d > 0 {
			shelfDays = d
			shelfSource = "category"
		}
	}

	expiryDefault := ""
	expiryHint := ""
	if shelfDays > 0 {
		expDate := time.Now().AddDate(0, 0, shelfDays).Format("2006-01-02")
		expiryDefault = expDate
		expiryHint = fmt.Sprintf("~%dd from %s", shelfDays, shelfSource)
	}

	qtyUnitDefault := ""
	qtyUnitHint := ""
	if m.defaults != nil && len(m.defaults.QuantityUnits) > 0 {
		var parts []string
		for i, u := range m.defaults.QuantityUnits {
			parts = append(parts, fmt.Sprintf("%d)%s", i+1, u.Name))
			if u.ID == m.defaults.QuID && qtyUnitDefault == "" {
				qtyUnitDefault = u.Name
			}
		}
		if qtyUnitDefault == "" {
			qtyUnitDefault = m.defaults.QuantityUnits[0].Name
		}
		qtyUnitHint = strings.Join(parts, " ")
	}

	locationDefault := ""
	locationHint := ""
	if m.defaults != nil && len(m.defaults.Locations) > 0 {
		var parts []string
		for i, loc := range m.defaults.Locations {
			parts = append(parts, fmt.Sprintf("%d)%s", i+1, loc.Name))
			if loc.ID == m.defaults.LocationID && locationDefault == "" {
				locationDefault = loc.Name
			}
		}
		if locationDefault == "" {
			locationDefault = m.defaults.Locations[0].Name
		}
		locationHint = strings.Join(parts, " ")
	}

	storeDefault := ""
	storeHint := ""
	if m.defaults != nil && len(m.defaults.Stores) > 0 {
		var parts []string
		for i, s := range m.defaults.Stores {
			parts = append(parts, fmt.Sprintf("%d)%s", i+1, s.Name))
		}
		storeHint = strings.Join(parts, " ")
	}

	fields := []ui.FormField{
		{Label: "Name", Default: defaultName, Required: true},
		{Label: "Short name", Default: ""},
		{Label: "Qty unit", Default: qtyUnitDefault, Hint: qtyUnitHint, Required: true},
		{Label: "Expires", Default: expiryDefault, Hint: expiryHint + " (YYYY-MM-DD, days, or blank=never)"},
		{Label: "Location", Default: locationDefault, Hint: locationHint},
		{Label: "Store", Default: storeDefault, Hint: storeHint},
		{Label: "Quantity", Default: "1"},
		{Label: "Price", Default: "", Hint: "total price paid"},
	}

	return ui.NewForm(fields)
}

func (m Model) buildExistingProductForm() ui.Form {
	shelfDays := m.currentProduct.DefaultBestBeforeDays
	expiryDefault := ""
	expiryHint := ""
	if shelfDays > 0 {
		expDate := time.Now().AddDate(0, 0, shelfDays).Format("2006-01-02")
		expiryDefault = expDate
		expiryHint = fmt.Sprintf("~%dd from product", shelfDays)
	}

	locationDefault := ""
	locationHint := ""
	if m.defaults != nil && len(m.defaults.Locations) > 0 {
		var parts []string
		for i, loc := range m.defaults.Locations {
			parts = append(parts, fmt.Sprintf("%d)%s", i+1, loc.Name))
		}
		// Default to the product's own location name
		for _, loc := range m.defaults.Locations {
			if loc.ID == m.currentProduct.LocationID {
				locationDefault = loc.Name
				break
			}
		}
		if locationDefault == "" {
			locationDefault = m.defaults.Locations[0].Name
		}
		locationHint = strings.Join(parts, " ")
	}

	storeDefault := ""
	storeHint := ""
	if m.defaults != nil && len(m.defaults.Stores) > 0 {
		var parts []string
		for i, s := range m.defaults.Stores {
			parts = append(parts, fmt.Sprintf("%d)%s", i+1, s.Name))
		}
		// Default to the product's own store name
		for _, s := range m.defaults.Stores {
			if s.ID == m.currentProduct.ShoppingLocationID {
				storeDefault = s.Name
				break
			}
		}
		storeHint = strings.Join(parts, " ")
	}

	qtyUnitHint := ""
	if m.defaults != nil {
		for _, u := range m.defaults.QuantityUnits {
			if u.ID == m.currentProduct.QuIDStock {
				qtyUnitHint = u.Name
				break
			}
		}
	}

	fields := []ui.FormField{
		{Label: "Expires", Default: expiryDefault, Hint: expiryHint + " (YYYY-MM-DD, days, or blank=never)"},
		{Label: "Location", Default: locationDefault, Hint: locationHint},
		{Label: "Store", Default: storeDefault, Hint: storeHint},
		{Label: "Quantity", Default: "1", Hint: qtyUnitHint},
		{Label: "Price", Default: "", Hint: "total price paid"},
	}

	return ui.NewForm(fields)
}

func (m Model) buildTransferForm() ui.Form {
	locationHint := ""
	fromDefault := ""
	toDefault := ""

	if m.defaults != nil && len(m.defaults.Locations) > 0 {
		var parts []string
		for i, loc := range m.defaults.Locations {
			parts = append(parts, fmt.Sprintf("%d)%s", i+1, loc.Name))
		}
		locationHint = strings.Join(parts, " ")

		for _, loc := range m.defaults.Locations {
			if loc.ID == m.currentProduct.LocationID {
				fromDefault = loc.Name
				break
			}
		}
		if fromDefault == "" {
			fromDefault = m.defaults.Locations[0].Name
		}
		for _, loc := range m.defaults.Locations {
			if loc.Name != fromDefault {
				toDefault = loc.Name
				break
			}
		}
	}

	fields := []ui.FormField{
		{Label: "Quantity", Default: "1"},
		{Label: "From", Default: fromDefault, Hint: locationHint},
		{Label: "To", Default: toDefault, Hint: locationHint},
	}

	return ui.NewForm(fields)
}

func (m Model) handleTransferFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.loading {
		return m, nil
	}

	cmd := m.form.Update(msg)

	if m.form.Cancelled {
		m.state = StateLookupView
		return m, nil
	}

	if m.form.Submitted {
		m.loading = true
		return m, m.submitTransfer()
	}

	return m, cmd
}

func (m Model) submitTransfer() tea.Cmd {
	product := m.currentProduct
	return func() tea.Msg {
		quantity := m.parseQuantity(m.form.Value(0))
		fromID := m.resolveLocation(m.form.Value(1))
		toID := m.resolveLocation(m.form.Value(2))

		if fromID == toID {
			return actionResultMsg{err: fmt.Errorf("from and to locations must differ")}
		}

		fromName := m.locationName(fromID)
		toName := m.locationName(toID)

		entry := ui.LogEntry{
			ProductName:  product.Name,
			Quantity:     quantity,
			Location:     fromName + " → " + toName,
			FromLocation: fromName,
			ToLocation:   toName,
			Action:       "transfer",
			Time:         time.Now(),
		}

		if m.testMode {
			entry.Success = true
			return actionResultMsg{entry: entry}
		}

		err := m.grocy.TransferStock(product.ID, quantity, fromID, toID)
		entry.Success = err == nil
		return actionResultMsg{entry: entry, err: err}
	}
}

func (m Model) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	cmd := m.form.Update(msg)

	if m.form.Cancelled {
		m.state = StateIdle
		m.currentProduct = nil
		m.offInfo = nil
		m.stockInfo = nil
		m.input.SetValue("")
		return m, m.input.Focus()
	}

	if m.form.Submitted {
		return m, m.submitForm()
	}

	return m, cmd
}

func (m Model) handleDisplayKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if key == "e" {
		// Enter edit name mode
		m.state = StateEditName
		ti := textinput.New()
		ti.SetValue(m.currentProduct.Name)
		ti.Focus()
		ti.CharLimit = 100
		m.editInput = ti
		return m, nil
	}

	// Otherwise delegate to form
	cmd := m.form.Update(msg)

	if m.form.Cancelled {
		m.state = StateIdle
		m.currentProduct = nil
		m.offInfo = nil
		m.stockInfo = nil
		m.input.SetValue("")
		return m, m.input.Focus()
	}

	if m.form.Submitted {
		return m, m.submitForm()
	}

	return m, cmd
}

func (m Model) handleConsumeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if key == "esc" {
		m.state = StateIdle
		m.input.SetValue("")
		return m, m.input.Focus()
	}
	return m, nil
}

func (m Model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+n" {
		return m.startManualProductEntry()
	}

	cmd := m.search.Update(msg)

	if m.search.Cancelled {
		m.state = StateIdle
		m.input.SetValue("")
		return m, m.input.Focus()
	}

	if m.search.Selected != nil {
		// Treat selected product as if it was found by barcode
		m.currentProduct = m.search.Selected
		m.currentUPC = ""
		m.isNewProduct = false
		m.offInfo = nil

		if m.mode == "consume" {
			m.state = StateConsume
			return m, m.loadStockForConsume()
		}

		if m.mode == "lookup" {
			m.state = StateLookupView
			m.stockInfo = nil
			return m, m.loadStock()
		}

		m.state = StateDisplay
		m.form = m.buildExistingProductForm()
		m.stockInfo = nil
		return m, m.loadStock()
	}

	return m, cmd
}

func (m *Model) applyPriceDefault() {
	const priceFieldIdx = 4
	if m.stockInfo == nil || m.stockInfo.LastPrice <= 0 {
		return
	}
	if priceFieldIdx >= len(m.form.Fields) {
		return
	}
	if m.form.Fields[priceFieldIdx].Input.Value() != "" || m.form.Fields[priceFieldIdx].Default != "" {
		return
	}
	priceStr := fmt.Sprintf("%.2f", m.stockInfo.LastPrice)
	m.form.Fields[priceFieldIdx].Default = priceStr
	m.form.Fields[priceFieldIdx].Input.Placeholder = priceStr
}

func (m Model) handleEditNameKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "enter":
		newName := strings.TrimSpace(m.editInput.Value())
		if newName != "" && newName != m.currentProduct.Name {
			oldName := m.currentProduct.Name
			if !m.testMode {
				if err := m.grocy.UpdateProductName(m.currentProduct.ID, newName); err != nil {
					logger.LogError("update product name: " + err.Error())
					m.statusMsg = "Error updating name: " + err.Error()
					m.statusErr = true
					m.state = StateDisplay
					m.applyPriceDefault()
					return m, nil
				}
			}
			m.currentProduct.Name = newName
			logger.LogEditName(newName, oldName)
			m.statusMsg = "Name updated"
			m.statusErr = false
		}
		m.state = StateDisplay
		m.applyPriceDefault()
		return m, nil
	case "esc":
		m.state = StateDisplay
		m.applyPriceDefault()
		return m, nil
	}

	var cmd tea.Cmd
	m.editInput, cmd = m.editInput.Update(msg)
	return m, cmd
}

func (m Model) handleLookupViewKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter":
		m.state = StateIdle
		m.currentProduct = nil
		m.offInfo = nil
		m.stockInfo = nil
		m.input.SetValue("")
		return m, m.input.Focus()
	case "p":
		if m.currentProduct != nil {
			m.state = StatePriceHistory
			m.priceHistory = nil
			m.priceHistoryCursor = 0
			return m, m.loadPriceHistory()
		}
	case "n":
		if m.currentProduct != nil {
			m.state = StateEditNotes
			ti := textinput.New()
			ti.SetValue(m.currentProduct.Description)
			ti.Focus()
			ti.CharLimit = 500
			ti.Width = 60
			m.editInput = ti
		}
	case "t":
		if m.currentProduct != nil {
			m.state = StateTransfer
			m.form = m.buildTransferForm()
		}
	}
	return m, nil
}

func (m Model) handleEditNotesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "enter":
		newNotes := strings.TrimSpace(m.editInput.Value())
		if newNotes != m.currentProduct.Description {
			if !m.testMode {
				if err := m.grocy.UpdateProductDescription(m.currentProduct.ID, newNotes); err != nil {
					logger.LogError("update product notes: " + err.Error())
					m.statusMsg = "Error updating notes: " + err.Error()
					m.statusErr = true
					m.state = StateLookupView
					return m, nil
				}
			}
			m.currentProduct.Description = newNotes
			logger.LogEditNotes(m.currentProduct.Name)
			m.statusMsg = "Notes updated"
			m.statusErr = false
		}
		m.state = StateLookupView
		return m, nil
	case "esc":
		m.state = StateLookupView
		return m, nil
	}

	var cmd tea.Cmd
	m.editInput, cmd = m.editInput.Update(msg)
	return m, cmd
}

func (m Model) handlePriceHistoryKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter", "p":
		m.state = StateLookupView
		return m, nil
	case "j", "down":
		if m.priceHistoryCursor < len(m.priceHistory)-1 {
			m.priceHistoryCursor++
		}
	case "k", "up":
		if m.priceHistoryCursor > 0 {
			m.priceHistoryCursor--
		}
	}
	return m, nil
}

func (m Model) handleMealPlanKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.state = StateIdle
		m.statusMsg = ""
		m.statusErr = false
		return m, m.input.Focus()
	case "j", "down":
		// Cap at ~3 lines per item (day header + entry + blank gap)
		if m.mealPlanOffset < len(m.mealPlan)*3+len(m.mealPlanRecipes) {
			m.mealPlanOffset++
		}
	case "k", "up":
		if m.mealPlanOffset > 0 {
			m.mealPlanOffset--
		}
	case "r":
		m.mealPlanLoaded = false
		m.mealPlan = nil
		m.mealPlanRecipes = nil
		m.mealPlanOffset = 0
		return m, m.loadMealPlan()
	}
	return m, nil
}

func (m Model) handleShoppingListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var shoppingCmd tea.Cmd
	if msg.String() == "y" || msg.String() == "Y" {
		shoppingCmd = m.addToShoppingList()
	}
	m.state = StateIdle
	m.currentProduct = nil
	m.offInfo = nil
	m.stockInfo = nil
	m.input.SetValue("")
	cmds := []tea.Cmd{m.input.Focus(), m.loadExpiringSoon()}
	if shoppingCmd != nil {
		cmds = append(cmds, shoppingCmd)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) addToShoppingList() tea.Cmd {
	product := m.currentProduct
	return func() tea.Msg {
		err := m.grocy.AddToShoppingList(product.ID)
		return shoppingListMsg{productName: product.Name, err: err}
	}
}

func (m Model) doExport() tea.Cmd {
	return func() tea.Msg {
		var entries []api.StockEntry
		if m.testMode {
			entries = []api.StockEntry{
				{ProductName: "Test Product", Amount: 1, LocationID: 1, BestBeforeDate: "2026-12-31"},
			}
		} else {
			var err error
			entries, err = m.grocy.GetStockSnapshot()
			if err != nil {
				return exportResultMsg{err: err}
			}
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return exportResultMsg{err: fmt.Errorf("home dir: %w", err)}
		}
		path := fmt.Sprintf("%s/grocy-export-%s.csv", home, time.Now().Format("20060102"))

		f, err := os.Create(path)
		if err != nil {
			return exportResultMsg{err: err}
		}
		defer f.Close()

		w := csv.NewWriter(f)
		if err := w.Write([]string{"name", "amount", "location", "best_before"}); err != nil {
			return exportResultMsg{err: err}
		}
		for _, e := range entries {
			locName := ""
			if m.defaults != nil {
				for _, loc := range m.defaults.Locations {
					if loc.ID == e.LocationID {
						locName = loc.Name
						break
					}
				}
			}
			bb := e.BestBeforeDate
			if bb == "2999-12-31" {
				bb = ""
			}
			if err := w.Write([]string{e.ProductName, strconv.FormatFloat(e.Amount, 'f', -1, 64), locName, bb}); err != nil {
				return exportResultMsg{err: err}
			}
		}
		w.Flush()
		if err := w.Error(); err != nil {
			return exportResultMsg{err: err}
		}

		return exportResultMsg{path: path}
	}
}

func (m Model) submitForm() tea.Cmd {
	return func() tea.Msg {
		var productName string
		var quantity float64
		var price float64
		var bestBefore string
		var locationID int

		var storeID int
		var storeErr error
		var quID int
		if m.isNewProduct {
			productName = m.form.Value(0)
			// shortName = m.form.Value(1)
			quID = m.resolveQuantityUnit(m.form.Value(2))
			bestBefore = m.form.Value(3)
			locationID = m.resolveLocation(m.form.Value(4))
			storeID, storeErr = m.resolveOrCreateStore(m.form.Value(5))
			quantity = m.parseQuantity(m.form.Value(6))
			price = m.parsePrice(m.form.Value(7))
		} else {
			productName = m.currentProduct.Name
			bestBefore = m.form.Value(0)
			locationID = m.resolveLocation(m.form.Value(1))
			storeID, storeErr = m.resolveOrCreateStore(m.form.Value(2))
			quantity = m.parseQuantity(m.form.Value(3))
			price = m.parsePrice(m.form.Value(4))
		}
		if storeErr != nil {
			return actionResultMsg{err: storeErr}
		}

		bestBefore = resolveExpiry(bestBefore)

		if m.testMode {
			locName := m.locationName(locationID)
			entry := ui.LogEntry{
				ProductName: productName,
				Quantity:    quantity,
				Location:    locName,
				Expiry:      bestBefore,
				Action:      "add",
				Success:     true,
				Time:        time.Now(),
			}
			return actionResultMsg{entry: entry}
		}

		// Create product if new
		var product *api.Product
		if m.isNewProduct {
			shortName := m.form.Value(1)
			shelfDays := m.daysFromExpiry(bestBefore)
			freezeDays := 365
			p, err := m.grocy.CreateProduct(productName, shelfDays, m.defaults, shortName, locationID, storeID, quID, &freezeDays, shelfDays)
			if err != nil {
				return actionResultMsg{err: fmt.Errorf("create product: %w", err)}
			}
			product = p
			if m.currentUPC != "" {
				m.grocy.AddBarcode(product.ID, m.currentUPC)
			}
		} else {
			product = m.currentProduct
			// Update product's default location if it changed
			if locationID != product.LocationID {
				m.grocy.UpdateProductLocation(product.ID, locationID)
			}
			// Update product's default store if it changed
			if storeID != product.ShoppingLocationID {
				m.grocy.UpdateProductStore(product.ID, storeID)
			}
		}

		err := m.grocy.AddStock(product.ID, quantity, price, bestBefore, locationID)
		locName := m.locationName(locationID)
		entry := ui.LogEntry{
			ProductName: productName,
			Quantity:    quantity,
			Location:    locName,
			Expiry:      bestBefore,
			Action:      "add",
			Success:     err == nil,
			Time:        time.Now(),
		}
		return actionResultMsg{entry: entry, err: err}
	}
}

func (m Model) resolveQuantityUnit(val string) int {
	if m.defaults == nil || len(m.defaults.QuantityUnits) == 0 {
		return 1
	}

	if n, err := strconv.Atoi(val); err == nil && n >= 1 && n <= len(m.defaults.QuantityUnits) {
		return m.defaults.QuantityUnits[n-1].ID
	}

	lower := strings.ToLower(val)
	for _, u := range m.defaults.QuantityUnits {
		if strings.HasPrefix(strings.ToLower(u.Name), lower) {
			return u.ID
		}
	}

	return m.defaults.QuID
}

func (m Model) resolveLocation(val string) int {
	if m.defaults == nil || len(m.defaults.Locations) == 0 {
		return 1
	}

	// Try as number (1-indexed)
	if n, err := strconv.Atoi(val); err == nil && n >= 1 && n <= len(m.defaults.Locations) {
		return m.defaults.Locations[n-1].ID
	}

	// Try name prefix match
	lower := strings.ToLower(val)
	for _, loc := range m.defaults.Locations {
		if strings.HasPrefix(strings.ToLower(loc.Name), lower) {
			return loc.ID
		}
	}

	return m.defaults.LocationID
}

func (m Model) resolveOrCreateStore(val string) (int, error) {
	if val == "" {
		return 0, nil
	}

	if m.defaults != nil && len(m.defaults.Stores) > 0 {
		// Try as number (1-indexed)
		if n, err := strconv.Atoi(val); err == nil && n >= 1 && n <= len(m.defaults.Stores) {
			return m.defaults.Stores[n-1].ID, nil
		}

		// Try name prefix match
		lower := strings.ToLower(val)
		for _, s := range m.defaults.Stores {
			if strings.HasPrefix(strings.ToLower(s.Name), lower) {
				return s.ID, nil
			}
		}
	}

	// No match — create a new store
	if m.testMode {
		return 0, nil
	}
	store, err := m.grocy.CreateStore(val)
	if err != nil {
		return 0, fmt.Errorf("create store: %w", err)
	}
	return store.ID, nil
}

func (m Model) locationName(id int) string {
	if m.defaults == nil {
		return ""
	}
	for _, loc := range m.defaults.Locations {
		if loc.ID == id {
			return loc.Name
		}
	}
	return ""
}

func (m Model) parseQuantity(val string) float64 {
	n, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
	if err != nil || n <= 0 {
		return 1
	}
	return n
}

func (m Model) parsePrice(val string) float64 {
	val = strings.TrimSpace(strings.TrimPrefix(val, "$"))
	if val == "" {
		return 0
	}
	result, ok := evalArith(val)
	if !ok || result < 0 || math.IsInf(result, 0) || math.IsNaN(result) {
		return 0
	}
	return result
}

// evalArith evaluates a simple arithmetic expression supporting +, -, *, /.
// Operator precedence: * and / before + and -.
func evalArith(expr string) (float64, bool) {
	tokens, ops := splitAdditive(expr)
	sum := 0.0
	for i, tok := range tokens {
		val, ok := evalMultiplicative(tok)
		if !ok {
			return 0, false
		}
		if i == 0 {
			sum = val
		} else {
			switch ops[i-1] {
			case '+':
				sum += val
			case '-':
				sum -= val
			}
		}
	}
	return sum, true
}

// splitAdditive splits expr on top-level '+' and '-' operators.
// It skips signs that are part of a scientific-notation exponent (e.g. "1e-3")
// and skips signs that immediately follow another operator (e.g. "6+-1.5" keeps
// "-1.5" as the second term rather than splitting again).
func splitAdditive(expr string) ([]string, []rune) {
	var terms []string
	var ops []rune
	start := 0
	for i, ch := range expr {
		if (ch == '+' || ch == '-') && i > 0 {
			prev := expr[i-1]
			if prev == 'e' || prev == 'E' {
				continue // exponent sign in scientific notation, not an operator
			}
			term := strings.TrimSpace(expr[start:i])
			if term == "" {
				continue // sign immediately follows another operator; treat as unary
			}
			terms = append(terms, term)
			ops = append(ops, ch)
			start = i + 1
		}
	}
	terms = append(terms, strings.TrimSpace(expr[start:]))
	return terms, ops
}

// evalMultiplicative evaluates a sequence of factors joined by * or /.
func evalMultiplicative(expr string) (float64, bool) {
	tokens, ops := splitMultiplicative(expr)
	result := 0.0
	for i, tok := range tokens {
		tok = strings.TrimSpace(tok)
		val, err := strconv.ParseFloat(tok, 64)
		if err != nil {
			return 0, false
		}
		if i == 0 {
			result = val
		} else {
			switch ops[i-1] {
			case '*':
				result *= val
			case '/':
				if val == 0 {
					return 0, false
				}
				result /= val
			}
		}
	}
	return result, true
}

// splitMultiplicative splits expr on '*' and '/' operators.
func splitMultiplicative(expr string) ([]string, []rune) {
	var terms []string
	var ops []rune
	start := 0
	for i, ch := range expr {
		if ch == '*' || ch == '/' {
			terms = append(terms, expr[start:i])
			ops = append(ops, ch)
			start = i + 1
		}
	}
	terms = append(terms, expr[start:])
	return terms, ops
}

func prependUPCHistory(history []string, upc string) []string {
	const maxHistory = 50
	result := make([]string, 0, len(history)+1)
	result = append(result, upc)
	for _, h := range history {
		if h != upc {
			result = append(result, h)
		}
	}
	if len(result) > maxHistory {
		result = result[:maxHistory]
	}
	return result
}

// resolveExpiry converts the raw expiry input to a YYYY-MM-DD date string.
// A plain positive integer is interpreted as days from today; "" or any
// non-positive integer becomes the sentinel never-expires date; anything
// else (e.g. "2006-01-02") is passed through as-is.
func resolveExpiry(raw string) string {
	if n, err := strconv.Atoi(raw); err == nil {
		if n <= 0 {
			return "2999-12-31"
		}
		return time.Now().AddDate(0, 0, n).Format("2006-01-02")
	}
	if raw == "" {
		return "2999-12-31"
	}
	return raw
}

func (m Model) daysFromExpiry(dateStr string) *int {
	if dateStr == "2999-12-31" {
		n := -1
		return &n
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil
	}
	days := int(time.Until(t).Hours() / 24)
	if days < 0 {
		return nil
	}
	return &days
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	if m.showHelp {
		return ui.RenderHelp(m.width, m.height)
	}

	header := ui.RenderHeader(m.mode, m.width)
	hSep := ui.StyleSeparator.Render(strings.Repeat("─", m.width))
	bSep := ui.StyleSeparator.Render(strings.Repeat("─", m.width))

	// header + hSep + bSep + inputLine = 4 fixed lines
	bodyH := m.height - 4
	if bodyH < 1 {
		bodyH = 1
	}

	var body string
	if m.width >= expPanelMinWidth {
		mainW := m.width - expPanelWidth - 1
		mainContent := m.renderMainContent(mainW, bodyH)
		panelContent := m.renderExpiringSoonPanel(bodyH)

		mainBlock := lipgloss.NewStyle().Width(mainW).Height(bodyH).Render(mainContent)

		var sepLines []string
		for i := 0; i < bodyH; i++ {
			sepLines = append(sepLines, ui.StyleSeparator.Render("│"))
		}
		sepBlock := strings.Join(sepLines, "\n")

		panelBlock := lipgloss.NewStyle().Width(expPanelWidth).Height(bodyH).Render(panelContent)

		body = lipgloss.JoinHorizontal(lipgloss.Top, mainBlock, sepBlock, panelBlock)
	} else {
		mainContent := m.renderMainContent(m.width, bodyH)
		body = lipgloss.NewStyle().Height(bodyH).Render(mainContent)
	}

	return strings.Join([]string{header, hSep, body, bSep, m.renderInputLine()}, "\n")
}

func (m Model) renderInputLine() string {
	switch m.state {
	case StateIdle:
		return " > " + m.input.View()
	case StateLookupView:
		return " " + ui.StyleHint.Render("n = notes  •  p = price history  •  t = transfer  •  Esc/Enter = dismiss")
	case StateTransfer:
		return " " + ui.StyleHint.Render("Tab/↓ next field  •  Enter submit  •  Esc cancel")
	case StatePriceHistory:
		return " " + ui.StyleHint.Render("j/k = navigate  •  Esc/p = back")
	case StateMealPlan:
		return " " + ui.StyleHint.Render("j/k = scroll  •  r = refresh  •  Esc/q = back")
	case StateEditNotes:
		return " " + ui.StyleHint.Render("Enter to save  •  Esc to cancel")
	case StateShoppingListPrompt:
		return " " + ui.StyleHint.Render("y = yes, any other key = no")
	default:
		if m.loading {
			return " " + ui.StyleHint.Render("loading...")
		}
		return " " + ui.StyleHint.Render("(form active)")
	}
}

func (m Model) renderMainContent(width, bodyH int) string {
	var sections []string

	maxLogLines := 5
	if logView := ui.RenderLog(m.logEntries, maxLogLines); logView != "" {
		sections = append(sections, logView)
		sections = append(sections, ui.StyleSeparator.Render(strings.Repeat("─", width)))
	}

	switch m.state {
	case StateIdle:
		if m.statusMsg != "" {
			if m.statusErr {
				sections = append(sections, " "+ui.StyleError.Render(m.statusMsg))
			} else {
				sections = append(sections, " "+ui.StyleSuccess.Render(m.statusMsg))
			}
			sections = append(sections, "")
		}
	case StateLookup:
		sections = append(sections, " "+ui.StyleInfo.Render(fmt.Sprintf("Looking up UPC %s...", m.currentUPC)))
	case StateDisplay, StateForm:
		sections = append(sections, m.renderProductInfo())
		sections = append(sections, "")
		sections = append(sections, m.form.View())
	case StateEditName:
		sections = append(sections, m.renderProductInfo())
		sections = append(sections, "")
		sections = append(sections, " "+ui.StyleBold.Render("Edit name: ")+m.editInput.View())
		sections = append(sections, " "+ui.StyleHint.Render("Enter to save, Esc to cancel"))
	case StateEditNotes:
		sections = append(sections, m.renderLookupView())
		sections = append(sections, "")
		sections = append(sections, " "+ui.StyleBold.Render("Notes: ")+m.editInput.View())
		sections = append(sections, " "+ui.StyleHint.Render("Enter to save, Esc to cancel"))
	case StateConsume:
		if m.currentProduct != nil {
			sections = append(sections, fmt.Sprintf(" %s %s", ui.StyleBold.Render("Consuming:"), m.currentProduct.Name))
			sections = append(sections, " "+ui.StyleHint.Render("Processing..."))
		}
	case StateSearch:
		sections = append(sections, m.search.View())
	case StateLookupView:
		sections = append(sections, m.renderLookupView())
	case StateTransfer:
		sections = append(sections, m.renderLookupView())
		sections = append(sections, "")
		sections = append(sections, " "+ui.StyleBold.Render("Transfer stock"))
		sections = append(sections, m.form.View())
	case StatePriceHistory:
		sections = append(sections, m.renderPriceHistoryView(bodyH))
	case StateMealPlan:
		sections = append(sections, m.renderMealPlanView(bodyH))
	case StateShoppingListPrompt:
		if m.currentProduct != nil {
			sections = append(sections, fmt.Sprintf(" %s %s",
				ui.StyleBold.Render("Consumed:"), m.currentProduct.Name))
			sections = append(sections, " "+ui.StyleWarning.Render("Stock is now zero."))
			sections = append(sections, "")
			sections = append(sections, " Add to shopping list? [y/N]")
		}
	}

	return strings.Join(sections, "\n")
}

func (m Model) renderExpiringSoonPanel(bodyH int) string {
	const daysColW = 3
	nameColW := expPanelWidth - 2 - daysColW // 1 for " " prefix, 1 for " " before days

	var lines []string
	lines = append(lines, " "+ui.StyleBold.Render("Expiring Soon"))
	lines = append(lines, " "+ui.StyleSeparator.Render(strings.Repeat("─", expPanelWidth-2)))

	if !m.expiringSoonLoaded {
		lines = append(lines, " "+ui.StyleHint.Render("Loading..."))
		return strings.Join(lines, "\n")
	}

	if len(m.expiringSoon) == 0 {
		lines = append(lines, " "+ui.StyleHint.Render("None this week"))
		return strings.Join(lines, "\n")
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	maxItems := bodyH - 2
	for i, item := range m.expiringSoon {
		if i >= maxItems {
			break
		}

		t, _ := time.Parse("2006-01-02", item.BestBeforeDate)
		days := int(t.Sub(today).Hours() / 24)

		var daysText string
		var daysStyle lipgloss.Style
		switch {
		case days < 0:
			daysText = "exp"
			daysStyle = ui.StyleError
		case days == 0:
			daysText = " 0d"
			daysStyle = ui.StyleError
		case days <= 2:
			daysText = fmt.Sprintf("%2dd", days)
			daysStyle = ui.StyleWarning
		default:
			daysText = fmt.Sprintf("%2dd", days)
			daysStyle = ui.StyleInfo
		}

		name := item.ProductName
		runes := []rune(name)
		if len(runes) > nameColW {
			name = string(runes[:nameColW-1]) + "…"
		} else if len(runes) < nameColW {
			name += strings.Repeat(" ", nameColW-len(runes))
		}

		prefix := " "
		if i == m.expPanelCursor {
			prefix = ui.StyleWarning.Render(">")
		}
		lines = append(lines, prefix+name+" "+daysStyle.Render(daysText))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderLookupView() string {
	var lines []string

	lines = append(lines, " "+ui.StyleBold.Render("Product Overview"))
	lines = append(lines, "")

	if m.currentProduct == nil {
		return strings.Join(lines, "\n")
	}

	lines = append(lines, fmt.Sprintf(" %s %s  %s",
		ui.StyleLabel.Render("Product:"),
		m.currentProduct.Name,
		ui.StyleHint.Render(fmt.Sprintf("[id:%d]", m.currentProduct.ID))))

	if m.currentUPC != "" {
		lines = append(lines, fmt.Sprintf(" %s %s",
			ui.StyleLabel.Render("UPC:"),
			m.currentUPC))
	}

	if m.stockInfo != nil {
		stockLine := fmt.Sprintf("%g", m.stockInfo.StockAmount)
		if m.defaults != nil {
			for _, u := range m.defaults.QuantityUnits {
				if u.ID == m.currentProduct.QuIDStock {
					stockLine += " " + u.Name
					break
				}
			}
		}
		lines = append(lines, fmt.Sprintf(" %s %s",
			ui.StyleLabel.Render("In stock:"),
			stockLine))
	} else {
		lines = append(lines, fmt.Sprintf(" %s %s",
			ui.StyleLabel.Render("In stock:"),
			ui.StyleHint.Render("loading...")))
	}

	if m.defaults != nil {
		for _, loc := range m.defaults.Locations {
			if loc.ID == m.currentProduct.LocationID {
				lines = append(lines, fmt.Sprintf(" %s %s",
					ui.StyleLabel.Render("Location:"),
					loc.Name))
				break
			}
		}
		for _, s := range m.defaults.Stores {
			if s.ID == m.currentProduct.ShoppingLocationID {
				lines = append(lines, fmt.Sprintf(" %s %s",
					ui.StyleLabel.Render("Store:"),
					s.Name))
				break
			}
		}
	}

	if m.currentProduct.DefaultBestBeforeDays > 0 {
		lines = append(lines, fmt.Sprintf(" %s %d days",
			ui.StyleLabel.Render("Shelf life:"),
			m.currentProduct.DefaultBestBeforeDays))
	} else if m.currentProduct.DefaultBestBeforeDays < 0 {
		lines = append(lines, fmt.Sprintf(" %s never",
			ui.StyleLabel.Render("Shelf life:")))
	}

	if m.stockInfo != nil && m.stockInfo.LastPrice > 0 {
		unitName := ""
		if m.defaults != nil {
			for _, u := range m.defaults.QuantityUnits {
				if u.ID == m.currentProduct.QuIDPurchase {
					unitName = u.Name
					break
				}
			}
		}
		priceLabel := fmt.Sprintf("$%.2f", m.stockInfo.LastPrice)
		if unitName != "" {
			priceLabel += "/" + unitName
		}
		lines = append(lines, fmt.Sprintf(" %s %s",
			ui.StyleLabel.Render("Last price:"),
			priceLabel))
	}

	if m.currentProduct.Description != "" {
		lines = append(lines, fmt.Sprintf(" %s %s",
			ui.StyleLabel.Render("Notes:"),
			m.currentProduct.Description))
	}

	if m.offInfo != nil && m.offInfo.Name != "" {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf(" %s %s",
			ui.StyleLabel.Render("OFF name:"),
			m.offInfo.Name))
		if m.offInfo.Categories != "" {
			cats := []rune(m.offInfo.Categories)
			display := string(cats)
			if len(cats) > 60 {
				display = string(cats[:57]) + "..."
			}
			lines = append(lines, fmt.Sprintf(" %s %s",
				ui.StyleLabel.Render("Categories:"),
				display))
		}
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderPriceHistoryView(bodyH int) string {
	var lines []string

	name := ""
	if m.currentProduct != nil {
		name = m.currentProduct.Name
	}
	lines = append(lines, fmt.Sprintf(" %s %s", ui.StyleBold.Render("Price History:"), name))
	lines = append(lines, "")

	if m.priceHistory == nil {
		lines = append(lines, " "+ui.StyleHint.Render("Loading..."))
		return strings.Join(lines, "\n")
	}

	if len(m.priceHistory) == 0 {
		lines = append(lines, " "+ui.StyleHint.Render("No purchase history found."))
		return strings.Join(lines, "\n")
	}

	unitName := ""
	if m.defaults != nil && m.currentProduct != nil {
		for _, u := range m.defaults.QuantityUnits {
			if u.ID == m.currentProduct.QuIDPurchase {
				unitName = u.Name
				break
			}
		}
	}

	// 2 header lines already consumed; clamp visible rows to remaining height
	maxItems := bodyH - 2
	if maxItems < 1 {
		maxItems = 1
	}

	// Scroll offset: keep cursor visible
	offset := 0
	if m.priceHistoryCursor >= maxItems {
		offset = m.priceHistoryCursor - maxItems + 1
	}

	for i, txn := range m.priceHistory {
		if i < offset {
			continue
		}
		if i-offset >= maxItems {
			break
		}

		date := txn.Date
		if len(date) >= 10 {
			date = date[:10]
		}

		var priceStr string
		if txn.Price > 0 {
			priceStr = fmt.Sprintf("$%.2f", txn.Price)
			if unitName != "" {
				priceStr += "/" + unitName
			}
		} else {
			priceStr = ui.StyleHint.Render("(no price)")
		}

		storeName := ""
		if m.defaults != nil && txn.ShoppingLocationID > 0 {
			for _, s := range m.defaults.Stores {
				if s.ID == txn.ShoppingLocationID {
					storeName = "  " + ui.StyleHint.Render(s.Name)
					break
				}
			}
		}

		prefix := "  "
		if i == m.priceHistoryCursor {
			prefix = ui.StyleWarning.Render("> ")
		}
		lines = append(lines, fmt.Sprintf("%s%s  %s%s",
			prefix, ui.StyleLabel.Render(date), priceStr, storeName))
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderMealPlanView(bodyH int) string {
	var lines []string
	lines = append(lines, " "+ui.StyleBold.Render("Meal Plan — Next 7 Days"))
	lines = append(lines, "")

	if !m.mealPlanLoaded {
		lines = append(lines, " "+ui.StyleHint.Render("Loading..."))
		return strings.Join(lines, "\n")
	}
	if len(m.mealPlan) == 0 {
		lines = append(lines, " "+ui.StyleHint.Render("No meals planned for the next 7 days."))
		return strings.Join(lines, "\n")
	}

	today := time.Now().Format("2006-01-02")
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")

	// Group items by day preserving order (items are sorted by day from API)
	type group struct {
		day   string
		items []api.MealPlanItem
	}
	var groups []group
	dayIndex := map[string]int{}
	for _, item := range m.mealPlan {
		if idx, ok := dayIndex[item.Day]; ok {
			groups[idx].items = append(groups[idx].items, item)
		} else {
			dayIndex[item.Day] = len(groups)
			groups = append(groups, group{day: item.Day, items: []api.MealPlanItem{item}})
		}
	}

	// Build scrollable content lines
	var content []string
	for _, g := range groups {
		t, err := time.Parse("2006-01-02", g.day)
		var dayLabel string
		if err != nil {
			dayLabel = g.day
		} else {
			switch g.day {
			case today:
				dayLabel = t.Format("Mon Jan 02") + "  " + ui.StyleSuccess.Render("(Today)")
			case tomorrow:
				dayLabel = t.Format("Mon Jan 02") + "  " + ui.StyleInfo.Render("(Tomorrow)")
			default:
				dayLabel = t.Format("Mon Jan 02")
			}
		}
		content = append(content, " "+ui.StyleLabel.Render(dayLabel))

		for _, item := range g.items {
			var desc string
			switch {
			case item.RecipeID != nil:
				name, ok := m.mealPlanRecipes[*item.RecipeID]
				if !ok {
					name = fmt.Sprintf("Recipe #%d", *item.RecipeID)
				}
				if item.RecipeServings > 0 {
					desc = fmt.Sprintf("%s  %s", name, ui.StyleHint.Render(fmt.Sprintf("(%.0f srv)", item.RecipeServings)))
				} else {
					desc = name
				}
			case item.ProductID != nil:
				name := fmt.Sprintf("Product #%d", *item.ProductID)
				for _, p := range m.allProducts {
					if p.ID == *item.ProductID {
						name = p.Name
						break
					}
				}
				if item.ProductAmount > 0 {
					desc = fmt.Sprintf("%s  %s", name, ui.StyleHint.Render(fmt.Sprintf("×%.0f", item.ProductAmount)))
				} else {
					desc = name
				}
			case item.Note != "":
				desc = item.Note
			default:
				continue
			}
			content = append(content, "   • "+desc)
			if item.Note != "" && (item.RecipeID != nil || item.ProductID != nil) {
				content = append(content, "     "+ui.StyleHint.Render(item.Note))
			}
		}
		content = append(content, "")
	}

	// Clamp scroll offset
	headerLines := len(lines)
	maxVisible := bodyH - headerLines
	if maxVisible < 1 {
		maxVisible = 1
	}
	maxOffset := len(content) - maxVisible
	if maxOffset < 0 {
		maxOffset = 0
	}
	offset := m.mealPlanOffset
	if offset > maxOffset {
		offset = maxOffset
	}

	end := offset + maxVisible
	if end > len(content) {
		end = len(content)
	}
	lines = append(lines, content[offset:end]...)

	return strings.Join(lines, "\n")
}

func (m Model) renderProductInfo() string {
	var lines []string

	if m.currentUPC != "" {
		lines = append(lines, fmt.Sprintf(" %s %s", ui.StyleBold.Render("UPC:"), m.currentUPC))
	}

	if m.isNewProduct {
		lines = append(lines, " "+ui.StyleWarning.Render("NEW PRODUCT"))
	} else if m.currentProduct != nil {
		lines = append(lines, " "+ui.StyleSuccess.Render("✓ Found in Grocy"))
		lines = append(lines, fmt.Sprintf(" %s %s  %s",
			ui.StyleLabel.Render("Product:"),
			m.currentProduct.Name,
			ui.StyleHint.Render(fmt.Sprintf("[id:%d]", m.currentProduct.ID))))
		if m.stockInfo != nil {
			lines = append(lines, fmt.Sprintf(" %s %g",
				ui.StyleLabel.Render("In stock:"),
				m.stockInfo.StockAmount))
		}
		if m.currentProduct.DefaultBestBeforeDays > 0 {
			lines = append(lines, fmt.Sprintf(" %s %d days",
				ui.StyleLabel.Render("Shelf life:"),
				m.currentProduct.DefaultBestBeforeDays))
		}
	}

	if m.offInfo != nil && m.offInfo.Name != "" {
		lines = append(lines, fmt.Sprintf(" %s %s  %s",
			ui.StyleLabel.Render("OFF:"),
			m.offInfo.Name,
			ui.StyleHint.Render("(Open Food Facts)")))
		if m.offInfo.Categories != "" {
			cats := []rune(m.offInfo.Categories)
			display := string(cats)
			if len(cats) > 60 {
				display = string(cats[:57]) + "..."
			}
			lines = append(lines, fmt.Sprintf(" %s %s",
				ui.StyleLabel.Render("Categories:"),
				display))
		}
		if m.offInfo.ShelfLifeDays != nil {
			lines = append(lines, fmt.Sprintf(" %s %d days",
				ui.StyleLabel.Render("Shelf life:"),
				*m.offInfo.ShelfLifeDays))
		}
	}

	return strings.Join(lines, "\n")
}
