package render_test

import (
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/render"
)

func TestMarkdown(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string // substring that must appear in the output
	}{
		{"paragraph", "hello world\n", "<p>hello world</p>"},
		{"emphasis", "*hi*\n", "<em>hi</em>"},
		{"strong", "**hi**\n", "<strong>hi</strong>"},
		{"heading", "# Title\n", "<h1"},
		{"link", "[x](https://example.com)\n", `href="https://example.com"`},
		{"gfm strikethrough", "~~gone~~\n", "<del>gone</del>"},
		{"gfm table", "| a | b |\n|---|---|\n| 1 | 2 |\n", "<table>"},
		{"gfm task list", "- [x] done\n", `type="checkbox"`},
		{"fenced code carries language class", "```go\nx := 1\n```\n", `<code class="language-go">`},
		{"raw html passes through", "<div class=\"note\">hi</div>\n", `<div class="note">`},
	}

	r := render.New()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := r.Markdown([]byte(tc.src))
			if err != nil {
				t.Fatalf("Markdown: %v", err)
			}
			if !strings.Contains(string(out), tc.want) {
				t.Errorf("output missing %q\ngot: %s", tc.want, out)
			}
		})
	}
}

// A site published to a subdirectory needs the links an author wrote against the
// site root moved under it. Everything else has to survive untouched, since a
// prefix applied to the wrong destination breaks a link that worked.
func TestMarkdownWithBasePath(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{"rooted link", "[About](/about.html)\n", `href="/cress/about.html"`},
		{"rooted image", "![Cress](/img/cress.webp)\n", `src="/cress/img/cress.webp"`},
		{"the home page", "[Home](/)\n", `href="/cress/"`},
		{"rooted link with a fragment", "[Setup](/guide.html#setup)\n", `href="/cress/guide.html#setup"`},
		{"reference link", "[About][a]\n\n[a]: /about.html\n", `href="/cress/about.html"`},

		{"relative link is untouched", "[About](about.html)\n", `href="about.html"`},
		{"parent-relative link is untouched", "[Up](../index.html)\n", `href="../index.html"`},
		{"absolute URL is untouched", "[Ext](https://example.com/x)\n", `href="https://example.com/x"`},
		{"protocol-relative URL is untouched", "[CDN](//cdn.example.com/x)\n", `href="//cdn.example.com/x"`},
		{"fragment is untouched", "[Top](#top)\n", `href="#top"`},
		{"mailto is untouched", "[Mail](mailto:hi@example.com)\n", `href="mailto:hi@example.com"`},
		// Raw HTML is passed through verbatim, so its links are the author's to get
		// right. Rewriting them would mean parsing the markup cress did not produce.
		{"raw html is untouched", "<a href=\"/about.html\">About</a>\n", `href="/about.html"`},
	}

	r := render.New(render.WithBasePath("/cress"))
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := r.Markdown([]byte(tc.src))
			if err != nil {
				t.Fatalf("Markdown: %v", err)
			}
			if !strings.Contains(string(out), tc.want) {
				t.Errorf("output missing %q\ngot: %s", tc.want, out)
			}
		})
	}
}

// The option is what turns rewriting on; without it the renderer emits what the
// author wrote, which is every site served from a domain root.
func TestMarkdownWithoutBasePath(t *testing.T) {
	for _, r := range []*render.Renderer{render.New(), render.New(render.WithBasePath(""))} {
		out, err := r.Markdown([]byte("[About](/about.html)\n"))
		if err != nil {
			t.Fatalf("Markdown: %v", err)
		}
		if !strings.Contains(string(out), `href="/about.html"`) {
			t.Errorf("link should be untouched, got: %s", out)
		}
	}
}

func TestMarkdownEmpty(t *testing.T) {
	out, err := render.New().Markdown(nil)
	if err != nil {
		t.Fatalf("Markdown(nil): %v", err)
	}
	if strings.TrimSpace(string(out)) != "" {
		t.Errorf("Markdown(nil) = %q, want empty", out)
	}
}
