package render

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// basePrefixer rewrites the links and images an author wrote against the site
// root so they resolve under a base path instead.
//
// Without it, a site served from a subdirectory breaks in the one place cress
// cannot see: "[About](/about.html)" in a content file is the author's own
// markup, not something the builder emits, so nothing else in the pipeline gets
// the chance to root it. Rewriting here keeps a single way to write an in-site
// link, whatever the site is later published under.
type basePrefixer struct {
	// basePath is rooted and carries no trailing slash, as config guarantees.
	basePath string
}

// Transform prefixes every site-rooted link and image destination in doc.
func (p *basePrefixer) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	// The walk cannot fail: the visitor below returns no error, and the only other
	// source would be a malformed tree the parser just produced.
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch node := n.(type) {
		case *ast.Link:
			node.Destination = p.prefix(node.Destination)
		case *ast.Image:
			node.Destination = p.prefix(node.Destination)
		}
		return ast.WalkContinue, nil
	})
}

// prefix roots dest under the base path, or returns it untouched when it is not
// a link into this site's root.
func (p *basePrefixer) prefix(dest []byte) []byte {
	if !siteRooted(dest) {
		return dest
	}
	return append([]byte(p.basePath), dest...)
}

// siteRooted reports whether dest addresses this site from its root, which is
// the only kind of destination a base path applies to. Left alone:
//
//   - a relative destination ("about.html"), which already resolves correctly
//     wherever the page is served from;
//   - an absolute URL ("https://example.com/x"), which names its own site;
//   - a protocol-relative URL ("//cdn.example.com/x"), where the leading slashes
//     are the scheme, not a path;
//   - anything else without a leading slash, including "#section" and
//     "mailto:someone@example.com".
func siteRooted(dest []byte) bool {
	return bytes.HasPrefix(dest, []byte("/")) && !bytes.HasPrefix(dest, []byte("//"))
}
