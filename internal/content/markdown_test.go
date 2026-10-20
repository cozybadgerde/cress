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
