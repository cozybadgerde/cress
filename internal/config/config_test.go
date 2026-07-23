package config

import "testing"

func TestIsHexColor(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"#9cb43b", true},
		{"#FFF", true},
		{"#ffffff", true},
		{"#12345678", true},
		{"", false},
		{"9cb43b", false},   // missing #
		{"#12", false},      // wrong length
		{"#1234", false},    // wrong length
		{"#gggggg", false},  // non-hex digits
		{"#9cb43b;", false}, // trailing junk
		{"red", false},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := isHexColor(tc.in); got != tc.want {
				t.Errorf("isHexColor(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
