package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type OFFClient struct {
	httpClient *http.Client
}

func NewOFFClient() *OFFClient {
	return &OFFClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *OFFClient) Lookup(upc string) (*OFFProduct, error) {
	url := fmt.Sprintf("https://world.openfoodfacts.org/api/v2/product/%s.json", upc)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "GrocyScanner/1.0 (barcode scanner TUI)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Status  int `json:"status"`
		Product struct {
			Name           string `json:"product_name"`
			Categories     string `json:"categories"`
			ExpirationDate string `json:"expiration_date"`
			ShelfLife       string `json:"shelf_life"`
		} `json:"product"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Status != 1 {
		return nil, nil
	}

	product := &OFFProduct{
		Name:       result.Product.Name,
		Categories: result.Product.Categories,
	}

	// Try to parse shelf life
	shelfStr := result.Product.ExpirationDate
	if shelfStr == "" {
		shelfStr = result.Product.ShelfLife
	}
	if shelfStr != "" {
		if days := parseShelfLife(shelfStr); days > 0 {
			product.ShelfLifeDays = &days
		}
	}

	// Fallback to category-based estimation
	if product.ShelfLifeDays == nil && product.Categories != "" {
		if days := EstimateShelfLife(product.Categories); days > 0 {
			product.ShelfLifeDays = &days
		}
	}

	return product, nil
}

var shelfLifeRe = regexp.MustCompile(`(\d+)\s*(day|month|year)`)

func parseShelfLife(s string) int {
	m := shelfLifeRe.FindStringSubmatch(strings.ToLower(s))
	if m == nil {
		return 0
	}
	val, _ := strconv.Atoi(m[1])
	switch {
	case strings.HasPrefix(m[2], "month"):
		return val * 30
	case strings.HasPrefix(m[2], "year"):
		return val * 365
	default:
		return val
	}
}
