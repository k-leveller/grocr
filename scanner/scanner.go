package scanner

import (
	"strings"
)

// FrozenShelfLife maps fridge shelf life thresholds to freezer shelf life.
var FrozenShelfLife = []struct {
	Threshold int
	Days      int
}{
	{14, 365},
	{90, 365},
	{180, 270},
}

const FrozenDefaultDays = 180

// CleanUPC validates and normalizes a UPC code.
// Returns empty string if invalid.
func CleanUPC(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "0x")

	for _, c := range s {
		if c < '0' || c > '9' {
			return ""
		}
	}

	switch len(s) {
	case 8, 12, 13:
		if validateCheckDigit(s) {
			return s
		}
		return ""
	default:
		return ""
	}
}

func validateCheckDigit(code string) bool {
	digits := make([]int, len(code))
	for i, c := range code {
		digits[i] = int(c - '0')
	}

	n := len(digits)
	checkDigit := digits[n-1]

	var total int
	for i := 0; i < n-1; i++ {
		if i%2 == 0 {
			if n == 8 {
				total += digits[i] * 3
			} else {
				total += digits[i]
			}
		} else {
			if n == 8 {
				total += digits[i]
			} else {
				total += digits[i] * 3
			}
		}
	}

	expected := (10 - total%10) % 10
	return expected == checkDigit
}

// AdjustShelfLifeForFreezer adjusts shelf life for frozen storage.
func AdjustShelfLifeForFreezer(shelfLifeDays *int) int {
	if shelfLifeDays == nil {
		return FrozenDefaultDays
	}
	days := *shelfLifeDays
	if days > 180 {
		return days
	}
	for _, entry := range FrozenShelfLife {
		if days <= entry.Threshold {
			return entry.Days
		}
	}
	return days
}
