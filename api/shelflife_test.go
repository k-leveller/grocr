package api

import "testing"

func TestEstimateShelfLife(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{name: "empty string", input: "", want: 0},
		{name: "unknown category", input: "widgets and gadgets", want: 0},
		{name: "dairy", input: "dairy products", want: 14},
		{name: "milk", input: "whole milk", want: 10},
		{name: "yogurt", input: "Greek yogurt", want: 21},
		{name: "cheese", input: "cheddar cheese", want: 30},
		{name: "chicken", input: "boneless chicken breast", want: 5},
		{name: "seafood", input: "Atlantic seafood", want: 3},
		{name: "frozen", input: "frozen dinner", want: 180},
		{name: "canned", input: "canned tomatoes", want: 730},
		{name: "bread", input: "whole wheat bread", want: 7},
		{name: "pasta", input: "spaghetti pasta", want: 365},
		{name: "beverage", input: "energy beverage", want: 365},
		{name: "spice", input: "ground black spice", want: 730},
		{name: "case insensitive", input: "DAIRY cream", want: 14},
		{name: "egg", input: "large egg whites", want: 21},
		{name: "snack", input: "crunchy snack mix", want: 90},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := EstimateShelfLife(tc.input)
			if got != tc.want {
				t.Errorf("EstimateShelfLife(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}
