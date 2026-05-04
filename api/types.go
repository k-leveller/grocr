package api

type Product struct {
	ID                              int    `json:"id"`
	Name                            string `json:"name"`
	LocationID                      int    `json:"location_id"`
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
	LocationID int
	QuID       int
	Locations  []Location
}

type OFFProduct struct {
	Name         string
	Categories   string
	ShelfLifeDays *int
}
