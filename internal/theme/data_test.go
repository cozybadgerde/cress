package theme

import (
	"bytes"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/cozybadgerde/cress/internal/config"
)

// contractTemplate reaches for every field a theme may read. html/template
// fails execution on a field that does not exist, so this renders only as long
// as the contract holds: rename or remove one and the test says which.
const contractTemplate = `` +
	`site:{{ .Site.Title }}|{{ .Site.Description }}|{{ .Site.BaseURL }}|` +
	`{{ .Site.BasePath }}|{{ .Site.Language }}|{{ .Site.Theme }}|{{ .Site.Logo }}|` +
	`{{ .Site.LogoDark }}|{{ .Site.Favicon }}|{{ .Site.Accent }}|{{ .Site.AccentDark }}|` +
	`{{ .Site.Footer }}|{{ .Site.Copyright }}
` +
	`{{ range .Nav.Main }}main:{{ .Title }}|{{ .URL }}|{{ .Active }}
{{ end }}` +
	`{{ range .Nav.Footer }}footer:{{ .Title }}|{{ .URL }}|{{ .Active }}
{{ end }}` +
	`page:{{ .Page.Title }}|{{ .Page.Description }}|{{ .Page.Language }}|` +
	`{{ .Page.URL }}|{{ .Page.AbsoluteURL }}|{{ .Page.IsHome }}|` +
	`{{ index .Page.Meta "custom" }}
` +
	`{{ .Page.Head }}
` +
	`{{ .Page.HTML }}`

// fullPageData populates every field, so an assertion on the output can tell an
// empty value apart from a field the template failed to reach.
func fullPageData() PageData {
	return PageData{
		Site: config.Site{
			Title:       "Site title",
			Description: "Site description",
			BaseURL:     "https://example.com/project",
			BasePath:    "/project",
			Language:    "en-GB",
			Theme:       "cress",
			Logo:        "/logo.svg",
			LogoDark:    "/logo-dark.svg",
			Favicon:     "/favicon.png",
			Accent:      "#4f7a4a",
			AccentDark:  "#8fbf88",
			Footer:      "Footer text",
			Copyright:   "(c) 2026 Cozy Badger",
		},
		Nav: NavView{
			Main:   []NavLink{{Title: "Home", URL: "/", Active: true}},
			Footer: []NavLink{{Title: "Imprint", URL: "/imprint.html"}},
		},
		Page: PageView{
			Title:       "Page title",
			Description: "Page description",
			Language:    "de",
			URL:         "/project/about.html",
			AbsoluteURL: "https://example.com/project/about.html",
			IsHome:      false,
			Meta:        map[string]any{"custom": "meta value"},
			HTML:        "<p>Body <em>markup</em></p>",
			Head:        `<meta name="description" content="Page description" />`,
		},
	}
}

func TestPageDataContract(t *testing.T) {
	thm, err := load("contract", fstest.MapFS{
		"templates/page.html": {Data: []byte(contractTemplate)},
	})
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	var buf bytes.Buffer
	if err := thm.Render(&buf, fullPageData()); err != nil {
		t.Fatalf("Render: %v", err)
	}
	got := buf.String()

	// Every value a template can read must arrive, so a field silently lost in a
	// refactor fails here rather than at a user's next build.
	want := []string{
		"Site title", "Site description", "https://example.com/project", "/project",
		"en-GB", "cress", "/logo.svg", "/logo-dark.svg", "/favicon.png",
		"#4f7a4a", "#8fbf88", "Footer text", "(c) 2026 Cozy Badger",
		"main:Home|/|true",
		"footer:Imprint|/imprint.html|false",
		"page:Page title|Page description|de|/project/about.html|" +
			"https://example.com/project/about.html|false|meta value",
		// Head is markup, so it must arrive as tags rather than as escaped text.
		`<meta name="description" content="Page description" />`,
	}
	for _, w := range want {
		if !strings.Contains(got, w) {
			t.Errorf("output is missing %q:\n%s", w, got)
		}
	}
}

func TestPageHTMLIsNotEscapedAgain(t *testing.T) {
	// .Page.HTML is already-rendered markup. Escaping it a second time would
	// show readers their own page source, so its template.HTML type is
	// load-bearing rather than decorative.
	thm, err := load("contract", fstest.MapFS{
		"templates/page.html": {Data: []byte(`{{ .Page.HTML }}`)},
	})
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	var buf bytes.Buffer
	data := PageData{Page: PageView{HTML: "<p>Body <em>markup</em></p>"}}
	if err := thm.Render(&buf, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if got := buf.String(); got != "<p>Body <em>markup</em></p>" {
		t.Errorf("Render = %q, want the markup verbatim", got)
	}
}
