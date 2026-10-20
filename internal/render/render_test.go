package render

import (
	"html/template"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/page"
)

func TestRender(t *testing.T) {
	renderer, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	output, err := renderer.Render(page.Page{
		Title: `<script>title()</script>`,
		URL:   "/about/",
		Body:  template.HTML("<p>Prepared body</p>"),
	})
	if err != nil {
		t.Fatal(err)
	}
	html := string(output)
	for _, want := range []string{
		"<!doctype html>", `<html lang="en">`, `<meta name="viewport"`,
		"<title>&lt;script&gt;title()&lt;/script&gt;</title>",
		"<main>", "<h1>&lt;script&gt;title()&lt;/script&gt;</h1>",
		"<p>Prepared body</p>", "</html>",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("output does not contain %q:\n%s", want, html)
		}
	}
	if strings.Contains(html, "<script>title()</script>") {
		t.Fatal("metadata escaped its text context")
	}
}

func TestRenderFailure(t *testing.T) {
	renderer := &Renderer{
		template: template.Must(template.New("page.gohtml").Parse("{{.Missing}}")),
	}
	if _, err := renderer.Render(page.Page{Title: "Title"}); err == nil {
		t.Fatal("template failure was hidden")
	}
}
