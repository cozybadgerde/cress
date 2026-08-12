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
	"github.com/yuin/goldmark/renderer"
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
	basePath  string
	highlight bool
}

// WithBasePath roots the site's own links under basePath, for a site served
// from a subdirectory rather than a domain root. A link or image an author
// wrote as "/about.html" is emitted as "<basePath>/about.html"; see basePrefixer
// for what is deliberately left alone. An empty basePath changes nothing.
func WithBasePath(basePath string) Option {
	return func(s *settings) { s.basePath = basePath }
}

// WithHighlighting tokenizes fenced code blocks, wrapping each token in a span
// carrying a class and no color. Without it a code block is emitted as it
// always was: the language class, and nothing inside it.
//
// It is an option rather than the default because turning it on rewrites the
// HTML of every code block on the site, which is a rendering decision that
// belongs to the author. A theme that ships no token stylesheet renders the
// spans exactly as it rendered the text they wrap.
func WithHighlighting() Option {
	return func(s *settings) { s.highlight = true }
}

// New builds a Renderer. Fenced code blocks are emitted as plain
// <pre><code class="language-...">, leaving syntax styling entirely to the
// theme's CSS; WithHighlighting adds the tokens for that CSS to reach, and
// still supplies no colors. Raw HTML in the source is passed through so authors
// can drop markup into their Markdown. An image that is the whole of its
// paragraph is wrapped in a <figure>, with its title as the caption.
func New(opts ...Option) *Renderer {
	var s settings
	for _, opt := range opts {
		opt(&s)
	}

	// The prefixer runs first so the wrapper never has to reason about a rewritten
	// tree; the image keeps its destination either way, since it survives as the
	// figure's child.
	transformers := []util.PrioritizedValue{}
	if s.basePath != "" {
		transformers = append(transformers, util.Prioritized(&basePrefixer{basePath: s.basePath}, 100))
	}
	transformers = append(transformers, util.Prioritized(figureWrapper{}, 200))

	extensions := []goldmark.Extender{extension.GFM}
	if s.highlight {
		extensions = append(extensions, highlightExtension())
	}

	return &Renderer{md: goldmark.New(
		goldmark.WithExtensions(extensions...),
		goldmark.WithParserOptions(parser.WithASTTransformers(transformers...)),
		goldmark.WithRendererOptions(
			ghtml.WithUnsafe(),
			renderer.WithNodeRenderers(util.Prioritized(figureRenderer{}, 100)),
		),
	)}
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
