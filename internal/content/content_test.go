package content_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/content"
)

func TestPrepare(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		title       string
		draft       bool
		wantError   string
		wantWarning []string
	}{
		{name: "title", source: "+++\ntitle = 'About'\n+++\nBody", title: "About"},
		{name: "default draft", source: "+++\ntitle = 'About'\ndraft = false\n+++", title: "About"},
		{name: "CRLF", source: "+++\r\ntitle = 'About'\r\n+++\r\nBody", title: "About"},
		{name: "closing delimiter at EOF", source: "+++\ntitle = 'About'\n+++", title: "About"},
		{name: "preserve title spacing", source: "+++\ntitle = ' About '\n+++", title: " About "},
		{name: "draft without title", source: "+++\ndraft = true\n+++\nSecret", draft: true},
		{name: "draft with title", source: "+++\ntitle = 'Private'\ndraft = true\n+++", title: "Private", draft: true},
		{name: "missing title", source: "Body", wantError: "published pages require a title"},
		{name: "empty frontmatter", source: "+++\n+++\nBody", wantError: "published pages require a title"},
		{name: "empty title", source: "+++\ntitle = ''\n+++", wantError: "title must be a nonempty string"},
		{name: "whitespace title", source: "+++\ntitle = ' \t\u0085\u00a0\u3000'\n+++", wantError: "title must be a nonempty string"},
		{name: "numeric title", source: "+++\ntitle = 123\n+++", wantError: "title must be a nonempty string"},
		{name: "table title", source: "+++\n[title]\nvalue = 'About'\n+++", wantError: "title must be a nonempty string"},
		{name: "string draft", source: "+++\ntitle = 'About'\ndraft = 'false'\n+++", wantError: "draft must be a boolean"},
		{name: "invalid draft title", source: "+++\ntitle = 123\ndraft = true\n+++", wantError: "title must be a nonempty string"},
		{name: "malformed TOML", source: "+++\ntitle = [\n+++", wantError: "parse frontmatter"},
		{name: "duplicate TOML key", source: "+++\ntitle = 'One'\ntitle = 'Two'\n+++", wantError: "parse frontmatter"},
		{name: "missing closing delimiter", source: "+++\ntitle = 'About'\nBody", wantError: "closing +++ delimiter"},
		{name: "opening delimiter only", source: "+++", wantError: "closing +++ delimiter"},
		{name: "invalid UTF-8", source: "+++\ntitle = 'About'\n+++\n\xff", wantError: "UTF-8"},
		{
			name: "unknown fields", source: "+++\ntitle = 'About'\nzebra = 1\nalpha = true\n+++", title: "About",
			wantWarning: []string{
				`content/about.md: unknown frontmatter field "alpha"`,
				`content/about.md: unknown frontmatter field "zebra"`,
			},
		},
		{
			name: "warning survives error", source: "+++\ntitle = 1\nextra = 1\n+++",
			wantError:   "title must be a nonempty string",
			wantWarning: []string{`content/about.md: unknown frontmatter field "extra"`},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			prepared, warnings, err := content.Prepare("about.md", []byte(test.source), defaultConfig(t))
			if !slices.Equal(warnings, test.wantWarning) {
				t.Fatalf("warnings = %q, want %q", warnings, test.wantWarning)
			}
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) ||
					!strings.Contains(err.Error(), "content/about.md") {
					t.Fatalf("error = %v, want source and %q", err, test.wantError)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if prepared.Data.Title != test.title || prepared.Draft != test.draft {
				t.Fatalf("prepared = %+v, want title %q and draft %v", prepared, test.title, test.draft)
			}
			if prepared.Draft && prepared.Data.Body != "" {
				t.Fatal("draft body was prepared for publication")
			}
		})
	}
}

func TestRoutes(t *testing.T) {
	tests := []struct {
		source      string
		destination string
		route       string
	}{
		{"index.md", "index.html", "/"},
		{"about.md", "about/index.html", "/about/"},
		{"about/index.md", "about/index.html", "/about/"},
		{"services/consulting.md", "services/consulting/index.html", "/services/consulting/"},
		{"services/index.md", "services/index.html", "/services/"},
		{"Our team.md", "Our team/index.html", "/Our%20team/"},
		{"é/50% #?.md", "é/50% #?/index.html", "/%C3%A9/50%25%20%23%3F/"},
		{"Index.md", "Index/index.html", "/Index/"},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			prepared, _, err := content.Prepare(test.source, []byte("+++\ntitle = 'Title'\n+++\nBody"), defaultConfig(t))
			if err != nil {
				t.Fatal(err)
			}
			if prepared.Destination != test.destination || prepared.Data.URL != test.route {
				t.Fatalf("destination = %q, route = %q; want %q and %q",
					prepared.Destination, prepared.Data.URL, test.destination, test.route)
			}
		})
	}
}

func TestRejectInvalidPaths(t *testing.T) {
	for _, name := range []string{"", ".", "../outside.md", "/absolute.md", "a/../b.md",
		"a//b.md", `a\b.md`, "a\x00.md", ".md", "...md", "about.MD", "file.txt", "\xff.md"} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := content.Prepare(name, []byte("+++\ntitle = 'Title'\n+++"), defaultConfig(t)); err == nil {
				t.Fatalf("accepted invalid source path %q", name)
			}
		})
	}
}

func TestMarkdownTrustAndTables(t *testing.T) {
	body := `
<script src="https://example.invalid/script.js">alert("trusted")</script>

<span onclick="run()">Inline</span>

[About](/about/) ![Alternative](/images/photo.jpg)

` + "`<b>literal</b>`" + `

` + "```unknown\n<script>literal()</script>\n```" + `

| Service | Price |
|---------|-------|
| Help    | 10    |
`
	prepared, _, err := content.Prepare("page.md", []byte("+++\ntitle = '<b>Title</b>'\n+++\n"+body), defaultConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	html := string(prepared.Data.Body)
	for _, want := range []string{
		`<script src="https://example.invalid/script.js">alert("trusted")</script>`,
		`<span onclick="run()">Inline</span>`,
		`href="/about/"`, `src="/images/photo.jpg"`, `alt="Alternative"`,
		`<code>&lt;b&gt;literal&lt;/b&gt;</code>`,
		`&lt;script&gt;literal()&lt;/script&gt;`, `<table>`, `<th>Service</th>`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered body does not contain %q:\n%s", want, html)
		}
	}
	if strings.Contains(html, "<script>literal()</script>") {
		t.Fatal("code example became active HTML")
	}
}
