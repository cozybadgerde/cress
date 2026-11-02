package scaffold

import (
	"strings"
	"testing"
)

func TestSelectTheme(t *testing.T) {
	for _, test := range []struct {
		name   string
		source string
		want   string
		err    string
	}{
		{"single anchor", "# Keep comments\nname = \"poppy\"\n", "# Keep comments\nname = \"clover\"\n", ""},
		{"missing anchor", "name = 'poppy'\n", "", "found 0"},
		{"duplicate anchor", "name = \"poppy\"\nname = \"poppy\"\n", "", "found 2"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := selectTheme([]byte(test.source), "clover")
			if test.err != "" {
				if err == nil || !strings.Contains(err.Error(), "embedded cress.toml") ||
					!strings.Contains(err.Error(), test.err) {
					t.Fatalf("expected contextual anchor error, got %v", err)
				}
				return
			}
			if err != nil || string(got) != test.want {
				t.Fatalf("source = %q, error = %v; want %q", got, err, test.want)
			}
		})
	}
}
