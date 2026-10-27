package content

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"slices"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extensionast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	goldhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type markup struct {
	ast.BaseInline
	value string
}

var markupKind = ast.NewNodeKind("PreparedMarkup")

func (node *markup) Kind() ast.NodeKind { return markupKind }
func (node *markup) Dump(source []byte, level int) {
	ast.DumpHelper(node, source, level, nil, nil)
}

type markupRenderer struct{}

type codeOptions struct {
	highlight  bool
	references references
}

func (markupRenderer) RegisterFuncs(register renderer.NodeRendererFuncRegisterer) {
	register.Register(markupKind, func(writer util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			_, err := writer.WriteString(node.(*markup).value)
			return ast.WalkSkipChildren, err
		}
		return ast.WalkContinue, nil
	})
}

func renderContent(prepared *Page, config Config, files map[string]reference) error {
	references := references{prefix: config.Site.URL, source: prepared.Source, files: files}
	if prepared.pageImage && prepared.Data.Image != nil {
		if err := resolvePageImage(prepared, references); err != nil {
			return err
		}
	}
	captions := &imageCaptionParser{captions: make(map[*ast.Image]string)}
	markdown := goldmark.New(
		goldmark.WithExtensions(extension.Table),
		goldmark.WithParserOptions(parser.WithInlineParsers(util.Prioritized(captions, 50))),
		goldmark.WithRendererOptions(goldhtml.WithUnsafe(),
			renderer.WithNodeRenderers(util.Prioritized(markupRenderer{}, 50))),
	)
	document := markdown.Parser().Parse(text.NewReader(prepared.body))
	if captions.err != nil {
		return fmt.Errorf("%s: %w", prepared.Source, captions.err)
	}
	for _, marker := range captions.markers {
		if parent := marker.Parent(); parent != nil {
			parent.RemoveChild(parent, marker)
		}
	}
	var images []*ast.Image
	var blocks []*ast.FencedCodeBlock
	var tables []*extensionast.Table
	err := ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch node := node.(type) {
		case *ast.Link:
			resolved, err := references.resolve(decodeDestination(node.Destination))
			if err != nil {
				return ast.WalkStop, err
			}
			node.Destination = util.EscapeHTML([]byte(resolved))
		case *ast.Image:
			resolved, err := references.resolveImage(decodeDestination(node.Destination))
			if err != nil {
				return ast.WalkStop, err
			}
			node.Destination = util.EscapeHTML([]byte(resolved))
			images = append(images, node)
		case *ast.FencedCodeBlock:
			blocks = append(blocks, node)
		case *extensionast.Table:
			tables = append(tables, node)
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return fmt.Errorf("%s: resolve Markdown: %w", prepared.Source, err)
	}
	if err := replaceCaptionedImages(prepared, markdown, images, captions.captions, references); err != nil {
		return err
	}
	for _, block := range blocks {
		rendered, err := renderCode(block, prepared.body, codeOptions{highlight: config.Highlight, references: references})
		if err != nil {
			return fmt.Errorf("%s: code block: %w", prepared.Source, err)
		}
		block.Parent().ReplaceChild(block.Parent(), block, &markup{value: rendered})
	}
	for _, table := range tables {
		var rendered bytes.Buffer
		if err := markdown.Renderer().Render(&rendered, prepared.body, table); err != nil {
			return fmt.Errorf("%s: render table: %w", prepared.Source, err)
		}
		table.Parent().ReplaceChild(table.Parent(), table, &markup{
			value: "<div class=\"table-wrapper\" tabindex=\"0\">\n" + rendered.String() + "</div>\n",
		})
	}
	var rendered bytes.Buffer
	if err := markdown.Renderer().Render(&rendered, prepared.body, document); err != nil {
		return fmt.Errorf("%s: render Markdown: %w", prepared.Source, err)
	}
	prepared.Data.Body = template.HTML(rendered.String())
	return nil
}

func resolvePageImage(prepared *Page, references references) error {
	image := prepared.Data.Image
	resolved, err := references.resolveImage(image.URL)
	if err != nil {
		return fmt.Errorf("%s: image.src: %w", prepared.Source, err)
	}
	image.URL = resolved
	for index := range image.Caption {
		if image.Caption[index].URL != "" {
			resolved, err := references.resolve(image.Caption[index].URL)
			if err != nil {
				return fmt.Errorf("%s: image.caption: %w", prepared.Source, err)
			}
			image.Caption[index].URL = resolved
		}
	}
	return nil
}

// replaceCaptionedImages keeps image indexes stable because inline caption
// IDs derive from them.
func replaceCaptionedImages(prepared *Page, markdown goldmark.Markdown, images []*ast.Image, captions map[*ast.Image]string, references references) error {
	for index, image := range images {
		caption, exists := captions[image]
		if !exists || caption == "" {
			continue
		}
		parts, err := parseCaption(caption, references.resolve)
		if err != nil {
			return fmt.Errorf("%s: image caption: %w", prepared.Source, err)
		}
		if len(parts) == 0 {
			continue
		}
		for ancestor := image.Parent(); ancestor != nil; ancestor = ancestor.Parent() {
			if ancestor.Kind() == ast.KindLink || ancestor.Kind() == ast.KindImage {
				return fmt.Errorf("%s: captioned images must not be nested inside links or image alternatives", prepared.Source)
			}
		}
		var rendered bytes.Buffer
		if err := markdown.Renderer().Render(&rendered, prepared.body, image); err != nil {
			return fmt.Errorf("%s: render image: %w", prepared.Source, err)
		}
		parent := image.Parent()
		if (parent.Kind() == ast.KindParagraph || parent.Kind() == ast.KindTextBlock) && parent.ChildCount() == 1 {
			parent.Parent().ReplaceChild(parent.Parent(), parent, &markup{
				value: "<figure>" + rendered.String() + "<figcaption>" + captionHTML(parts) + "</figcaption></figure>\n",
			})
		} else {
			id := "cress-caption-" + strconv.Itoa(index+1)
			parent.ReplaceChild(parent, image, &markup{
				value: `<span class="captioned-image" role="group" aria-describedby="` + id + `">` +
					rendered.String() + `<span id="` + id + `" class="image-caption">` + captionHTML(parts) + "</span></span>",
			})
		}
	}
	return nil
}

func renderCode(block *ast.FencedCodeBlock, source []byte, options codeOptions) (string, error) {
	var info string
	if block.Info != nil {
		info = strings.TrimSpace(string(block.Info.Value(source)))
	}
	language := string(block.Language(source))
	attribute := strings.TrimSpace(strings.TrimPrefix(info, language))
	if strings.HasPrefix(info, "caption=") {
		language = ""
		attribute = info
	}
	var caption string
	if strings.Contains(attribute, "caption") {
		value, consumed, err := readCaptionAttribute(attribute)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(attribute[consumed:]) != "" {
			return "", fmt.Errorf("unexpected text after code caption attribute")
		}
		caption = value
	}
	parts, err := parseCaption(caption, options.references.resolve)
	if err != nil {
		return "", err
	}
	var code bytes.Buffer
	for index := 0; index < block.Lines().Len(); index++ {
		segment := block.Lines().At(index)
		code.Write(segment.Value(source))
	}
	formatted := html.EscapeString(code.String())
	if options.highlight {
		formatted, err = highlightCode(code.String(), language)
		if err != nil {
			return "", err
		}
	}
	class := ""
	if language != "" {
		class = ` class="language-` + html.EscapeString(language) + `"`
	}
	result := "<pre><code" + class + ">" + formatted + "</code></pre>\n"
	if len(parts) != 0 {
		result = "<figure>" + result + "<figcaption>" + captionHTML(parts) + "</figcaption></figure>\n"
	}
	return result, nil
}

func highlightCode(code, language string) (string, error) {
	lexer := findLexer(language)
	if lexer == nil {
		return html.EscapeString(code), nil
	}
	tokens, err := chroma.Coalesce(lexer).Tokenise(&chroma.TokeniseOptions{State: "root"}, code)
	if err != nil {
		return "", fmt.Errorf("highlight %s: %w", language, err)
	}
	var output bytes.Buffer
	formatter := chromahtml.New(chromahtml.WithClasses(true), chromahtml.PreventSurroundingPre(true))
	if err := formatter.Format(&output, styles.Fallback, tokens); err != nil {
		return "", fmt.Errorf("format highlighted code: %w", err)
	}
	return output.String(), nil
}

func findLexer(language string) chroma.Lexer {
	if language == "" {
		return nil
	}
	lexer := lexers.Get(language)
	if lexer == nil {
		return nil
	}
	config := lexer.Config()
	if strings.EqualFold(config.Name, language) || slices.ContainsFunc(config.Aliases, func(alias string) bool {
		return strings.EqualFold(alias, language)
	}) {
		return lexer
	}
	return nil
}
