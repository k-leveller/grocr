package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/k-leveller/grocr/config"
)

// newTestClient creates a GrocyClient pointed at a test HTTP server.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*GrocyClient, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	cfg := &config.Config{BaseURL: srv.URL, APIKey: "test-key"}
	return NewGrocyClient(cfg), srv
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func TestNewGrocyClient(t *testing.T) {
	cfg := &config.Config{BaseURL: "https://grocy.example.com/", APIKey: "abc123"}
	c := NewGrocyClient(cfg)
	if c.baseURL != "https://grocy.example.com" {
		t.Errorf("baseURL = %q, want trailing slash stripped", c.baseURL)
	}
	if c.apiKey != "abc123" {
		t.Errorf("apiKey = %q, want %q", c.apiKey, "abc123")
	}
}

func TestNewGrocyClient_TLSSkipVerify(t *testing.T) {
	cfg := &config.Config{BaseURL: "https://grocy.local", APIKey: "key", TLSSkipVerify: true}
	c := NewGrocyClient(cfg)
	if c.httpClient == nil {
		t.Fatal("expected non-nil http client")
	}
}

func TestDisplayNameUserfield(t *testing.T) {
	cfg := &config.Config{BaseURL: "http://x", APIKey: "k", DisplayNameUserfield: "short_name"}
	c := NewGrocyClient(cfg)
	if c.DisplayNameUserfield() != "short_name" {
		t.Errorf("DisplayNameUserfield() = %q, want %q", c.DisplayNameUserfield(), "short_name")
	}
}

func TestGetProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("GROCY-API-KEY") != "test-key" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			writeJSON(w, Product{ID: 42, Name: "Test Product"})
		})
		p, err := client.GetProduct(42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.ID != 42 || p.Name != "Test Product" {
			t.Errorf("unexpected product: %+v", p)
		}
	})

	t.Run("HTTP error", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "not found", http.StatusNotFound)
		})
		_, err := client.GetProduct(99)
		if err == nil {
			t.Error("expected error for 404, got nil")
		}
	})
}

func TestGetAllProducts(t *testing.T) {
	products := []Product{
		{ID: 1, Name: "Apple"},
		{ID: 2, Name: "Banana"},
	}
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, products)
	})
	got, err := client.GetAllProducts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
	if got[0].Name != "Apple" || got[1].Name != "Banana" {
		t.Errorf("unexpected products: %+v", got)
	}
}

func TestFindProductByBarcode(t *testing.T) {
	t.Run("barcode not found returns nil", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, []ProductBarcode{})
		})
		p, err := client.FindProductByBarcode("000000000000")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p != nil {
			t.Errorf("expected nil, got %+v", p)
		}
	})

	t.Run("barcode found returns product", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/objects/product_barcodes") {
				writeJSON(w, []ProductBarcode{{ID: 1, ProductID: 5, Barcode: "012345678905"}})
			} else {
				writeJSON(w, Product{ID: 5, Name: "Oats"})
			}
		})
		p, err := client.FindProductByBarcode("012345678905")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p == nil {
			t.Fatal("expected non-nil product")
		}
		if p.ID != 5 || p.Name != "Oats" {
			t.Errorf("unexpected product: %+v", p)
		}
	})

	t.Run("HTTP error propagated", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "server error", http.StatusInternalServerError)
		})
		_, err := client.FindProductByBarcode("000000000000")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestProductNameExists(t *testing.T) {
	t.Run("name exists", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, []Product{{ID: 1, Name: "Milk"}})
		})
		ok, err := client.ProductNameExists("Milk")
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Error("expected true, got false")
		}
	})

	t.Run("name does not exist", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, []Product{})
		})
		ok, err := client.ProductNameExists("Nothing")
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Error("expected false, got true")
		}
	})
}

func TestGetLocations(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, []Location{{ID: 1, Name: "Fridge"}, {ID: 2, Name: "Freezer"}})
	})
	locs, err := client.GetLocations()
	if err != nil {
		t.Fatal(err)
	}
	if len(locs) != 2 || locs[0].Name != "Fridge" {
		t.Errorf("unexpected locations: %+v", locs)
	}
}

func TestGetStores(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, []Store{{ID: 1, Name: "Whole Foods"}, {ID: 2, Name: "Costco"}})
	})
	stores, err := client.GetStores()
	if err != nil {
		t.Fatal(err)
	}
	if len(stores) != 2 || stores[1].Name != "Costco" {
		t.Errorf("unexpected stores: %+v", stores)
	}
}

func TestGetQuantityUnits(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, []QuantityUnit{{ID: 1, Name: "Piece"}, {ID: 2, Name: "Oz"}})
	})
	units, err := client.GetQuantityUnits()
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 2 {
		t.Fatalf("len = %d, want 2", len(units))
	}
}

func TestGetDefaults_pieceUnit(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/locations"):
			writeJSON(w, []Location{{ID: 3, Name: "Pantry"}})
		case strings.HasSuffix(r.URL.Path, "/quantity_units"):
			writeJSON(w, []QuantityUnit{{ID: 10, Name: "Oz"}, {ID: 11, Name: "Piece"}})
		case strings.HasSuffix(r.URL.Path, "/shopping_locations"):
			writeJSON(w, []Store{{ID: 5, Name: "Market"}})
		}
	})
	d, err := client.GetDefaults()
	if err != nil {
		t.Fatal(err)
	}
	if d.LocationID != 3 {
		t.Errorf("LocationID = %d, want 3", d.LocationID)
	}
	// "Piece" should be selected as the preferred quantity unit
	if d.QuID != 11 {
		t.Errorf("QuID = %d, want 11 (Piece)", d.QuID)
	}
	if len(d.Stores) != 1 || d.Stores[0].Name != "Market" {
		t.Errorf("unexpected stores: %+v", d.Stores)
	}
}

func TestGetDefaults_firstUnitFallback(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/locations"):
			writeJSON(w, []Location{{ID: 1, Name: "Fridge"}})
		case strings.HasSuffix(r.URL.Path, "/quantity_units"):
			// No "piece/pcs/ea/each" unit
			writeJSON(w, []QuantityUnit{{ID: 7, Name: "Kg"}, {ID: 8, Name: "Litre"}})
		case strings.HasSuffix(r.URL.Path, "/shopping_locations"):
			writeJSON(w, []Store{})
		}
	})
	d, err := client.GetDefaults()
	if err != nil {
		t.Fatal(err)
	}
	// Falls back to first unit
	if d.QuID != 7 {
		t.Errorf("QuID = %d, want 7 (first unit)", d.QuID)
	}
}

func TestGetDefaults_emptyResponses(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, []interface{}{})
	})
	d, err := client.GetDefaults()
	if err != nil {
		t.Fatal(err)
	}
	// Should use hardcoded fallbacks
	if d.LocationID != 1 {
		t.Errorf("LocationID = %d, want 1 (default)", d.LocationID)
	}
	if d.QuID != 1 {
		t.Errorf("QuID = %d, want 1 (default)", d.QuID)
	}
}

func TestAddStock(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "expected POST", http.StatusMethodNotAllowed)
				return
			}
			w.WriteHeader(http.StatusOK)
			writeJSON(w, map[string]string{})
		})
		err := client.AddStock(1, 2.0, 3.99, "2026-12-31", 1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("HTTP error", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "server error", http.StatusInternalServerError)
		})
		err := client.AddStock(1, 1.0, 0, "", 0)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestConsumeStock(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "expected POST", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, map[string]string{})
	})
	err := client.ConsumeStock(1, 1.0, false)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTransferStock(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{})
	})
	err := client.TransferStock(1, 2.0, 1, 2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetStock(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, StockInfo{StockAmount: 5.5, LastPrice: 3.99})
	})
	info, err := client.GetStock(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.StockAmount != 5.5 || info.LastPrice != 3.99 {
		t.Errorf("unexpected stock info: %+v", info)
	}
}

func TestGetStockSnapshot(t *testing.T) {
	// Stock entries should be returned sorted by product name
	raw := []struct {
		ProductID      int     `json:"product_id"`
		Amount         float64 `json:"amount"`
		LocationID     int     `json:"location_id"`
		BestBeforeDate string  `json:"best_before_date"`
		Product        struct {
			Name string `json:"name"`
		} `json:"product"`
	}{
		{ProductID: 2, Amount: 3, Product: struct{ Name string `json:"name"` }{Name: "Zucchini"}},
		{ProductID: 1, Amount: 1, Product: struct{ Name string `json:"name"` }{Name: "Apple"}},
		{ProductID: 3, Amount: 2, Product: struct{ Name string `json:"name"` }{Name: "Mango"}},
	}

	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, raw)
	})
	entries, err := client.GetStockSnapshot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("len = %d, want 3", len(entries))
	}
	// Verify sorted order
	names := []string{entries[0].ProductName, entries[1].ProductName, entries[2].ProductName}
	want := []string{"Apple", "Mango", "Zucchini"}
	for i, n := range names {
		if n != want[i] {
			t.Errorf("entries[%d].ProductName = %q, want %q", i, n, want[i])
		}
	}
}

func TestAddToShoppingList(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "expected POST", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, map[string]string{})
	})
	err := client.AddToShoppingList(1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddBarcode(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{})
	})
	err := client.AddBarcode(1, "012345678905")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetPurchaseHistory(t *testing.T) {
	txns := []StockTransaction{
		{ID: 1, ProductID: 5, Type: "purchase", Amount: 2},
		{ID: 2, ProductID: 5, Type: "purchase", Amount: 1},
	}
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, txns)
	})
	got, err := client.GetPurchaseHistory(5, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

func TestGetExpiringSoon(t *testing.T) {
	// Items with dates in the past or near future should be included;
	// items far in the future and sentinel dates should be excluded.
	raw := []struct {
		ProductID      int    `json:"product_id"`
		BestBeforeDate string `json:"best_before_date"`
		Product        struct {
			Name string `json:"name"`
		} `json:"product"`
	}{
		{ProductID: 1, BestBeforeDate: "2020-01-01", Product: struct{ Name string `json:"name"` }{Name: "Expired"}},
		{ProductID: 2, BestBeforeDate: "2999-12-31", Product: struct{ Name string `json:"name"` }{Name: "Never Expires"}},
		{ProductID: 3, BestBeforeDate: "", Product: struct{ Name string `json:"name"` }{Name: "No Date"}},
		{ProductID: 4, BestBeforeDate: "2099-12-31", Product: struct{ Name string `json:"name"` }{Name: "Far Future"}},
	}

	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, raw)
	})
	items, err := client.GetExpiringSoon(14)
	if err != nil {
		t.Fatal(err)
	}
	// Only "Expired" should be included (within 14 days from now means ≤ today+14)
	if len(items) != 1 {
		t.Errorf("len = %d, want 1; items = %+v", len(items), items)
	}
	if len(items) > 0 && items[0].ProductName != "Expired" {
		t.Errorf("expected %q, got %q", "Expired", items[0].ProductName)
	}
}

func TestGetUserfields(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]interface{}{"short_name": "Oats", "rating": 5})
	})
	fields, err := client.GetProductUserfields(1)
	if err != nil {
		t.Fatal(err)
	}
	if fields["short_name"] != "Oats" {
		t.Errorf("short_name = %q, want %q", fields["short_name"], "Oats")
	}
}

func TestCreateStore(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{"created_object_id": "7"})
	})
	store, err := client.CreateStore("Trader Joe's")
	if err != nil {
		t.Fatal(err)
	}
	if store.ID != 7 || store.Name != "Trader Joe's" {
		t.Errorf("unexpected store: %+v", store)
	}
}

func TestGetRecipes(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, []Recipe{{ID: 1, Name: "Pasta Bolognese"}, {ID: 2, Name: "Caesar Salad"}})
	})
	recipes, err := client.GetRecipes()
	if err != nil {
		t.Fatal(err)
	}
	if len(recipes) != 2 {
		t.Errorf("len = %d, want 2", len(recipes))
	}
}

func TestGetRecipeFulfillment(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, RecipeFulfillment{RecipeID: 1, NeedFulfilled: true})
	})
	f, err := client.GetRecipeFulfillment(1)
	if err != nil {
		t.Fatal(err)
	}
	if !f.NeedFulfilled {
		t.Error("expected NeedFulfilled = true")
	}
}

func TestCreateProduct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/products") {
				writeJSON(w, map[string]string{"created_object_id": "42"})
			} else {
				writeJSON(w, Product{ID: 42, Name: "New Oats"})
			}
		})
		shelf := 180
		p, err := client.CreateProduct("New Oats", &shelf, &Defaults{LocationID: 1, QuID: 1}, "", 1, 0, 1, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.ID != 42 {
			t.Errorf("product ID = %d, want 42", p.ID)
		}
	})

	t.Run("HTTP error", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "server error", http.StatusInternalServerError)
		})
		_, err := client.CreateProduct("Fail", nil, &Defaults{}, "", 0, 0, 0, nil, nil)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestUpdateProductName(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "expected PUT", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, map[string]string{})
	})
	if err := client.UpdateProductName(1, "New Name"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateProductDescription(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{})
	})
	if err := client.UpdateProductDescription(1, "A tasty snack"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateProductLocation(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{})
	})
	if err := client.UpdateProductLocation(1, 3); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateProductStore(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{})
	})
	if err := client.UpdateProductStore(1, 5); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSetUserfields(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{})
	})
	if err := client.SetUserfields("products", 1, map[string]string{"display_name": "Oats"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetMealPlan(t *testing.T) {
	items := []MealPlanItem{
		{ID: 1, Day: "2026-06-08", Note: "Breakfast"},
		{ID: 2, Day: "2026-06-09", Note: "Lunch"},
	}
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, items)
	})
	got, err := client.GetMealPlan("2026-06-08", "2026-06-14")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}
