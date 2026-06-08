package api

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
	ProductID      int
	ProductName    string
	BestBeforeDate string
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
	ID   int    `json:"id"`
	Name string `json:"name"`
}
