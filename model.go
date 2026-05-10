package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kevin/grocy-scanner/api"
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
)

type Model struct {
	state    AppState
	mode     string // "add" or "consume"
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
			m.statusMsg = "Lookup error: " + msg.err.Error()
			m.statusErr = true
			m.state = StateIdle
			return m, m.input.Focus()
		}
		return m.handleLookupResult(msg)

	case actionResultMsg:
		m.loading = false
		if msg.err != nil {
			m.statusMsg = "Error: " + msg.err.Error()
			m.statusErr = true
		} else {
			m.logEntries = append([]ui.LogEntry{msg.entry}, m.logEntries...)
			m.statusMsg = ""
		}
		m.state = StateIdle
		m.currentProduct = nil
		m.offInfo = nil
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
			if m.mode == "add" {
				m.mode = "consume"
			} else {
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

	if m.isNewProduct {
		m.state = StateForm
		m.form = m.buildNewProductForm()
	} else {
		m.state = StateDisplay
		m.form = m.buildExistingProductForm()
	}

	return m, nil
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

	locationDefault := ""
	locationHint := ""
	if m.defaults != nil && len(m.defaults.Locations) > 0 {
		var parts []string
		for i, loc := range m.defaults.Locations {
			parts = append(parts, fmt.Sprintf("%d)%s", i+1, loc.Name))
		}
		locationDefault = m.defaults.Locations[0].Name
		locationHint = strings.Join(parts, " ")
	}

	fields := []ui.FormField{
		{Label: "Name", Default: defaultName},
		{Label: "Short name", Default: ""},
		{Label: "Expires", Default: expiryDefault, Hint: expiryHint},
		{Label: "Location", Default: locationDefault, Hint: locationHint},
		{Label: "Quantity", Default: "1"},
		{Label: "Price", Default: ""},
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

	fields := []ui.FormField{
		{Label: "Expires", Default: expiryDefault, Hint: expiryHint},
		{Label: "Location", Default: locationDefault, Hint: locationHint},
		{Label: "Quantity", Default: "1"},
		{Label: "Price", Default: ""},
	}

	return ui.NewForm(fields)
}

func (m Model) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	cmd := m.form.Update(msg)

	if m.form.Cancelled {
		m.state = StateIdle
		m.currentProduct = nil
		m.offInfo = nil
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

		m.state = StateDisplay
		m.form = m.buildExistingProductForm()
		return m, nil
	}

	return m, cmd
}

func (m Model) handleEditNameKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "enter":
		newName := strings.TrimSpace(m.editInput.Value())
		if newName != "" && newName != m.currentProduct.Name {
			if !m.testMode {
				m.grocy.UpdateProductName(m.currentProduct.ID, newName)
			}
			m.currentProduct.Name = newName
			m.statusMsg = "Name updated"
			m.statusErr = false
		}
		m.state = StateDisplay
		return m, nil
	case "esc":
		m.state = StateDisplay
		return m, nil
	}

	var cmd tea.Cmd
	m.editInput, cmd = m.editInput.Update(msg)
	return m, cmd
}

func (m Model) submitForm() tea.Cmd {
	return func() tea.Msg {
		var productName string
		var quantity int
		var price float64
		var bestBefore string
		var locationID int

		if m.isNewProduct {
			productName = m.form.Value(0)
			// shortName = m.form.Value(1)
			bestBefore = m.form.Value(2)
			locationID = m.resolveLocation(m.form.Value(3))
			quantity = m.parseQuantity(m.form.Value(4))
			price = m.parsePrice(m.form.Value(5))
		} else {
			productName = m.currentProduct.Name
			bestBefore = m.form.Value(0)
			locationID = m.resolveLocation(m.form.Value(1))
			quantity = m.parseQuantity(m.form.Value(2))
			price = m.parsePrice(m.form.Value(3))
		}

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
			p, err := m.grocy.CreateProduct(productName, shelfDays, m.defaults, shortName, locationID, &freezeDays, shelfDays)
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

func (m Model) parseQuantity(val string) int {
	n, err := strconv.Atoi(val)
	if err != nil || n < 1 {
		return 1
	}
	return n
}

func (m Model) parsePrice(val string) float64 {
	val = strings.TrimPrefix(val, "$")
	f, err := strconv.ParseFloat(val, 64)
	if err != nil || f < 0 {
		return 0
	}
	return f
}

func (m Model) daysFromExpiry(dateStr string) *int {
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
	} else if m.loading {
		inputBar += " " + ui.StyleHint.Render("loading...")
	} else {
		inputBar += " " + ui.StyleHint.Render("(form active)")
	}

	return content + "\n" + inputBar
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
			cats := m.offInfo.Categories
			if len(cats) > 60 {
				cats = cats[:57] + "..."
			}
			lines = append(lines, fmt.Sprintf(" %s %s",
				ui.StyleLabel.Render("Categories:"),
				cats))
		}
		if m.offInfo.ShelfLifeDays != nil {
			lines = append(lines, fmt.Sprintf(" %s %d days",
				ui.StyleLabel.Render("Shelf life:"),
				*m.offInfo.ShelfLifeDays))
		}
	}

	return strings.Join(lines, "\n")
}
