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

// An image alone in its paragraph becomes a figure, so a caption has somewhere
// to live. The title carries text a licence or a disclosure obliges the author
// to show, which is why it has to be visible rather than a tooltip.
func TestMarkdownFigure(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []string
		omit []string
	}{
		{
			name: "titled image becomes a captioned figure",
			src:  "![Boats](/img/harbor.webp \"Photo: Jane Doe, CC BY-SA 4.0\")\n",
			want: []string{
				"<figure>",
				`<img src="/img/harbor.webp" alt="Boats">`,
				"<figcaption>Photo: Jane Doe, CC BY-SA 4.0</figcaption>",
				"</figure>",
			},
			// The caption already shows this text; a title would repeat it as a tooltip.
			omit: []string{"title="},
		},
		{
			name: "untitled image still wraps, with no caption",
			src:  "![Boats](/img/harbor.webp)\n",
			want: []string{"<figure>", `alt="Boats"`, "</figure>"},
			omit: []string{"<figcaption>"},
		},
		{
			name: "empty title emits no caption",
			src:  "![Boats](/img/harbor.webp \"\")\n",
			want: []string{"<figure>"},
			omit: []string{"<figcaption>"},
		},
		{
			name: "empty alt is preserved, since decorative is a valid choice",
			src:  "![](/img/harbor.webp \"Photo: Jane Doe\")\n",
			want: []string{`alt=""`, "<figcaption>Photo: Jane Doe</figcaption>"},
		},
		{
			name: "the figure replaces the paragraph rather than nesting in one",
			src:  "![Boats](/img/harbor.webp)\n",
			omit: []string{"<p>"},
		},
		{
			name: "a caption is escaped, not markup",
			src:  "![Boats](/img/harbor.webp \"<b>Jane</b> & Co\")\n",
			want: []string{"<figcaption>&lt;b&gt;Jane&lt;/b&gt; &amp; Co</figcaption>"},
			omit: []string{"<b>Jane</b>"},
		},
		{
			name: "a caption is text, not Markdown",
			src:  "![Boats](/img/harbor.webp \"Photo by *Jane*\")\n",
			want: []string{"<figcaption>Photo by *Jane*</figcaption>"},
			omit: []string{"<em>"},
		},
		{
			name: "an image inside a sentence is left alone",
			src:  "See ![Boats](/img/harbor.webp \"A tooltip\") here.\n",
			want: []string{"<p>", `title="A tooltip"`},
			omit: []string{"<figure>"},
		},
		{
			name: "two images in one paragraph are the author's arrangement",
			src:  "![One](/img/a.webp)\n![Two](/img/b.webp)\n",
			want: []string{"<p>"},
			omit: []string{"<figure>"},
		},
		{
			name: "a linked image is not a figure",
			src:  "[![Boats](/img/harbor.webp \"A tooltip\")](/img/full.webp)\n",
			want: []string{"<p>", `href="/img/full.webp"`, `title="A tooltip"`},
			omit: []string{"<figure>"},
		},
		{
			// Everything after the first replacement has to survive: the wrapper
			// detaches nodes from the tree it is walking.
			name: "surrounding content survives the rewrite",
			src:  "Before\n\n![One](/img/a.webp)\n\nBetween\n\n![Two](/img/b.webp)\n\nAfter\n",
			want: []string{"<p>Before</p>", "<p>Between</p>", "<p>After</p>", `alt="One"`, `alt="Two"`},
		},
		{
			name: "a figure nested in a list item is wrapped too",
			src:  "- item\n\n  ![Boats](/img/harbor.webp \"Photo: Jane Doe\")\n",
			want: []string{"<li>", "<figure>", "<figcaption>Photo: Jane Doe</figcaption>"},
		},
	}

	r := render.New()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			out, err := r.Markdown([]byte(tc.src))
			if err != nil {
				t.Fatalf("Markdown: %v", err)
			}
			assertHTML(t, string(out), tc.want, tc.omit)
		})
	}
}

// assertHTML reports every string in want that out is missing, and every string
// in omit that it should not have carried.
func assertHTML(t *testing.T, out string, want, omit []string) {
	t.Helper()
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("output missing %q\ngot: %s", w, out)
		}
	}
	for _, o := range omit {
		if strings.Contains(out, o) {
			t.Errorf("output should not contain %q\ngot: %s", o, out)
		}
	}
}

// The two transformers have to compose: a captioned image is still an image
// whose destination needs rooting under the base path.
func TestMarkdownFigureWithBasePath(t *testing.T) {
	out, err := render.New(render.WithBasePath("/cress")).
		Markdown([]byte("![Boats](/img/harbor.webp \"Photo: Jane Doe\")\n"))
	if err != nil {
		t.Fatalf("Markdown: %v", err)
	}
	assertHTML(t, string(out),
		[]string{`src="/cress/img/harbor.webp"`, "<figcaption>Photo: Jane Doe</figcaption>"}, nil)
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
