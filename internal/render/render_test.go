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
		Type:  page.Ordinary,
		Site:  page.Site{Language: "en", URL: "/"},
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
	if _, err := renderer.Render(page.Page{Title: "Title", Type: page.Ordinary}); err == nil {
		t.Fatal("template failure was hidden")
	}
}

func TestPreparedMetadataAndPageTypes(t *testing.T) {
	renderer, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	model := page.Page{
		Title: "<Title>", Description: `A "description"`, Type: page.Ordinary,
		Image: &page.Image{URL: "/company/photo.jpg", Alt: "<Alternative>"},
		Body:  template.HTML("<p>Trusted body</p>"),
		Site: page.Site{
			Title: "<Business>", Copyright: "<Copyright>", Language: "de-CH", URL: "/company/",
			Navigation: page.Navigation{
				Main:   []page.Link{{Label: "<First>", URL: "/company/first/"}, {Label: "Second", URL: "https://example.net/"}},
				Footer: []page.Link{{Label: "<Footer>", URL: "/company/contact/"}},
			},
			Custom: map[string]any{"raw": "<script>custom()</script>"},
		},
	}
	for _, pageType := range []page.Type{page.Ordinary, page.Landing} {
		t.Run(string(pageType), func(t *testing.T) {
			model.Type = pageType
			output, err := renderer.Render(model)
			if err != nil {
				t.Fatal(err)
			}
			html := string(output)
			for _, expected := range []string{
				`<html lang="de-CH">`, `content="A &#34;description&#34;"`,
				`<img src="/company/photo.jpg" alt="&lt;Alternative&gt;">`,
				`href="/company/">&lt;Business&gt;</a>`, `>&lt;First&gt;</a>`,
				`href="https://example.net/"`, `>&lt;Footer&gt;</a>`,
				`<p>&lt;Copyright&gt;</p>`, `<p>Trusted body</p>`,
			} {
				if !strings.Contains(html, expected) {
					t.Errorf("missing %q:\n%s", expected, html)
				}
			}
			if (pageType == page.Landing) != strings.Contains(html, "<section>") {
				t.Errorf("page type did not select its own template:\n%s", html)
			}
			if strings.Count(html, "<html") != 1 || strings.Count(html, "<main>") != 1 ||
				strings.Contains(html, "<script>") {
				t.Errorf("invalid composition or active ordinary values:\n%s", html)
			}
		})
	}
	for _, invalidType := range []page.Type{"", "unknown", page.List} {
		model.Type = invalidType
		if _, err := renderer.Render(model); err == nil {
			t.Fatalf("unsupported type %q rendered via fallback", invalidType)
		}
	}
}

func TestEmptyOptionalMarkup(t *testing.T) {
	renderer, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	output, err := renderer.Render(page.Page{
		Title: "Title", Type: page.Ordinary, Site: page.Site{Language: "en", URL: "/"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, unwanted := range []string{"<header>", "<footer>", "<nav", "<img", `name="description"`} {
		if strings.Contains(string(output), unwanted) {
			t.Errorf("empty optional markup %q was rendered", unwanted)
		}
	}
}
