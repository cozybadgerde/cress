package theme

import (
	"bytes"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

func TestLoadAndRender(t *testing.T) {
	fsys := fstest.MapFS{
		"templates/page.html": {Data: []byte(`<main>{{ .Page.Title }}</main>`)},
		"static/style.css":    {Data: []byte("body{}")},
	}

	thm, err := load("test", fsys)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if thm.Name() != "test" {
		t.Errorf("Name() = %q, want %q", thm.Name(), "test")
	}

	var buf bytes.Buffer
	if err := thm.Render(&buf, PageData{Page: PageView{Title: "hello"}}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if got := buf.String(); got != "<main>hello</main>" {
		t.Errorf("Render = %q", got)
	}

	if thm.StaticFS() == nil {
		t.Fatal("StaticFS() = nil, want the static tree")
	}
	if _, err := fs.Stat(thm.StaticFS(), "style.css"); err != nil {
		t.Errorf("static tree missing style.css: %v", err)
	}
}

func TestLoadWithoutStatic(t *testing.T) {
	fsys := fstest.MapFS{
		"templates/page.html": {Data: []byte(`<main></main>`)},
	}

	thm, err := load("nostatic", fsys)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if thm.StaticFS() != nil {
		t.Error("StaticFS() should be nil when the theme ships no static dir")
	}
}

func TestLoadMissingEntryTemplate(t *testing.T) {
	fsys := fstest.MapFS{
		"templates/partial.html": {Data: []byte(`{{ define "x" }}{{ end }}`)},
	}

	if _, err := load("bad", fsys); err == nil {
		t.Fatal("expected an error when page.html is missing")
	}
}

func TestLoadPartials(t *testing.T) {
	fsys := fstest.MapFS{
		"templates/page.html":              {Data: []byte(`<main>{{ template "greeting" . }}</main>`)},
		"templates/partials/greeting.html": {Data: []byte(`{{ define "greeting" }}hi {{ .Page.Title }}{{ end }}`)},
	}

	thm, err := load("partials", fsys)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	var buf bytes.Buffer
	if err := thm.Render(&buf, PageData{Page: PageView{Title: "there"}}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if got, want := buf.String(), "<main>hi there</main>"; got != want {
		t.Errorf("Render = %q, want %q", got, want)
	}
}

// A partial is reached by its {{ define }} name, so where it sits is the theme's
// business: nesting must work, and must mean nothing.
func TestLoadPartialsNested(t *testing.T) {
	fsys := fstest.MapFS{
		"templates/page.html":                  {Data: []byte(`{{ template "nav" . }}|{{ template "deep" . }}`)},
		"templates/partials/nav/main.html":     {Data: []byte(`{{ define "nav" }}NAV{{ end }}`)},
		"templates/partials/a/b/c/buried.html": {Data: []byte(`{{ define "deep" }}DEEP{{ end }}`)},
	}

	thm, err := load("nested", fsys)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	var buf bytes.Buffer
	if err := thm.Render(&buf, PageData{}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if got, want := buf.String(), "NAV|DEEP"; got != want {
		t.Errorf("Render = %q, want %q", got, want)
	}
}

// The separation this layout exists for: a partial renders when a layout calls
// it, and is invisible to a page asking for a layout by name.
func TestPartialIsNotALayout(t *testing.T) {
	fsys := fstest.MapFS{
		"templates/page.html":              {Data: []byte(`PAGE`)},
		"templates/404.html":               {Data: []byte(`NOTFOUND`)},
		"templates/partials/greeting.html": {Data: []byte(`{{ define "greeting" }}hi{{ end }}`)},
	}

	thm, err := load("layouts", fsys)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	for _, name := range []string{"page.html", "404.html"} {
		if !thm.HasLayout(name) {
			t.Errorf("HasLayout(%q) = false, want true for a file under templates/", name)
		}
	}
	// Neither the define's name nor the file it came from may be reached as a
	// layout, whether or not a page thinks to add the extension.
	for _, name := range []string{"greeting", "greeting.html", "partials/greeting.html", "templates/partials/greeting.html"} {
		if thm.HasLayout(name) {
			t.Errorf("HasLayout(%q) = true, want false for a partial", name)
		}
	}
}

// Layout names come from file names, so a {{ define }} inside a layout is a
// fragment that layout composes with rather than a layout a page may name.
func TestLayoutsAreFileNamesOnly(t *testing.T) {
	fsys := fstest.MapFS{
		"templates/page.html": {Data: []byte(`{{ define "sidebar" }}SIDE{{ end }}PAGE`)},
	}

	thm, err := load("defines", fsys)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !thm.HasLayout("page.html") {
		t.Error(`HasLayout("page.html") = false, want true`)
	}
	for _, name := range []string{"sidebar", "sidebar.html"} {
		if thm.HasLayout(name) {
			t.Errorf("HasLayout(%q) = true, want false for a define inside a layout", name)
		}
	}
}

// A define may not wear a layout's shape. One taking the name of a layout that
// exists would replace it, leaving the layout set still listing the name while
// every page rendered through the define instead, with nothing said about it.
func TestDefineMayNotEndInHTML(t *testing.T) {
	tests := []struct {
		name    string
		fsys    fstest.MapFS
		file    string
		defined string
	}{
		{
			name: "a partial taking an existing layout's name",
			fsys: fstest.MapFS{
				"templates/page.html":          {Data: []byte(`REAL`)},
				"templates/partials/bits.html": {Data: []byte(`{{ define "page.html" }}HIJACKED{{ end }}`)},
			},
			file:    "templates/partials/bits.html",
			defined: "page.html",
		},
		{
			// Rejected even though nothing collides yet, so that adding
			// sidebar.html later cannot quietly turn this into the case above.
			name: "a partial taking a name no layout has yet",
			fsys: fstest.MapFS{
				"templates/page.html":          {Data: []byte(`REAL`)},
				"templates/partials/bits.html": {Data: []byte(`{{ define "sidebar.html" }}X{{ end }}`)},
			},
			file:    "templates/partials/bits.html",
			defined: "sidebar.html",
		},
		{
			name: "a nested partial",
			fsys: fstest.MapFS{
				"templates/page.html":              {Data: []byte(`REAL`)},
				"templates/partials/a/b/deep.html": {Data: []byte(`{{ define "page.html" }}HIJACKED{{ end }}`)},
			},
			file:    "templates/partials/a/b/deep.html",
			defined: "page.html",
		},
		{
			// A partials directory is not required to reach the collision: one
			// layout can define another layout's name just as well.
			name: "a layout taking another layout's name",
			fsys: fstest.MapFS{
				"templates/page.html":    {Data: []byte(`REAL`)},
				"templates/landing.html": {Data: []byte(`{{ define "page.html" }}HIJACKED{{ end }}LANDING`)},
			},
			file:    "templates/landing.html",
			defined: "page.html",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := load("bad", tc.fsys)
			if err == nil {
				t.Fatal("expected an error for a define ending in .html")
			}
			// The theme author has to find both halves to fix it: the file that
			// holds the define, and the name it took.
			for _, want := range []string{tc.file, tc.defined} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not name %q", err, want)
				}
			}
		})
	}
}

// The shape a theme switch produces: a theme whose partial carries the name
// another theme uses for a layout, minus the extension. It is legal, it loads,
// and a page asking for that layout must still miss, because the lookup asks for
// the name with the extension and only files answer to those.
func TestPartialNamedLikeALayoutStem(t *testing.T) {
	fsys := fstest.MapFS{
		"templates/page.html":          {Data: []byte(`PAGE`)},
		"templates/partials/bits.html": {Data: []byte(`{{ define "landing" }}PARTIAL{{ end }}`)},
	}

	thm, err := load("third", fsys)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if thm.HasLayout("landing.html") {
		t.Error(`HasLayout("landing.html") = true, want false: "landing" is a partial, not a layout`)
	}
}

func TestLoadPartialsIgnoresNonHTML(t *testing.T) {
	fsys := fstest.MapFS{
		"templates/page.html":            {Data: []byte(`PAGE`)},
		"templates/partials/notes.txt":   {Data: []byte(`{{ this is not a template`)},
		"templates/partials/README.md":   {Data: []byte(`# how these are organised`)},
		"templates/partials/usable.html": {Data: []byte(`{{ define "usable" }}U{{ end }}`)},
	}

	if _, err := load("mixed", fsys); err != nil {
		t.Fatalf("load: %v, want non-template files under partials/ to be ignored", err)
	}
}

// The directory is optional, and template.ParseFS treats a pattern matching no
// files as an error, so a theme without partials must still load.
func TestLoadWithoutPartials(t *testing.T) {
	fsys := fstest.MapFS{
		"templates/page.html": {Data: []byte(`PAGE`)},
	}

	if _, err := load("nopartials", fsys); err != nil {
		t.Fatalf("load: %v", err)
	}
}

func TestLoadNoTemplates(t *testing.T) {
	fsys := fstest.MapFS{
		"static/style.css": {Data: []byte("body{}")},
	}

	_, err := load("empty", fsys)
	if err == nil {
		t.Fatal("expected an error when no templates match")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("error should name the theme: %v", err)
	}
}
