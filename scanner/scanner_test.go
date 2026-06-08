package scanner

import (
	"testing"
)

func TestCleanUPC(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty string", input: "", want: ""},
		{name: "no digits", input: "abc-def", want: ""},
		{name: "too long 14 digits", input: "12345678901234", want: ""},
		{name: "valid EAN-13", input: "4006381333931", want: "4006381333931"},
		{name: "valid UPC-A", input: "012345678905", want: "012345678905"},
		{name: "valid UPC-E 8 digit", input: "01234565", want: "01234565"},
		{name: "strips leading and trailing spaces", input: " 4006381333931 ", want: "4006381333931"},
		{name: "strips dashes", input: "4006-381-333-931", want: "4006381333931"},
		{name: "strips mixed non-digit chars", input: "400 638.133-3931", want: "4006381333931"},
		{name: "store internal 6 digit", input: "123456", want: "123456"},
		{name: "store internal 10 digit", input: "1234567890", want: "1234567890"},
		// Scanner strips leading zero; function pads back to EAN-13
		{name: "pad 11 digits to EAN-13", input: "12345678905", want: "0012345678905"},
		// Digits only but 5 chars - too short, no valid checksum padding
		{name: "5 digits too short", input: "12345", want: ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CleanUPC(tc.input)
			if got != tc.want {
				t.Errorf("CleanUPC(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestValidateCheckDigit(t *testing.T) {
	tests := []struct {
		name  string
		code  string
		valid bool
	}{
		{name: "valid EAN-13", code: "4006381333931", valid: true},
		{name: "invalid EAN-13 wrong check", code: "4006381333932", valid: false},
		{name: "valid UPC-A", code: "012345678905", valid: true},
		{name: "invalid UPC-A wrong check", code: "012345678904", valid: false},
		{name: "valid UPC-E", code: "01234565", valid: true},
		{name: "invalid UPC-E wrong check", code: "01234566", valid: false},
		{name: "wrong length 7", code: "1234567", valid: false},
		{name: "wrong length 11", code: "12345678901", valid: false},
		{name: "wrong length 14", code: "12345678901234", valid: false},
		{name: "empty string", code: "", valid: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := validateCheckDigit(tc.code)
			if got != tc.valid {
				t.Errorf("validateCheckDigit(%q) = %v, want %v", tc.code, got, tc.valid)
			}
		})
	}
}

func TestAdjustShelfLifeForFreezer(t *testing.T) {
	ptr := func(n int) *int { return &n }

	tests := []struct {
		name  string
		input *int
		want  int
	}{
		{name: "nil uses default", input: nil, want: FrozenDefaultDays},
		{name: "5 days maps to 365", input: ptr(5), want: 365},
		{name: "14 days maps to 365", input: ptr(14), want: 365},
		{name: "30 days maps to 365", input: ptr(30), want: 365},
		{name: "90 days maps to 365", input: ptr(90), want: 365},
		{name: "120 days maps to 270", input: ptr(120), want: 270},
		{name: "180 days maps to 270", input: ptr(180), want: 270},
		{name: "181 days returned as-is", input: ptr(181), want: 181},
		{name: "365 days returned as-is", input: ptr(365), want: 365},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := AdjustShelfLifeForFreezer(tc.input)
			if got != tc.want {
				t.Errorf("AdjustShelfLifeForFreezer(%v) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}
