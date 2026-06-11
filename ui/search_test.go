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
