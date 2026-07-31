package render

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	ghtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// kindFigure identifies the figure node to the renderer, which dispatches on
// node kind alone.
var kindFigure = ast.NewNodeKind("Figure")

// figure is the block a standalone image becomes: the image itself as its only
// child, plus the visible caption text, which may be empty.
type figure struct {
	ast.BaseBlock
	caption []byte
}

// Kind implements ast.Node.
func (n *figure) Kind() ast.NodeKind { return kindFigure }

// Dump implements ast.Node, printing the tree for debugging.
func (n *figure) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Caption": string(n.caption)}, nil)
}

// figureWrapper turns an image that is the whole of its paragraph into a
// <figure>, promoting the CommonMark image title to a <figcaption>.
//
// Captions carry obligations an alt text cannot: an attribution the licence
// demands, or a disclosure the law does. Both have to be visible to everyone,
// which alt text is not, and both are common enough that hand-written <figure>
// markup is the wrong price to ask. Reusing the image title keeps the source
// plain CommonMark, so the same document still renders elsewhere, with the
// caption degrading to a tooltip.
type figureWrapper struct{}

// Transform replaces every standalone-image paragraph in doc with a figure.
func (figureWrapper) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	// Collect first, rewrite after. Detaching a node mid-walk clears its sibling
	// links, and the walk reads those to find what to visit next, so replacing in
	// place would silently drop the rest of the document.
	var standalone []*ast.Paragraph

	// The walk cannot fail: the visitor below returns no error, and the only other
	// source would be a malformed tree the parser just produced.
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if para, ok := n.(*ast.Paragraph); ok && soleImage(para) != nil {
			standalone = append(standalone, para)
		}
		return ast.WalkContinue, nil
	})

	for _, para := range standalone {
		img := soleImage(para)
		fig := &figure{caption: img.Title}
		// The caption is the visible copy of this text, so leaving the title in
		// place would render it a second time as a tooltip.
		img.Title = nil

		para.RemoveChild(para, img)
		fig.AppendChild(fig, img)

		parent := para.Parent()
		parent.ReplaceChild(parent, para, fig)
	}
}

// soleImage returns the image para consists of, or nil if para is anything
// else. An image sharing its paragraph with text is part of a sentence, not a
// figure, and two images in one paragraph are the author's own arrangement.
func soleImage(para *ast.Paragraph) *ast.Image {
	if para.ChildCount() != 1 {
		return nil
	}
	img, _ := para.FirstChild().(*ast.Image)
	return img
}

// figureRenderer writes the HTML for a figure node.
type figureRenderer struct{}

// RegisterFuncs implements renderer.NodeRenderer.
func (r figureRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindFigure, r.render)
}

// render emits the figure around its already-rendered image. A figure without a
// caption still wraps, so a theme has one shape to style either way.
func (figureRenderer) render(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString("<figure>\n")
		return ast.WalkContinue, nil
	}

	fig, ok := node.(*figure)
	if !ok {
		return ast.WalkStop, nil
	}
	if len(fig.caption) > 0 {
		_, _ = w.WriteString("\n<figcaption>")
		// The same writer goldmark uses for a title attribute: it escapes the text
		// and resolves the character references an author may have written.
		ghtml.DefaultWriter.Write(w, fig.caption)
		_, _ = w.WriteString("</figcaption>")
	}
	_, _ = w.WriteString("\n</figure>\n")
	return ast.WalkContinue, nil
}
