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

func TestMarkdownEmpty(t *testing.T) {
	out, err := render.New().Markdown(nil)
	if err != nil {
		t.Fatalf("Markdown(nil): %v", err)
	}
	if strings.TrimSpace(string(out)) != "" {
		t.Errorf("Markdown(nil) = %q, want empty", out)
	}
}
