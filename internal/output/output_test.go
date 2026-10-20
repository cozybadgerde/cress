package output_test

import (
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/output"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		paths     []string
		wantError bool
	}{
		{name: "empty"},
		{name: "ownership marker", paths: []string{".cress-output.json"}, wantError: true},
		{name: "marker alias", paths: []string{".CRESS-OUTPUT.JSON"}, wantError: true},
		{name: "marker directory alias", paths: []string{".CRESS-OUTPUT.JSON/child"}, wantError: true},
		{name: "pages", paths: []string{"index.html", "about/index.html", "blog/first/index.html"}},
		{name: "duplicate", paths: []string{"about/index.html", "about/index.html"}, wantError: true},
		{name: "case collision", paths: []string{"About/index.html", "about/index.html"}, wantError: true},
		{name: "Unicode case collision", paths: []string{"Straße/index.html", "STRASSE/index.html"}, wantError: true},
		{name: "Unicode normalization collision", paths: []string{"café/index.html", "cafe\u0301/index.html"}, wantError: true},
		{name: "parent file first", paths: []string{"index.html", "index.html/index.html"}, wantError: true},
		{name: "parent file last", paths: []string{"index.html/index.html", "index.html"}, wantError: true},
		{name: "case parent collision", paths: []string{"INDEX.HTML", "index.html/index.html"}, wantError: true},
		{name: "shared directory", paths: []string{"blog/a/index.html", "blog/b/index.html"}},
		{name: "empty path", paths: []string{""}, wantError: true},
		{name: "dot", paths: []string{"."}, wantError: true},
		{name: "traversal", paths: []string{"../outside"}, wantError: true},
		{name: "nested traversal", paths: []string{"a/../outside"}, wantError: true},
		{name: "absolute", paths: []string{"/outside"}, wantError: true},
		{name: "backslash", paths: []string{`a\b`}, wantError: true},
		{name: "null byte", paths: []string{"a\x00b"}, wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var files []output.File
			for index, path := range test.paths {
				files = append(files, output.File{
					Path: path, Source: []string{"content/first.md", "content/second.md", "content/third.md"}[index],
				})
			}
			err := output.Validate(files)
			if (err != nil) != test.wantError {
				t.Fatalf("error = %v, want error %v", err, test.wantError)
			}
			if err != nil && len(files) == 2 {
				for _, file := range files {
					if !strings.Contains(err.Error(), file.Source) {
						t.Errorf("collision error does not name %s: %v", file.Source, err)
					}
				}
			}
		})
	}
}
