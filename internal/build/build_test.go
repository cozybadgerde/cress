package build

import "testing"

func TestExpandYear(t *testing.T) {
	tests := []struct {
		name      string
		copyright string
		want      string
	}{
		{"empty stays empty", "", ""},
		{"no token is untouched", "(c) 2019 Cozy Badger", "(c) 2019 Cozy Badger"},
		{"token expands", "(c) {year} Cozy Badger", "(c) 2026 Cozy Badger"},
		{"every occurrence expands", "{year}-{year}", "2026-2026"},
		{"a range keeps its literal start", "(c) 2019-{year} Cozy Badger", "(c) 2019-2026 Cozy Badger"},
		{"the token alone is the whole notice", "{year}", "2026"},
		// Only the exact token is recognised; anything near it is the author's text.
		{"a near miss is left alone", "(c) {Year} {years} { year }", "(c) {Year} {years} { year }"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := expandYear(tc.copyright, 2026); got != tc.want {
				t.Errorf("expandYear(%q, 2026) = %q, want %q", tc.copyright, got, tc.want)
			}
		})
	}
}
