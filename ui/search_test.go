package ui

import (
	"testing"

	"github.com/k-leveller/grocr/api"
)

func TestSearchFilterSortsZeroStockLast(t *testing.T) {
	products := []api.Product{
		{ID: 1, Name: "apple juice"},
		{ID: 2, Name: "apple sauce"},
		{ID: 3, Name: "apple pie"},
		{ID: 4, Name: "apple butter"},
	}
	stock := map[int]float64{1: 0, 2: 3, 3: 0, 4: 1}

	s := NewSearch(products, nil, nil, stock, false)
	s.Input.SetValue("apple")
	s.UpdateFilter()

	if len(s.Filtered) != 4 {
		t.Fatalf("expected 4 results, got %d", len(s.Filtered))
	}
	// In-stock items keep their relative order, zero-stock items sink to the bottom.
	wantOrder := []int{2, 4, 1, 3}
	for i, want := range wantOrder {
		if s.Filtered[i].ID != want {
			t.Errorf("result %d: got product ID %d, want %d", i, s.Filtered[i].ID, want)
		}
	}
}

func TestSearchFilterNoStockAmounts(t *testing.T) {
	products := []api.Product{
		{ID: 1, Name: "apple juice"},
		{ID: 2, Name: "apple sauce"},
	}

	s := NewSearch(products, nil, nil, nil, false)
	s.Input.SetValue("apple")
	s.UpdateFilter()

	if len(s.Filtered) != 2 {
		t.Fatalf("expected 2 results, got %d", len(s.Filtered))
	}
	if s.Filtered[0].ID != 1 || s.Filtered[1].ID != 2 {
		t.Errorf("expected original order preserved, got %v", []int{s.Filtered[0].ID, s.Filtered[1].ID})
	}
}

func TestUpdateFilterKeepsHighlightOnSameProduct(t *testing.T) {
	products := []api.Product{
		{ID: 1, Name: "apple juice"},
		{ID: 2, Name: "apple sauce"},
	}

	// apple juice is out of stock, so it sorts below apple sauce.
	s := NewSearch(products, nil, nil, map[int]float64{1: 0, 2: 3}, false)
	s.Input.SetValue("apple")
	s.UpdateFilter()
	s.Cursor = 1 // highlighting "apple juice"
	if s.Filtered[s.Cursor].ID != 1 {
		t.Fatalf("setup: highlighted product %d, want 1", s.Filtered[s.Cursor].ID)
	}

	// A refresh restocks apple juice and empties apple sauce, swapping the rows.
	s.StockAmounts = map[int]float64{1: 5, 2: 0}
	s.UpdateFilter()

	if s.Filtered[s.Cursor].ID != 1 {
		t.Errorf("highlight moved to product %d, want 1 (apple juice)", s.Filtered[s.Cursor].ID)
	}
}

func TestUpdateFilterResetsHighlightWhenProductGone(t *testing.T) {
	products := []api.Product{
		{ID: 1, Name: "apple juice"},
		{ID: 2, Name: "apple sauce"},
	}

	s := NewSearch(products, nil, nil, nil, false)
	s.Input.SetValue("apple")
	s.UpdateFilter()
	s.Cursor = 1

	// apple sauce was deleted in Grocy.
	s.Products = []api.Product{{ID: 1, Name: "apple juice"}}
	s.UpdateFilter()

	if s.Cursor != 0 {
		t.Errorf("Cursor = %d, want 0 after the highlighted product disappeared", s.Cursor)
	}
}

func TestUpdateFilterDropsUnknownLocationFilter(t *testing.T) {
	products := []api.Product{{ID: 1, Name: "apple juice", LocationID: 3}}

	s := NewSearch(products, []api.Location{{ID: 9, Name: "Cellar"}}, nil, nil, false)
	s.LocFilter = 4 // a location that no longer exists after a refresh
	s.Input.SetValue("apple")
	s.UpdateFilter()

	if s.LocFilter != 0 {
		t.Errorf("LocFilter = %d, want 0 once the location is gone", s.LocFilter)
	}
	if len(s.Filtered) != 1 {
		t.Errorf("expected the product to be visible again, got %d results", len(s.Filtered))
	}
}

func TestUpdateFilterKeepsLocationFilterWhenLocationsUnavailable(t *testing.T) {
	products := []api.Product{
		{ID: 1, Name: "apple juice", LocationID: 3},
		{ID: 2, Name: "apple sauce", LocationID: 4},
	}

	s := NewSearch(products, nil, nil, nil, false)
	s.LocFilter = 3
	s.Input.SetValue("apple")
	s.UpdateFilter()

	if s.LocFilter != 3 {
		t.Errorf("LocFilter = %d, want 3 kept when the location list is unavailable", s.LocFilter)
	}
	if len(s.Filtered) != 1 || s.Filtered[0].ID != 1 {
		t.Errorf("unexpected results: %+v", s.Filtered)
	}
}
