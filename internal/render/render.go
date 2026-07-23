// Package render converts Markdown source into HTML. It owns the goldmark
// configuration (CommonMark plus GitHub-flavored extensions) and nothing else:
// page assembly and templating live in the theme and build packages.
package render

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	ghtml "github.com/yuin/goldmark/renderer/html"
)

// Renderer converts Markdown to HTML with a fixed, shared configuration.
type Renderer struct {
	md goldmark.Markdown
}

// New builds a Renderer. Fenced code blocks are emitted as plain
// <pre><code class="language-...">, leaving syntax styling entirely to the
// theme's CSS. Raw HTML in the source is passed through so authors can drop
// markup into their Markdown.
func New() *Renderer {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(ghtml.WithUnsafe()),
	)
	return &Renderer{md: md}
}

// Markdown converts src to an HTML fragment.
func (r *Renderer) Markdown(src []byte) (template.HTML, error) {
	var buf bytes.Buffer
	if err := r.md.Convert(src, &buf); err != nil {
		return "", fmt.Errorf("render markdown: %w", err)
	}
	// #nosec G203 -- src is the site author's own trusted content.
	return template.HTML(buf.String()), nil
}
