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

func TestSyntaxHighlightingStyles(t *testing.T) {
	renderer, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	for _, enabled := range []bool{true, false} {
		output, err := renderer.Render(page.Page{
			Title: "Code", Type: page.Ordinary,
			Site: page.Site{Syntax: page.Syntax{Highlight: enabled}},
		})
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(output), "pre .k,") != enabled {
			t.Fatalf("enabled=%v: skeleton ignored syntax setting: %s", enabled, output)
		}
		if !strings.Contains(string(output), "pre { padding:") {
			t.Fatal("ordinary code styling was removed")
		}
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
	for _, pageType := range []page.Type{page.Ordinary, page.Landing, page.List} {
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
	for _, invalidType := range []page.Type{"", "unknown"} {
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

func TestTitleImageCaptionEscaping(t *testing.T) {
	renderer, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	model := page.Page{
		Title: "Title", Type: page.Ordinary, Site: page.Site{Language: "en", URL: "/"},
		Image: &page.Image{
			URL: "/team.jpg", Alt: "<Team>",
			Caption: []page.CaptionPart{
				{Text: "<script>not HTML</script> by "},
				{Text: "<Alex>", URL: "/credits/?q=1&view=full"},
			},
		},
	}
	output, err := renderer.Render(model)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`<figure><img src="/team.jpg" alt="&lt;Team&gt;">`,
		`<figcaption>&lt;script&gt;not HTML&lt;/script&gt; by <a href="/credits/?q=1&amp;view=full">&lt;Alex&gt;</a></figcaption></figure>`,
	} {
		if !strings.Contains(string(output), want) {
			t.Errorf("missing %q in %s", want, output)
		}
	}
}

func TestPreparedListPresentation(t *testing.T) {
	renderer, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	model := page.Page{
		Title: "Journal", Type: page.List, Site: page.Site{Language: "en", URL: "/"},
		Canonical: "https://example.org/journal/", Body: template.HTML("<p>Introduction</p>"),
		Entries: []page.Entry{
			{Title: "<First>", URL: "/first/", PublishedAt: "2026-10-20", UpdatedAt: "2026-10-21", Excerpt: "<Excerpt>"},
			{Title: "Second", URL: "/second/", Description: "Not an excerpt"},
		},
	}
	output, err := renderer.Render(model)
	if err != nil {
		t.Fatal(err)
	}
	html := string(output)
	for _, want := range []string{
		`<link rel="canonical" href="https://example.org/journal/">`,
		`<h2><a href="/first/">&lt;First&gt;</a></h2>`,
		`<time datetime="2026-10-20">2026-10-20</time>`,
		`<time datetime="2026-10-21">2026-10-21</time>`, `<p>&lt;Excerpt&gt;</p>`,
		`<h2><a href="/second/">Second</a></h2>`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("missing %q in %s", want, html)
		}
	}
	if strings.Index(html, "<p>Introduction</p>") > strings.Index(html, "<ul>") ||
		strings.Index(html, "&lt;First&gt;</a>") > strings.Index(html, ">Second</a>") ||
		strings.Contains(html, "<Excerpt>") || strings.Contains(html, "Not an excerpt") {
		t.Fatalf("list metadata, intro or prepared order changed: %s", html)
	}
	model.Entries = nil
	output, err = renderer.Render(model)
	if err != nil || !strings.Contains(string(output), "<p>Introduction</p>") ||
		strings.Contains(string(output), "<ul>") {
		t.Fatalf("empty list lost intro or added empty markup: %s, %v", output, err)
	}
}
