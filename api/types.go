package api

import (
	"encoding/json"
	"strconv"
)

type Product struct {
	ID                              int    `json:"id"`
	Name                            string `json:"name"`
	Description                     string `json:"description"`
	LocationID                      int    `json:"location_id"`
	ShoppingLocationID              int    `json:"shopping_location_id"`
	QuIDPurchase                    int    `json:"qu_id_purchase"`
	QuIDStock                       int    `json:"qu_id_stock"`
	DefaultBestBeforeDays           int    `json:"default_best_before_days"`
	DefaultBestBeforeDaysAfterFreez int    `json:"default_best_before_days_after_freezing"`
	DefaultBestBeforeDaysAfterThaw  int    `json:"default_best_before_days_after_thawing"`
}

type Location struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Store struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type QuantityUnit struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProductBarcode struct {
	ID        int    `json:"id"`
	ProductID int    `json:"product_id"`
	Barcode   string `json:"barcode"`
}

type StockInfo struct {
	StockAmount float64 `json:"stock_amount"`
	LastPrice   float64 `json:"last_price"`
}

type Defaults struct {
	LocationID    int
	QuID          int
	Locations     []Location
	Stores        []Store
	QuantityUnits []QuantityUnit
}

type OFFProduct struct {
	Name          string
	Categories    string
	ShelfLifeDays *int
}

type ExpiringItem struct {
	ProductID        int
	ProductName      string
	ProductShortName string
	BestBeforeDate   string
}

type StockEntry struct {
	ProductName    string
	Amount         float64
	LocationID     int
	BestBeforeDate string
}

type StockTransaction struct {
	ID          int     `json:"id"`
	ProductID   int     `json:"product_id"`
	Date        string  `json:"row_created_timestamp"`
	Type        string  `json:"transaction_type"`
	Amount      float64 `json:"amount"`
	Price       float64 `json:"price"`
	ShoppingLocationID int `json:"shopping_location_id"`
}

type MealPlanItem struct {
	ID             int     `json:"id"`
	Day            string  `json:"day"`
	RecipeID       *int    `json:"recipe_id"`
	RecipeServings float64 `json:"recipe_servings"`
	Note           string  `json:"note"`
	ProductID      *int    `json:"product_id"`
	ProductAmount  float64 `json:"product_amount"`
}

type Recipe struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	BaseServings int    `json:"base_servings"`
}

type RecipeFulfillment struct {
	RecipeID                      int  `json:"recipe_id"`
	NeedFulfilled                 bool `json:"need_fulfilled"`
	NeedFulfilledWithShoppingList bool `json:"need_fulfilled_with_shopping_list"`
	MissingProductsCount          int  `json:"missing_products_count"`
}

// UnmarshalJSON handles Grocy returning numeric/bool fields as JSON strings
// (e.g. "need_fulfilled":"0") instead of native types.
func (r *RecipeFulfillment) UnmarshalJSON(data []byte) error {
	var raw struct {
		RecipeID                      json.RawMessage `json:"recipe_id"`
		NeedFulfilled                 json.RawMessage `json:"need_fulfilled"`
		NeedFulfilledWithShoppingList json.RawMessage `json:"need_fulfilled_with_shopping_list"`
		MissingProductsCount          json.RawMessage `json:"missing_products_count"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.RecipeID = parseFlexInt(raw.RecipeID)
	r.NeedFulfilled = parseFlexBool(raw.NeedFulfilled)
	r.NeedFulfilledWithShoppingList = parseFlexBool(raw.NeedFulfilledWithShoppingList)
	r.MissingProductsCount = parseFlexInt(raw.MissingProductsCount)
	return nil
}

// parseFlexBool decodes a JSON value that may be a native bool, an integer
// 0/1, or a quoted string "0"/"1"/"true"/"false".
func parseFlexBool(data json.RawMessage) bool {
	if len(data) == 0 {
		return false
	}
	var b bool
	if json.Unmarshal(data, &b) == nil {
		return b
	}
	var n int
	if json.Unmarshal(data, &n) == nil {
		return n != 0
	}
	var s string
	if json.Unmarshal(data, &s) == nil {
		return s == "1" || s == "true"
	}
	return false
}

// parseFlexInt decodes a JSON value that may be a native integer or a quoted
// string like "2".
func parseFlexInt(data json.RawMessage) int {
	if len(data) == 0 {
		return 0
	}
	var n int
	if json.Unmarshal(data, &n) == nil {
		return n
	}
	var s string
	if json.Unmarshal(data, &s) == nil {
		v, _ := strconv.Atoi(s)
		return v
	}
	return 0
}
