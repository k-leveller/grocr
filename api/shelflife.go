package api

import "strings"

// shelfLifeFallback maps keywords to estimated shelf life in days.
var shelfLifeFallback = map[string]int{
	"dairy":     14,
	"milk":      10,
	"yogurt":    21,
	"cheese":    30,
	"meat":      5,
	"bacon":     7,
	"sausage":   7,
	"ham":       7,
	"poultry":   5,
	"chicken":   5,
	"turkey":    5,
	"seafood":   3,
	"fish":      3,
	"shrimp":    3,
	"frozen":    180,
	"canned":    730,
	"produce":   7,
	"fruit":     7,
	"vegetable": 7,
	"bread":     7,
	"bakery":    5,
	"snack":     90,
	"chips":     90,
	"cereal":    180,
	"pasta":     365,
	"rice":      365,
	"beverage":  365,
	"juice":     14,
	"soda":      270,
	"water":     730,
	"condiment": 365,
	"sauce":     180,
	"spice":     730,
	"oil":       365,
	"egg":       21,
	"deli":      5,
}

// EstimateShelfLife estimates shelf life from text (product name or categories).
func EstimateShelfLife(text string) int {
	lower := strings.ToLower(text)
	for keyword, days := range shelfLifeFallback {
		if strings.Contains(lower, keyword) {
			return days
		}
	}
	return 0
}
