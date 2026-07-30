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
