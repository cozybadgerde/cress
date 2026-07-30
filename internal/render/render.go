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
	"github.com/yuin/goldmark/parser"
	ghtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// Renderer converts Markdown to HTML with a fixed, shared configuration.
type Renderer struct {
	md goldmark.Markdown
}

// Option configures a Renderer. Every option is a tunable with a working
// default, so New() alone builds the renderer cress used before any of them
// existed.
type Option func(*settings)

// settings collects the optional configuration New applies.
type settings struct {
	basePath string
}

// WithBasePath roots the site's own links under basePath, for a site served
// from a subdirectory rather than a domain root. A link or image an author
// wrote as "/about.html" is emitted as "<basePath>/about.html"; see basePrefixer
// for what is deliberately left alone. An empty basePath changes nothing.
func WithBasePath(basePath string) Option {
	return func(s *settings) { s.basePath = basePath }
}

// New builds a Renderer. Fenced code blocks are emitted as plain
// <pre><code class="language-...">, leaving syntax styling entirely to the
// theme's CSS. Raw HTML in the source is passed through so authors can drop
// markup into their Markdown.
func New(opts ...Option) *Renderer {
	var s settings
	for _, opt := range opts {
		opt(&s)
	}

	goldmarkOpts := []goldmark.Option{
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(ghtml.WithUnsafe()),
	}
	if s.basePath != "" {
		goldmarkOpts = append(goldmarkOpts, goldmark.WithParserOptions(
			parser.WithASTTransformers(util.Prioritized(&basePrefixer{basePath: s.basePath}, 100)),
		))
	}
	return &Renderer{md: goldmark.New(goldmarkOpts...)}
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
