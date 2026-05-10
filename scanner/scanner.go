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
// Very forgiving: strips non-digits, tries zero-padding, and accepts any
// 6-13 digit string as a plausible barcode (store-internal codes, etc.).
// Returns empty string only if there are no digits at all.
func CleanUPC(raw string) string {
	// Strip non-digit characters (spaces, dashes, etc.)
	var sb strings.Builder
	for _, c := range strings.TrimSpace(raw) {
		if c >= '0' && c <= '9' {
			sb.WriteRune(c)
		}
	}
	s := sb.String()
	if s == "" {
		return ""
	}

	// Try exact match at valid lengths first
	if validateCheckDigit(s) {
		return s
	}

	// Try zero-padding to standard lengths (scanners often strip leading zeros)
	for _, length := range []int{13, 12, 8} {
		if len(s) <= length {
			candidate := strings.Repeat("0", length-len(s)) + s
			if validateCheckDigit(candidate) {
				return candidate
			}
		}
	}

	// Accept any 6-13 digit string as a plausible barcode
	if len(s) >= 6 && len(s) <= 13 {
		return s
	}

	return ""
}

func validateCheckDigit(code string) bool {
	if len(code) != 8 && len(code) != 12 && len(code) != 13 {
		return false
	}

	digits := make([]int, len(code))
	for i, c := range code {
		digits[i] = int(c - '0')
	}

	n := len(digits)
	checkDigit := digits[n-1]

	var total int
	for i := 0; i < n-1; i++ {
		// Weight by position parity, length-aware (matches GS1 standard):
		// UPC-E (8): i even → ×1, i odd → ×3
		// UPC-A (12): i even → ×3, i odd → ×1
		// EAN-13 (13): i even → ×1, i odd → ×3
		// Unified: (n-1-i)%2 == 0 → ×1, else → ×3
		if (n-1-i)%2 == 0 {
			total += digits[i]
		} else {
			total += digits[i] * 3
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
