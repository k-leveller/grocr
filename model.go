package main

import (
	"fmt"
	"math"
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

	// Status message
	statusMsg string
	statusErr bool

	// Loading
	loading    bool
	lookupSeq  int
}

// Messages
type lookupResultMsg struct {
	seq     int
	product *api.Product
	offInfo *api.OFFProduct
	err     error
}

type actionResultMsg struct {
	entry ui.LogEntry
	err   error
}

type stockInfoMsg struct {
	productID int
	info      *api.StockInfo
}

type productsLoadedMsg struct {
	products []api.Product
}

type defaultsLoadedMsg struct {
	defaults *api.Defaults
	err      error
}

func NewModel(grocy *api.GrocyClient, off *api.OFFClient, testMode bool) Model {
	ti := textinput.New()
	ti.Placeholder = "Scan UPC or type command..."
	ti.Focus()
	ti.CharLimit = 20

	return Model{
		state:    StateIdle,
		mode:     "add",
		input:    ti,
		grocy:    grocy,
		off:      off,
		testMode: testMode,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.loadDefaults(),
		m.loadProducts(),
	)
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
			if msg.entry.Action == "consume" {
				logger.LogConsume(msg.entry.ProductName, msg.entry.Quantity)
			} else {
				logger.LogAdd(msg.entry.ProductName, msg.entry.Quantity, msg.entry.Location, msg.entry.Expiry)
			}
			m.statusMsg = ""
		}
		m.state = StateIdle
		m.currentProduct = nil
		m.offInfo = nil
		m.stockInfo = nil
		m.input.SetValue("")
		return m, m.input.Focus()

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
	case StateEditName:
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
			m.search = ui.NewSearch(m.allProducts)
			m.input.Blur()
			return m, nil
		}
	case "ctrl+n":
		m.input.SetValue("")
		return m.startManualProductEntry()
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
			return m, nil
		}
		m.currentUPC = upc
		m.loading = true
		m.statusMsg = ""
		m.lookupSeq++
		m.input.Blur()
		return m, m.lookupUPC(upc, m.lookupSeq)
	}

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
		err = m.grocy.ConsumeStock(m.currentProduct.ID, 1)
		entry := ui.LogEntry{
			ProductName: m.currentProduct.Name,
			Quantity:    1,
			Action:      "consume",
			Success:     err == nil,
			Time:        time.Now(),
		}
		return actionResultMsg{entry: entry, err: err}
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
	}
	return m, nil
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

	var sections []string

	// Header
	header := ui.RenderHeader(m.mode, m.width)
	sections = append(sections, header)
	sections = append(sections, ui.StyleSeparator.Render(strings.Repeat("─", m.width)))

	// Log
	maxLogLines := 5
	if logView := ui.RenderLog(m.logEntries, maxLogLines); logView != "" {
		sections = append(sections, logView)
	}

	// Separator between log and current scan
	if len(m.logEntries) > 0 {
		sections = append(sections, ui.StyleSeparator.Render(strings.Repeat("─", m.width)))
	}

	// Main content area
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
		if m.state == StateEditName {
			sections = append(sections, " "+ui.StyleBold.Render("Edit name: ")+m.editInput.View())
		} else {
			sections = append(sections, m.form.View())
		}
	case StateEditName:
		sections = append(sections, m.renderProductInfo())
		sections = append(sections, "")
		sections = append(sections, " "+ui.StyleBold.Render("Edit name: ")+m.editInput.View())
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
	}

	// Fill remaining space
	content := strings.Join(sections, "\n")
	contentHeight := lipgloss.Height(content)
	remainingHeight := m.height - contentHeight - 2 // 2 for input bar + separator
	if remainingHeight > 0 {
		content += strings.Repeat("\n", remainingHeight)
	}

	// Input bar at bottom
	inputBar := ui.StyleSeparator.Render(strings.Repeat("─", m.width)) + "\n"
	if m.state == StateIdle {
		inputBar += " > " + m.input.View()
	} else if m.state == StateLookupView {
		inputBar += " " + ui.StyleHint.Render("Esc or Enter to dismiss")
	} else if m.loading {
		inputBar += " " + ui.StyleHint.Render("loading...")
	} else {
		inputBar += " " + ui.StyleHint.Render("(form active)")
	}

	return content + "\n" + inputBar
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
