package theme

import (
	"strings"
	"testing"
	"testing/fstest"
)

// findingsFor validates a theme assembled from the given files, where each key
// is a path under the theme root, and returns what it found.
func findingsFor(t *testing.T, files map[string]string) []Finding {
	t.Helper()

	fsys := fstest.MapFS{}
	for path, body := range files {
		fsys[path] = &fstest.MapFile{Data: []byte(body)}
	}
	collected, findings := collectTemplates(fsys)
	if len(findings) != 0 {
		t.Fatalf("the theme did not parse: %v", findings)
	}
	return checkFields(collected)
}

// A field check that reports working templates is worse than none at all: a
// genuine unknown field already fails the build, so the only thing this can add
// is trust. Everything here must stay silent.
func TestCheckFieldsStaysSilent(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
	}{
		{
			name:  "plain fields",
			files: map[string]string{"templates/page.html": `{{ .Site.Title }}{{ .Page.Title }}{{ .Page.HTML }}`},
		},
		{
			name:  "the dot inside a range is the element",
			files: map[string]string{"templates/page.html": `{{ range .Nav.Main }}{{ .Title }}{{ .URL }}{{ .Active }}{{ end }}`},
		},
		{
			name:  "range with one variable binds the element",
			files: map[string]string{"templates/page.html": `{{ range $link := .Nav.Footer }}{{ $link.Title }}{{ end }}`},
		},
		{
			name:  "range with two variables binds index and element",
			files: map[string]string{"templates/page.html": `{{ range $i, $link := .Nav.Main }}{{ $i }}{{ $link.URL }}{{ end }}`},
		},
		{
			name:  "with enters the value",
			files: map[string]string{"templates/page.html": `{{ with .Page }}{{ .Title }}{{ .IsHome }}{{ end }}`},
		},
		{
			name:  "the else branch keeps the outer dot",
			files: map[string]string{"templates/page.html": `{{ with .Nav.Main }}{{ else }}{{ .Page.Title }}{{ end }}`},
		},
		{
			name:  "if does not enter the value",
			files: map[string]string{"templates/page.html": `{{ if .Page.IsHome }}{{ .Site.Title }}{{ end }}`},
		},
		{
			name:  "a variable carries its type",
			files: map[string]string{"templates/page.html": `{{ $p := .Page }}{{ $p.Title }}{{ $p.URL }}`},
		},
		{
			name:  "the root is reachable as $",
			files: map[string]string{"templates/page.html": `{{ range .Nav.Main }}{{ $.Site.Title }}{{ end }}`},
		},
		{
			// Meta is the documented escape hatch for front matter cress does not
			// define, so every key in it is correct by construction.
			name:  "any key under Page.Meta",
			files: map[string]string{"templates/page.html": `{{ .Page.Meta.subtitle }}{{ .Page.Meta.a.b.c }}`},
		},
		{
			name:  "a function call result is not tracked",
			files: map[string]string{"templates/page.html": `{{ (index .Nav.Main 0).Title }}{{ printf "%s" .Site.Title }}`},
		},
		{
			name: "a partial reached with the whole page data",
			files: map[string]string{
				"templates/page.html":           `{{ template "head" . }}`,
				"templates/partials/head.html":  `{{ define "head" }}{{ .Site.Title }}{{ .Page.Title }}{{ end }}`,
				"templates/partials/other.html": `{{ define "other" }}{{ end }}`,
			},
		},
		{
			// A fragment's dot is whatever its caller passed, and passing a part of
			// the page data is legal. Checking this against PageData would report
			// .Title, which is correct against PageView.
			name: "a partial reached with part of the page data",
			files: map[string]string{
				"templates/page.html":          `{{ template "byline" .Page }}`,
				"templates/partials/head.html": `{{ define "byline" }}{{ .Title }}{{ .Language }}{{ end }}`,
			},
		},
		{
			name: "a partial reached with a nav entry",
			files: map[string]string{
				"templates/page.html":          `{{ range .Nav.Main }}{{ template "link" . }}{{ end }}`,
				"templates/partials/link.html": `{{ define "link" }}{{ .Title }}{{ .URL }}{{ end }}`,
			},
		},
		{
			name: "a partial reached with two different types",
			files: map[string]string{
				"templates/page.html":          `{{ template "both" .Page }}{{ template "both" .Site }}`,
				"templates/partials/both.html": `{{ define "both" }}{{ .Title }}{{ .Anything }}{{ end }}`,
			},
		},
		{
			name: "a partial reached with no data at all",
			files: map[string]string{
				"templates/page.html":            `{{ template "static" }}`,
				"templates/partials/static.html": `{{ define "static" }}{{ .Whatever }}{{ end }}`,
			},
		},
		{
			name: "a partial nothing calls",
			files: map[string]string{
				"templates/page.html":            `hello`,
				"templates/partials/unused.html": `{{ define "unused" }}{{ .Nonsense }}{{ end }}`,
			},
		},
		{
			name: "a partial calling a partial",
			files: map[string]string{
				"templates/page.html":           `{{ template "outer" . }}`,
				"templates/partials/outer.html": `{{ define "outer" }}{{ template "inner" .Page }}{{ end }}`,
				"templates/partials/inner.html": `{{ define "inner" }}{{ .Title }}{{ end }}`,
			},
		},
		{
			// Nothing about a cycle is correct, but a validator that hangs on one is
			// worse than one that says nothing about it.
			name: "partials calling each other in a cycle",
			files: map[string]string{
				"templates/page.html":       `{{ template "a" . }}`,
				"templates/partials/a.html": `{{ define "a" }}{{ template "b" . }}{{ end }}`,
				"templates/partials/b.html": `{{ define "b" }}{{ template "a" . }}{{ .Site.Title }}{{ end }}`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if findings := findingsFor(t, tc.files); len(findings) != 0 {
				t.Errorf("reported a working template: %v", findings)
			}
		})
	}
}

func TestCheckFieldsReports(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{
			name:  "a misspelled page field",
			files: map[string]string{"templates/page.html": `{{ .Page.Titel }}`},
			want:  ".Page.Titel is not a field of PageView (did you mean Title?)",
		},
		{
			name:  "a misspelled site field",
			files: map[string]string{"templates/page.html": `{{ .Site.Titel }}`},
			want:  ".Site.Titel is not a field of Site (did you mean Title?)",
		},
		{
			name:  "a field of the top level",
			files: map[string]string{"templates/page.html": `{{ .Pages }}`},
			want:  ".Pages is not a field of PageData (did you mean Page?)",
		},
		{
			name:  "inside a range, against the element",
			files: map[string]string{"templates/page.html": `{{ range .Nav.Main }}{{ .Titel }}{{ end }}`},
			want:  ".Titel is not a field of NavLink (did you mean Title?)",
		},
		{
			name:  "inside a with, against the entered value",
			files: map[string]string{"templates/page.html": `{{ with .Page }}{{ .Titel }}{{ end }}`},
			want:  ".Titel is not a field of PageView (did you mean Title?)",
		},
		{
			name:  "through a variable",
			files: map[string]string{"templates/page.html": `{{ $p := .Page }}{{ $p.Titel }}`},
			want:  "$p.Titel is not a field of PageView (did you mean Title?)",
		},
		{
			name:  "through the root",
			files: map[string]string{"templates/page.html": `{{ range .Nav.Main }}{{ $.Site.Titel }}{{ end }}`},
			want:  "$.Site.Titel is not a field of Site (did you mean Title?)",
		},
		{
			// Main is a slice, so a field of it can never resolve however it is
			// spelled. The mistake is reaching for one entry without ranging.
			name:  "a field of a slice",
			files: map[string]string{"templates/page.html": `{{ .Nav.Main.Title }}`},
			want:  ".Nav.Main.Title reads a field of a list; range over .Nav.Main to reach each entry",
		},
		{
			name:  "a field with no near miss gets no suggestion",
			files: map[string]string{"templates/page.html": `{{ .Page.Pumpernickel }}`},
			want:  ".Page.Pumpernickel is not a field of PageView",
		},
		{
			name: "inside a partial reached with the whole page data",
			files: map[string]string{
				"templates/page.html":          `{{ template "head" . }}`,
				"templates/partials/head.html": `{{ define "head" }}{{ .Page.Titel }}{{ end }}`,
			},
			want: ".Page.Titel is not a field of PageView (did you mean Title?)",
		},
		{
			name: "inside a partial reached with part of the page data",
			files: map[string]string{
				"templates/page.html":            `{{ template "byline" .Page }}`,
				"templates/partials/byline.html": `{{ define "byline" }}{{ .Titel }}{{ end }}`,
			},
			want: ".Titel is not a field of PageView (did you mean Title?)",
		},
		{
			// The build never runs this: no page names the layout, so nothing
			// executes it. Finding it without content is the whole point.
			name: "inside a layout no page names",
			files: map[string]string{
				"templates/page.html": `ok`,
				"templates/wide.html": `{{ .Page.Titel }}`,
			},
			want: ".Page.Titel is not a field of PageView (did you mean Title?)",
		},
		{
			// Nor this: the branch is only taken on the home page.
			name:  "inside a branch no page takes",
			files: map[string]string{"templates/page.html": `{{ if .Page.IsHome }}{{ .Page.Titel }}{{ end }}`},
			want:  ".Page.Titel is not a field of PageView (did you mean Title?)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			findings := findingsFor(t, tc.files)
			if len(findings) != 1 {
				t.Fatalf("got %d findings, want 1: %v", len(findings), findings)
			}
			if findings[0].Message != tc.want {
				t.Errorf("message = %q, want %q", findings[0].Message, tc.want)
			}
		})
	}
}

// Only the first bad name in a chain is reported. Everything after it fails
// because of that one, and listing the consequences buries the cause.
func TestCheckFieldsReportsOnePerChain(t *testing.T) {
	findings := findingsFor(t, map[string]string{"templates/page.html": `{{ .Page.Titel.Deeper.Deepest }}`})
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1: %v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Message, ".Page.Titel is not") {
		t.Errorf("message = %q, want it to name the first bad step", findings[0].Message)
	}
}

func TestCheckFieldsReportsLineAndFile(t *testing.T) {
	findings := findingsFor(t, map[string]string{
		"templates/page.html":          "line one\n{{ template \"head\" . }}\n",
		"templates/partials/head.html": "{{ define \"head\" }}\n\n{{ .Page.Titel }}\n{{ end }}",
	})
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1: %v", len(findings), findings)
	}
	if findings[0].File != "templates/partials/head.html" {
		t.Errorf("File = %q, want the partial that holds the mistake", findings[0].File)
	}
	if findings[0].Line != 3 {
		t.Errorf("Line = %d, want 3", findings[0].Line)
	}
}

func TestDistance(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"title", "title", 0},
		{"title", "titel", 2},
		{"title", "titl", 1},
		{"title", "", 5},
		{"", "title", 5},
		{"abc", "xyz", 3},
	}
	for _, tc := range tests {
		t.Run(tc.a+"/"+tc.b, func(t *testing.T) {
			if got := distance(tc.a, tc.b); got != tc.want {
				t.Errorf("distance(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
