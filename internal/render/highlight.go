package render

import (
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	ghtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// chromaClass marks a <pre> whose contents were tokenized, so a stylesheet has
// something to scope its token rules under. It is chroma's own convention
// rather than a name cress invented, which is what lets any published chroma
// stylesheet apply to a cress site unchanged.
const chromaClass = "chroma"

// languageClassPrefix is the class a fenced code block carries, naming the
// language its info string declared.
const languageClassPrefix = "language-"

// highlightExtension tokenizes fenced code blocks, emitting classes and no
// colors.
//
// Classes only, never inline styles: cress emits semantic HTML and leaves every
// color to the theme, and a baked-in chroma color scheme would be the core
// styling the one element it had promised not to. With no stylesheet the spans
// are inert and code renders exactly as it did before, which is what makes this
// safe to hand to a theme that has not heard of it.
func highlightExtension() goldmark.Extender {
	return highlighting.NewHighlighting(
		highlighting.WithFormatOptions(
			chromahtml.WithClasses(true),
			chromahtml.WithPreWrapper(nopPreWrapper{}),
		),
		highlighting.WithWrapperRenderer(writeCodeBlock),
	)
}

// nopPreWrapper stops chroma from emitting a <pre><code> of its own. The
// wrapper is writeCodeBlock's job, and leaving chroma's in place would nest a
// second code block inside the first.
type nopPreWrapper struct{}

// Start implements chromahtml.PreWrapper.
func (nopPreWrapper) Start(_ bool, _ string) string { return "" }

// End implements chromahtml.PreWrapper.
func (nopPreWrapper) End(_ bool) string { return "" }

// writeCodeBlock wraps a highlighted block in the same markup cress emits
// without highlighting, so turning the option on adds tokens inside the code
// element and changes nothing around it.
//
// It exists because the extension's own wrapper drops the language class on
// exactly the blocks it recognizes, which would leave a theme styling
// `.language-go` for every language chroma cannot read and none of the ones it
// can. The chroma class is added only where there are tokens to scope: a fence
// with no language, or one naming a language chroma does not know, comes out
// byte for byte as it does today.
func writeCodeBlock(w util.BufWriter, ctx highlighting.CodeBlockContext, entering bool) {
	if !entering {
		_, _ = w.WriteString("</code></pre>\n")
		return
	}

	_, _ = w.WriteString("<pre")
	if ctx.Highlighted() {
		_, _ = w.WriteString(` class="` + chromaClass + `"`)
	}
	_, _ = w.WriteString("><code")
	if language, ok := ctx.Language(); ok {
		_, _ = w.WriteString(` class="` + languageClassPrefix)
		// The info string is the author's text, so it is escaped on the way into an
		// attribute. This is the writer goldmark uses for the same class.
		ghtml.DefaultWriter.Write(w, language)
		_, _ = w.WriteString(`"`)
	}
	_, _ = w.WriteString(">")
}
