package render_test

import (
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/render"
)

// Highlighting adds tokens inside the code element and changes nothing around
// it, so a theme written against the markup cress has always emitted keeps
// working with the option on.
func TestMarkdownWithHighlighting(t *testing.T) {
	out := mustRender(t, render.New(render.WithHighlighting()), "```go\nx := 1 // note\n```\n")

	assertHTML(t, out,
		[]string{
			`<pre class="chroma">`,
			`<code class="language-go">`,
			`<span class="c1">// note</span>`,
			"</code></pre>",
		},
		// Classes only. A color in the output would be the core styling the one
		// element it promised to leave to the theme.
		[]string{"style="},
	)
}

// The option is what turns tokenizing on. Without it the renderer emits what it
// always emitted, which is every site that has not asked for anything.
func TestMarkdownWithoutHighlighting(t *testing.T) {
	out := mustRender(t, render.New(), "```go\nx := 1\n```\n")

	assertHTML(t, out, []string{`<pre><code class="language-go">x := 1`}, []string{"<span", "chroma"})
}

// A block chroma cannot tokenize has no tokens to show, so it has to come out
// as it does with the option off. Anything else would be a class promising
// colors that never arrive, or markup that moved for no reason.
func TestMarkdownHighlightingLeavesUntokenizedBlocksAlone(t *testing.T) {
	tests := []struct {
		name string
		src  string
	}{
		{"a fence naming no language", "```\nplain\n```\n"},
		{"a fence naming a language chroma does not know", "```notalang\nplain\n```\n"},
		{"an indented code block", "    indented\n"},
		{"inline code", "some `code` here\n"},
		// The info string is the author's text, and it reaches an attribute either
		// way, so both paths have to escape it the same.
		{"a language name that would break out of the attribute", "```go\"onclick=alert(1)\nx\n```\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plain := mustRender(t, render.New(), tc.src)
			highlighted := mustRender(t, render.New(render.WithHighlighting()), tc.src)
			if highlighted != plain {
				t.Errorf("output moved with nothing to tokenize\n got: %s\nwant: %s", highlighted, plain)
			}
			if strings.Contains(highlighted, `"onclick`) {
				t.Errorf("language name reached the attribute unescaped\ngot: %s", highlighted)
			}
		})
	}
}

// The extensions have to compose: highlighting replaces how a code block is
// rendered, and everything else about a page is unaffected by that.
func TestMarkdownHighlightingLeavesTheRestOfThePageAlone(t *testing.T) {
	src := "# Title\n\n![Boats](/img/harbor.webp \"Photo: Jane Doe\")\n\n" +
		"| a | b |\n|---|---|\n| 1 | 2 |\n\n[About](/about.html)\n"

	out := mustRender(t, render.New(render.WithBasePath("/cress"), render.WithHighlighting()), src)

	assertHTML(t, out, []string{
		"<h1",
		"<figcaption>Photo: Jane Doe</figcaption>",
		`src="/cress/img/harbor.webp"`,
		"<table>",
		`href="/cress/about.html"`,
	}, nil)
}

// mustRender renders src, failing the test if it cannot.
func mustRender(t *testing.T, r *render.Renderer, src string) string {
	t.Helper()
	out, err := r.Markdown([]byte(src))
	if err != nil {
		t.Fatalf("Markdown(%q): %v", src, err)
	}
	return string(out)
}
