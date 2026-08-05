package build

import "testing"

func TestLayoutTemplate(t *testing.T) {
	tests := []struct {
		name   string
		layout string
		want   string
	}{
		{"a plain name", "landing", "landing.html"},
		{"a hyphenated name", "full-bleed", "full-bleed.html"},
		// Layout names are free-form because the vocabulary is the theme's, so
		// nothing here is rejected or escaped. A name no theme defines simply
		// misses the template lookup and the page falls back, which is what
		// makes leaving it alone safe.
		{"a name with a path separator", "../secrets", "../secrets.html"},
		{"a name that already ends in .html", "page.html", "page.html.html"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := layoutTemplate(tc.layout); got != tc.want {
				t.Errorf("layoutTemplate(%q) = %q, want %q", tc.layout, got, tc.want)
			}
		})
	}
}

func TestMissingLayoutWarning(t *testing.T) {
	got := missingLayoutWarning("guides/setup.md", "wide", "cress")
	want := `guides/setup.md names layout "wide", which theme "cress" does not define`
	if got != want {
		t.Errorf("missingLayoutWarning() = %q, want %q", got, want)
	}
}
