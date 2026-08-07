package config

import "testing"

func TestCleanBasePath(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"/", ""}, // a domain root, not a directory named ""
		{"/cress", "/cress"},
		{"cress", "/cress"},
		{"/cress/", "/cress"},
		{"cress/", "/cress"},
		{"/deep/nested/path", "/deep/nested/path"},
		{"/deep/nested/path/", "/deep/nested/path"},
		// Nothing else is trimmed: a path segment is whatever the URL said it was.
		{"/Mixed Case", "/Mixed Case"},
		{"//doubled", "/doubled"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := cleanBasePath(tc.in); got != tc.want {
				t.Errorf("cleanBasePath(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

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

func TestValidThemeName(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"cress", true},
		{"my-theme", true},
		{"theme.2", true},
		{"", false},
		{".", false},
		{"..", false},
		{"a/b", false},
		{"../elsewhere", false},
		{"../../etc/passwd", false},
		{"/absolute", false},
		{"themes/mine", false},
		{"trailing/", false},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := ValidThemeName(tc.in); got != tc.want {
				t.Errorf("ValidThemeName(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
