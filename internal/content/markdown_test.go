package content_test

import (
	"html"
	"regexp"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/content"
)

func prepareBody(t *testing.T, body string) (string, error) {
	t.Helper()
	prepared, _, err := content.Prepare("page.md", []byte("+++\ntitle = 'Page'\n+++\n"+body), defaultConfig(t))
	return string(prepared.Data.Body), err
}

func TestStandaloneLinkParagraphs(t *testing.T) {
	tests := []struct {
		name, body, want string
	}{
		{"inline", "[Services](/services/)", `<p class="standalone-link"><a href="/services/">Services</a></p>`},
		{"formatted label", "[**Our** services](/services/)", `<p class="standalone-link"><a href="/services/"><strong>Our</strong> services</a></p>`},
		{"reference", "[Services][services]\n\n[services]: /services/", `<p class="standalone-link"><a href="/services/">Services</a></p>`},
		{"autolink", "<https://example.org/>", `<p class="standalone-link"><a href="https://example.org/">https://example.org/</a></p>`},
		{"before text", "See [services](/services/)", `<p>See <a href="/services/">services</a></p>`},
		{"after text", "[Services](/services/) first.", `<p><a href="/services/">Services</a> first.</p>`},
		{"wrapped link", "**[Services](/services/)**", `<p><strong><a href="/services/">Services</a></strong></p>`},
		{"two links", "[One](/one/) [Two](/two/)", `<p><a href="/one/">One</a> <a href="/two/">Two</a></p>`},
		{"tight list", "- [Services](/services/)", "<ul>\n<li><a href=\"/services/\">Services</a></li>\n</ul>"},
		{"blockquote", "> [Services](/services/)", "<blockquote>\n<p class=\"standalone-link\"><a href=\"/services/\">Services</a></p>\n</blockquote>"},
		{"code", "`[Services](/services/)`", "<p><code>[Services](/services/)</code></p>"},
		{"HTML paragraph", `<p><a href="/services/">Services</a></p>`, `<p><a href="/services/">Services</a></p>`},
		{"HTML link", `<a href="/services/">Services</a>`, `<p><a href="/services/">Services</a></p>`},
		{"hard break", "[Services](/services/)\\\nNext line", "<p><a href=\"/services/\">Services</a><br>\nNext line</p>"},
		{"entity text", "[Services](/services/)&nbsp;", "<p><a href=\"/services/\">Services</a>\u00a0</p>"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rendered, err := prepareBody(t, test.body)
			if err != nil {
				t.Fatal(err)
			}
			if strings.TrimSpace(rendered) != test.want {
				t.Errorf("rendered %q, want %q", rendered, test.want)
			}
		})
	}
}

func TestMarkdownStrikethrough(t *testing.T) {
	tests := []struct {
		name, body string
		want       []string
		absent     []string
	}{
		{
			name: "inline formatting and resolved links",
			body: "~~Old **wording** and [services](/services/)~~; ~previous~.",
			want: []string{`<del>Old <strong>wording</strong> and <a href="/services/">services</a></del>`,
				"<del>previous</del>"},
		},
		{
			name: "table cell and heading",
			body: "## ~~Old~~ title\n\n| Status |\n| --- |\n| ~~Previous~~ |\n",
			want: []string{`<h2 id="old-title"><del>Old</del> title</h2>`, "<td><del>Previous</del></td>"},
		},
		{
			name: "literal boundaries and no linkification",
			body: "`~~literal~~`\n\n```\n~~literal~~\n```\n\n<div>~~HTML~~</div>\n\n\\~\\~escaped\\~\\~\n\nhttps://example.org/\n\n~~~unmatched~~~",
			want: []string{"<code>~~literal~~</code>", "<pre><code>~~literal~~\n</code></pre>",
				"<div>~~HTML~~</div>", "<p>~~escaped~~</p>", "<p>https://example.org/</p>"},
			absent: []string{"<del>", "<a "},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rendered, err := prepareBody(t, test.body)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range test.want {
				if !strings.Contains(rendered, want) {
					t.Errorf("missing %q in %s", want, rendered)
				}
			}
			for _, absent := range test.absent {
				if strings.Contains(rendered, absent) {
					t.Errorf("unexpected %q in %s", absent, rendered)
				}
			}
		})
	}
}

func TestMarkdownTaskLists(t *testing.T) {
	tests := []struct {
		name, body string
		want       []string
		labels     int
	}{
		{
			name: "checked markers and ordinary items",
			body: "- [ ] Pending\n- [x] Complete\n- [X] Also complete\n- Ordinary\n",
			want: []string{`<label><input disabled="" type="checkbox"> Pending</label>`,
				`<label><input checked="" disabled="" type="checkbox"> Complete</label>`,
				`<label><input checked="" disabled="" type="checkbox"> Also complete</label>`,
				"<li>Ordinary</li>"},
			labels: 3,
		},
		{
			name: "ordered nested and inline content",
			body: "1. [x] ~~Old~~ **scope**\n   - [ ] Read [services](/services/)\n2. Ordinary\n",
			want: []string{"<ol>", `<label><input checked="" disabled="" type="checkbox"> <del>Old</del> <strong>scope</strong></label>`,
				"<ul>", `<label><input disabled="" type="checkbox"> Read <a href="/services/">services</a></label>`,
				"</label>\n<ul>", "<li>Ordinary</li>"},
			labels: 2,
		},
		{
			name: "loose paragraphs",
			body: "- [ ] First paragraph\n\n  More detail with [contact](/contact/).\n\n  - [x] Nested\n",
			want: []string{`<p><label><input disabled="" type="checkbox"> First paragraph</label></p>`,
				`<p>More detail with <a href="/contact/">contact</a>.</p>`,
				`<label><input checked="" disabled="" type="checkbox"> Nested</label>`},
			labels: 2,
		},
		{
			name: "captioned image",
			body: `- [ ] Review ![Team](/team.jpg){caption="Photo credit"}`,
			want: []string{`<label><input disabled="" type="checkbox"> Review <span class="captioned-image"`,
				`<span id="cress-caption-1" class="image-caption">Photo credit</span></span></label>`},
			labels: 1,
		},
		{
			name: "literal boundaries and non-task markers",
			body: "[ ] Not a list\n\n- [y] Not a task\n- Later [x] marker\n\n`- [ ] Code`\n\n```\n- [x] Code\n```\n\n<ul><li>[x] HTML</li></ul>",
			want: []string{"<p>[ ] Not a list</p>", "<li>[y] Not a task</li>", "<li>Later [x] marker</li>",
				"<code>- [ ] Code</code>", "<pre><code>- [x] Code\n</code></pre>", "<ul><li>[x] HTML</li></ul>"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rendered, err := prepareBody(t, test.body)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range test.want {
				if !strings.Contains(rendered, want) {
					t.Errorf("missing %q in %s", want, rendered)
				}
			}
			if strings.Count(rendered, "<label>") != test.labels || strings.Count(rendered, "</label>") != test.labels ||
				strings.Count(rendered, `disabled="" type="checkbox"`) != test.labels {
				t.Errorf("want %d labeled disabled checkboxes in %s", test.labels, rendered)
			}
		})
	}
}

func TestMarkdownHeadingTargets(t *testing.T) {
	tests := []struct {
		name, body string
		want       []string
	}{
		{
			name: "heading syntaxes and levels",
			body: "# One\n\n## Two ##\n\n### Three\n\n#### Four\n\n##### Five\n\n###### Six\n\nUnderlined\n==========\n\nSection\n-------\n\n> ## Nested\n\n##\n",
			want: []string{`<h1 id="one">One</h1>`, `<h2 id="two">Two</h2>`,
				`<h3 id="three">Three</h3>`, `<h4 id="four">Four</h4>`,
				`<h5 id="five">Five</h5>`, `<h6 id="six">Six</h6>`,
				`<h1 id="underlined">Underlined</h1>`, `<h2 id="section">Section</h2>`,
				`<h2 id="nested">Nested</h2>`, `<h2 id="heading"></h2>`},
		},
		{
			name: "explicit IDs have priority",
			body: "## Services\n\n## Services\n\n## Stable {#services}\n\n## Suffix {#services-1}\n",
			want: []string{`<h2 id="services-2">Services</h2>`, `<h2 id="services-3">Services</h2>`,
				`<h2 id="services">Stable</h2>`, `<h2 id="services-1">Suffix</h2>`},
		},
		{
			name: "explicit setext and case",
			body: "Stable {#Services}\n------\n\n## Services\n\n## Services\n",
			want: []string{`<h2 id="Services">Stable</h2>`, `<h2 id="services">Services</h2>`,
				`<h2 id="services-1">Services</h2>`},
		},
		{
			name: "non-ASCII and empty slugs",
			body: "## Über uns\n\n## 日本語\n\n## !!!\n\n## Visit {#öffnungszeiten}\n",
			want: []string{`<h2 id="ber-uns">Über uns</h2>`, `<h2 id="heading">日本語</h2>`,
				`<h2 id="heading-1">!!!</h2>`, `<h2 id="öffnungszeiten">Visit</h2>`},
		},
		{
			name: "native source-based slugging",
			body: "## Our **team** & services_now!\n\nFirst line\nLast line\n---------\n",
			want: []string{`<h2 id="our-team--services-now">Our <strong>team</strong> &amp; services_now!</h2>`,
				"<h2 id=\"last-line\">First line\nLast line</h2>"},
		},
		{
			name: "reserved automatic targets and caption",
			body: "## Main\n\n## Page title\n\n## Cress caption 1\n\n## Cress caption 1\n\n## Stable {#heading-cress-caption-1}\n\n" +
				`Before ![Team](/team.jpg){caption="Credit"} after.`,
			want: []string{`<h2 id="main-1">Main</h2>`, `<h2 id="page-title-1">Page title</h2>`,
				`<h2 id="heading-cress-caption-1-1">Cress caption 1</h2>`,
				`<h2 id="heading-cress-caption-1-2">Cress caption 1</h2>`,
				`<h2 id="heading-cress-caption-1">Stable</h2>`,
				`aria-describedby="cress-caption-1"`, `<span id="cress-caption-1" class="image-caption">Credit</span>`},
		},
		{
			name: "raw HTML and code stay opaque",
			body: "<h2 id=\"main\">Authored</h2>\n\n## Main\n\n```\n## Literal {#main}\n```\n\n`## Literal {#main}`",
			want: []string{`<h2 id="main">Authored</h2>`, `<h2 id="main-1">Main</h2>`,
				"## Literal {#main}\n</code>", `<code>## Literal {#main}</code>`},
		},
		{
			name: "explicit ID is escaped",
			body: `## Stable {id="quote\"<&"}`,
			want: []string{`<h2 id="quote&quot;&lt;&amp;">Stable</h2>`},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rendered, err := prepareBody(t, test.body)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range test.want {
				if !strings.Contains(rendered, want) {
					t.Errorf("missing %q in %s", want, rendered)
				}
			}
		})
	}
}

func TestInvalidHeadingTargets(t *testing.T) {
	tests := []struct{ body, message string }{
		{"## One {#same}\n\n## Two {#same}", `duplicate explicit heading ID "same"`},
		{"## One {#main}", `explicit heading ID "main" is reserved`},
		{"## One {#page-title}", `explicit heading ID "page-title" is reserved`},
		{"## One {#cress-caption-1}", `explicit heading ID "cress-caption-1" is reserved`},
		{"## One {#cress-other}", `explicit heading ID "cress-other" is reserved`},
		{`## One {#}`, "must be nonempty text"},
		{`## One {id=true}`, "must be nonempty text"},
		{`## One {id=123}`, "must be nonempty text"},
		{`## One {id="two words"}`, "without ASCII whitespace"},
		{`## One {id="two\twords"}`, "without ASCII whitespace"},
		{"## One {id=\"null\x00byte\"}", "without ASCII whitespace or null characters"},
		{`## One {.custom}`, `unsupported heading attribute "class"`},
	}
	for _, test := range tests {
		t.Run(test.body, func(t *testing.T) {
			_, err := prepareBody(t, test.body)
			if err == nil || !strings.Contains(err.Error(), "content/page.md: heading targets:") ||
				!strings.Contains(err.Error(), test.message) {
				t.Fatalf("error = %v, want source-aware error containing %q", err, test.message)
			}
		})
	}
}

func TestMarkdownTableWrapper(t *testing.T) {
	for _, test := range []struct{ name, prefix string }{
		{"standalone", ""}, {"blockquote", "> "},
	} {
		t.Run(test.name, func(t *testing.T) {
			body := test.prefix + "| Name | Link |\n" + test.prefix + "|:-----|-----:|\n" +
				test.prefix + "| Team & friends | [Contact](/contact/) |\n"
			rendered, err := prepareBody(t, body)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{
				"<div class=\"table-wrapper\" tabindex=\"0\">\n<table>",
				`<th style="text-align:left">Name</th>`,
				`<td style="text-align:left">Team &amp; friends</td>`,
				`<td style="text-align:right"><a href="/contact/">Contact</a></td>`,
				"</table>\n</div>",
			} {
				if !strings.Contains(rendered, want) {
					t.Errorf("missing %q in %s", want, rendered)
				}
			}
		})
	}
	rendered, err := prepareBody(t, "<table><tr><td>Authored HTML</td></tr></table>")
	if err != nil || strings.Contains(rendered, "table-wrapper") {
		t.Fatalf("authored table changed: %s, %v", rendered, err)
	}
}

func TestImageCaptions(t *testing.T) {
	tests := []struct {
		name, body string
		want       []string
		absent     []string
	}{
		{
			name: "standalone",
			body: `![Our **team**](/team.jpg "Hover"){caption="Photo by [Alex](/credits/) <b>not HTML</b>"}`,
			want: []string{`<figure><img src="/team.jpg" alt="Our team" title="Hover">`,
				`<figcaption>Photo by <a href="/credits/">Alex</a> &lt;b&gt;not HTML&lt;/b&gt;</figcaption></figure>`},
			absent: []string{"{caption", "<p><figure>", "<b>not HTML"},
		},
		{
			name: "inline",
			body: `Before ![Team](/team.jpg){caption="Credit"} after.`,
			want: []string{`<p>Before <span class="captioned-image" role="group" aria-describedby="cress-caption-1">`,
				`<img src="/team.jpg" alt="Team"><span id="cress-caption-1" class="image-caption">Credit</span></span> after.</p>`},
			absent: []string{"<figure", "figcaption"},
		},
		{
			name: "multiple",
			body: `![One](/one.jpg){caption="One"} and ![Two](/two.jpg){caption="Two"}`,
			want: []string{`aria-describedby="cress-caption-1"`, `aria-describedby="cress-caption-2"`},
		},
		{
			name:   "empty",
			body:   `![Decorative](/team.jpg){caption=""}`,
			want:   []string{`<p><img src="/team.jpg" alt="Decorative"></p>`},
			absent: []string{"figure", "caption="},
		},
		{
			name: "escaped attribute",
			body: `![Team](/team.jpg){caption="A \"quote\" and \\ path"}`,
			want: []string{`<figcaption>A &#34;quote&#34; and \ path</figcaption>`},
		},
		{
			name: "list",
			body: `- ![Team](/team.jpg){caption="Credit"}`,
			want: []string{"<li>\n<figure><img src=\"/team.jpg\" alt=\"Team\"><figcaption>Credit</figcaption></figure>"},
		},
		{
			name: "raw HTML only caption",
			body: `![Team](/team.jpg){caption="<script>alert(1)</script>"}`,
			want: []string{`<figcaption>&lt;script&gt;alert(1)&lt;/script&gt;</figcaption>`},
		},
		{
			name: "reference image",
			body: "![Team][team]{caption=\"Credit\"}\n\n[team]: /team.jpg \"Hover\"",
			want: []string{`<figure><img src="/team.jpg" alt="Team" title="Hover"><figcaption>Credit</figcaption></figure>`},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rendered, err := prepareBody(t, test.body)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range test.want {
				if !strings.Contains(rendered, want) {
					t.Errorf("missing %q in %s", want, rendered)
				}
			}
			for _, absent := range test.absent {
				if strings.Contains(rendered, absent) {
					t.Errorf("unexpected %q in %s", absent, rendered)
				}
			}
		})
	}
}

func TestInvalidCaptions(t *testing.T) {
	for _, body := range []string{
		`![Team](/team.jpg){caption='single quotes'}`,
		`![Team](/team.jpg){caption="unfinished}`,
		`![Team](/team.jpg){caption="credit" extra}`,
		`![Team](/team.jpg){caption="**bold**"}`,
		`![Team](/team.jpg){caption="*emphasis*"}`,
		"![Team](/team.jpg){caption=\"`code`\"}",
		`![Team](/team.jpg){caption="![media](/photo.jpg)"}`,
		`![Team](/team.jpg){caption="[Credit](javascript:alert)"}`,
		`![Team](/team.jpg){caption="[Credit](//example.org/)"}`,
		`![Team](/team.jpg){caption="[**Bold link**](/credits/)"}`,
		`[![Team](/team.jpg){caption="Credit"}](/about/)`,
		"```go caption='bad'\ncode\n```",
		"```go caption=\"unfinished\ncode\n```",
		"```go caption=\"Credit\" extra\ncode\n```",
		"```go other caption=\"Credit\"\ncode\n```",
		"```go caption=\"**bold**\"\ncode\n```",
	} {
		t.Run(body, func(t *testing.T) {
			_, err := prepareBody(t, body)
			if err == nil || !strings.Contains(err.Error(), "content/page.md:") {
				t.Fatalf("error = %v, want source-aware caption error", err)
			}

		})
	}
}

func TestCodeWhitespace(t *testing.T) {
	tags := regexp.MustCompile(`<[^>]*>`)
	for _, code := range []string{"\tvar text = \"<&>\"\n", "var text = \"<&>\"\r\n", "var text = \"<&>\""} {
		for _, highlight := range []bool{true, false} {
			config := defaultConfig(t)
			config.Highlight = highlight
			prepared, _, err := content.Prepare("page.md", []byte("+++\ntitle = 'Code'\n+++\n```go\n"+code), config)
			if err != nil {
				t.Fatal(err)
			}
			rendered := string(prepared.Data.Body)
			start := strings.Index(rendered, "<code")
			start += strings.Index(rendered[start:], ">") + 1
			end := strings.Index(rendered, "</code>")
			literal := html.UnescapeString(tags.ReplaceAllString(rendered[start:end], ""))
			want := code
			if !strings.HasSuffix(want, "\n") {
				want += "\n"
			}
			if literal != want {
				t.Errorf("highlight=%v changed code: %q, want %q", highlight, literal, want)
			}
		}
	}
}

func TestCodeHighlightingAndCaptions(t *testing.T) {
	const code = "var message = \"<script>&literal</script>\"\n"
	tags := regexp.MustCompile(`<[^>]*>`)
	for _, enabled := range []bool{true, false} {
		for _, language := range []string{"go", "not-a-real-language", "example.go", ""} {
			for _, attribute := range []string{"", ` caption="From [guide](/guide/) <b>escaped</b>"`} {
				t.Run(language+attribute+map[bool]string{true: "/on", false: "/off"}[enabled], func(t *testing.T) {
					config := defaultConfig(t)
					config.Highlight = enabled
					prepared, _, err := content.Prepare("page.md",
						[]byte("+++\ntitle = 'Code'\n+++\n```"+language+attribute+"\n"+code+"```\n"), config)
					if err != nil {
						t.Fatal(err)
					}
					rendered := string(prepared.Data.Body)
					start := strings.Index(rendered, "<code")
					start += strings.Index(rendered[start:], ">") + 1
					end := strings.Index(rendered, "</code>")
					literal := html.UnescapeString(tags.ReplaceAllString(rendered[start:end], ""))
					if literal != code {
						t.Fatalf("code changed: %q, want %q; HTML: %s", literal, code, rendered)
					}
					if strings.Contains(rendered, "<span") != (enabled && language == "go") {
						t.Fatalf("unexpected highlighting: %s", rendered)
					}
					if attribute != "" && (!strings.HasPrefix(rendered, "<figure><pre>") ||
						!strings.Contains(rendered, `<figcaption>From <a href="/guide/">guide</a> &lt;b&gt;escaped&lt;/b&gt;</figcaption></figure>`)) {
						t.Fatalf("unexpected code caption: %s", rendered)
					}
					if strings.Contains(rendered, "<script>") {
						t.Fatalf("code or caption was trusted as HTML: %s", rendered)
					}
				})
			}
		}
	}
}

func TestMarkdownURLBoundaries(t *testing.T) {
	config := defaultConfig(t)
	config.Site.URL = "/company/"
	body := `[Root](/about/?q=1#team) [same](#section) [query](?q=2) [external](https://example.org/about.md) [mail](mailto:hello@example.org)

## Section

<a href="about.md">raw HTML</a>

![Root](/images/team.jpg)

![Entity](/images/team&amp;amp;.jpg)

[Entity](/about&amp;amp;/?q=&amp;amp;)
`
	prepared, _, err := content.Prepare("page.md", []byte("+++\ntitle = 'Page'\n+++\n"+body), config)
	if err != nil {
		t.Fatal(err)
	}
	rendered := string(prepared.Data.Body)
	for _, want := range []string{
		`href="/company/about/?q=1#team"`, `href="#section"`, `href="?q=2"`,
		`href="https://example.org/about.md"`, `href="mailto:hello@example.org"`,
		`<a href="about.md">raw HTML</a>`, `src="/company/images/team.jpg"`,
		`src="/company/images/team&amp;amp;.jpg"`, `href="/company/about&amp;amp;/?q=&amp;amp;"`,
		`<h2 id="section">Section</h2>`,
	} {
		if !strings.Contains(rendered, want) {
			t.Errorf("missing %q in %s", want, rendered)
		}
	}
}

func TestTitleImageCaptions(t *testing.T) {
	config, _, err := content.ParseConfig([]byte(`
[site]
path = "/company/"
[site.image]
src = "/team.jpg"
alt = "Team"
caption = "By [Alex](/credits/) <b>escaped</b>"
`))
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		prepared, _, err := content.Prepare("page.md", []byte("+++\ntitle = 'Page'\n+++\nBody"), config)
		if err != nil {
			t.Fatal(err)
		}
		if prepared.Data.Image.Caption[1].URL != "/company/credits/" {
			t.Fatalf("inherited image was prefixed more than once: %+v", prepared.Data.Image)
		}
	}
}
