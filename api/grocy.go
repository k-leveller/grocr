package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kevin/grocy-scanner/config"
)

type GrocyClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewGrocyClient(cfg *config.Config) *GrocyClient {
	return &GrocyClient{
		baseURL:    strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:     cfg.APIKey,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *GrocyClient) request(method, path string, body interface{}, params map[string]string) ([]byte, error) {
	u := c.baseURL + "/api" + path
	if len(params) > 0 {
		v := url.Values{}
		for k, val := range params {
			v.Add(k, val)
		}
		u += "?" + v.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("GROCY-API-KEY", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *GrocyClient) FindProductByBarcode(upc string) (*Product, error) {
	data, err := c.request("GET", "/objects/product_barcodes", nil, map[string]string{
		"query[]": "barcode=" + upc,
	})
	if err != nil {
		return nil, err
	}

	var barcodes []ProductBarcode
	if err := json.Unmarshal(data, &barcodes); err != nil {
		return nil, err
	}
	if len(barcodes) == 0 {
		return nil, nil
	}

	productID := barcodes[0].ProductID
	return c.GetProduct(productID)
}

func (c *GrocyClient) GetProduct(id int) (*Product, error) {
	data, err := c.request("GET", fmt.Sprintf("/objects/products/%d", id), nil, nil)
	if err != nil {
		return nil, err
	}
	var p Product
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *GrocyClient) GetAllProducts() ([]Product, error) {
	data, err := c.request("GET", "/objects/products", nil, nil)
	if err != nil {
		return nil, err
	}
	var products []Product
	if err := json.Unmarshal(data, &products); err != nil {
		return nil, err
	}
	return products, nil
}

func (c *GrocyClient) ProductNameExists(name string) (bool, error) {
	data, err := c.request("GET", "/objects/products", nil, map[string]string{
		"query[]": "name=" + name,
	})
	if err != nil {
		return false, err
	}
	var products []Product
	if err := json.Unmarshal(data, &products); err != nil {
		return false, err
	}
	return len(products) > 0, nil
}

func (c *GrocyClient) GetLocations() ([]Location, error) {
	data, err := c.request("GET", "/objects/locations", nil, nil)
	if err != nil {
		return nil, err
	}
	var locs []Location
	if err := json.Unmarshal(data, &locs); err != nil {
		return nil, err
	}
	return locs, nil
}

func (c *GrocyClient) GetQuantityUnits() ([]QuantityUnit, error) {
	data, err := c.request("GET", "/objects/quantity_units", nil, nil)
	if err != nil {
		return nil, err
	}
	var units []QuantityUnit
	if err := json.Unmarshal(data, &units); err != nil {
		return nil, err
	}
	return units, nil
}

func (c *GrocyClient) GetDefaults() (*Defaults, error) {
	defaults := &Defaults{LocationID: 1, QuID: 1}

	locs, err := c.GetLocations()
	if err == nil && len(locs) > 0 {
		defaults.LocationID = locs[0].ID
		defaults.Locations = locs
	}

	units, err := c.GetQuantityUnits()
	if err == nil && len(units) > 0 {
		for _, u := range units {
			name := strings.ToLower(u.Name)
			if name == "piece" || name == "pcs" || name == "ea" || name == "each" {
				defaults.QuID = u.ID
				break
			}
		}
		if defaults.QuID == 1 && len(units) > 0 {
			defaults.QuID = units[0].ID
		}
	}

	return defaults, nil
}

func (c *GrocyClient) CreateProduct(name string, shelfLifeDays *int, defaults *Defaults, shortName string, locationID int, daysAfterFreezing, daysAfterThawing *int) (*Product, error) {
	locID := locationID
	if locID == 0 {
		locID = defaults.LocationID
	}

	data := map[string]interface{}{
		"name":           name,
		"location_id":    locID,
		"qu_id_purchase": defaults.QuID,
		"qu_id_stock":    defaults.QuID,
	}
	if shelfLifeDays != nil {
		data["default_best_before_days"] = *shelfLifeDays
	}
	if daysAfterFreezing != nil {
		data["default_best_before_days_after_freezing"] = *daysAfterFreezing
	}
	if daysAfterThawing != nil {
		data["default_best_before_days_after_thawing"] = *daysAfterThawing
	}

	respData, err := c.request("POST", "/objects/products", data, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		CreatedObjectID int `json:"created_object_id"`
	}
	if err := json.Unmarshal(respData, &result); err != nil {
		return nil, err
	}

	if result.CreatedObjectID > 0 {
		if shortName != "" {
			c.SetUserfields("products", result.CreatedObjectID, map[string]string{"short_name": shortName})
		}
		return c.GetProduct(result.CreatedObjectID)
	}

	return nil, fmt.Errorf("product creation returned no ID")
}

func (c *GrocyClient) UpdateProductName(productID int, newName string) error {
	data := map[string]interface{}{"name": newName}
	_, err := c.request("PUT", fmt.Sprintf("/objects/products/%d", productID), data, nil)
	return err
}

func (c *GrocyClient) AddBarcode(productID int, barcode string) error {
	data := map[string]interface{}{
		"product_id": productID,
		"barcode":    barcode,
	}
	_, err := c.request("POST", "/objects/product_barcodes", data, nil)
	return err
}

func (c *GrocyClient) AddStock(productID int, amount int, price float64, bestBefore string, locationID int) error {
	data := map[string]interface{}{
		"amount":           amount,
		"transaction_type": "purchase",
		"price":            price,
	}
	if bestBefore != "" {
		data["best_before_date"] = bestBefore
	}
	if locationID > 0 {
		data["location_id"] = locationID
	}

	_, err := c.request("POST", fmt.Sprintf("/stock/products/%d/add", productID), data, nil)
	return err
}

func (c *GrocyClient) ConsumeStock(productID int, amount int) error {
	data := map[string]interface{}{
		"amount":           amount,
		"transaction_type": "consume",
		"spoiled":          false,
	}
	_, err := c.request("POST", fmt.Sprintf("/stock/products/%d/consume", productID), data, nil)
	return err
}

func (c *GrocyClient) GetStock(productID int) (*StockInfo, error) {
	data, err := c.request("GET", fmt.Sprintf("/stock/products/%d", productID), nil, nil)
	if err != nil {
		return nil, err
	}
	var info StockInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (c *GrocyClient) SetUserfields(entity string, objectID int, fields map[string]string) error {
	_, err := c.request("PUT", fmt.Sprintf("/userfields/%s/%d", entity, objectID), fields, nil)
	return err
}
