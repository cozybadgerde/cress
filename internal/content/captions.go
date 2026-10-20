package content

import (
	"fmt"
	"html"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"

	"github.com/cozybadgerde/cress/internal/page"
)

func parseCaption(value string, resolve func(string) (string, error)) ([]page.CaptionPart, error) {
	if value == "" {
		return nil, nil
	}
	if strings.ContainsAny(value, "\r\n") {
		return nil, fmt.Errorf("caption must be a single line of text and links")
	}
	source := []byte(value)
	captionParser := parser.NewParser(
		parser.WithBlockParsers(util.Prioritized(parser.NewParagraphParser(), 100)),
		parser.WithInlineParsers(parser.DefaultInlineParsers()...),
	)
	document := captionParser.Parse(text.NewReader(source))
	if document.ChildCount() == 0 {
		return nil, nil
	}
	paragraph, ok := document.FirstChild().(*ast.Paragraph)
	if !ok || document.ChildCount() != 1 {
		return nil, fmt.Errorf("caption supports only text and links")
	}
	var parts []page.CaptionPart
	for node := paragraph.FirstChild(); node != nil; node = node.NextSibling() {
		if link, ok := node.(*ast.Link); ok {
			label, err := captionText(link, source)
			if err != nil {
				return nil, err
			}
			destination := decodeDestination(link.Destination)
			parsed, err := parseURL(destination)
			if err != nil {
				return nil, err
			}
			if parsed.Host != "" && parsed.Scheme == "" {
				return nil, fmt.Errorf("caption links must not be protocol-relative")
			}
			url, err := resolve(destination)
			if err != nil {
				return nil, err
			}
			parts = append(parts, page.CaptionPart{Text: label, URL: url})
			continue
		}
		value, err := captionNodeText(node, source)
		if err != nil {
			return nil, err
		}
		parts = append(parts, page.CaptionPart{Text: value})
	}
	return parts, nil
}

func captionText(parent ast.Node, source []byte) (string, error) {
	var output strings.Builder
	for node := parent.FirstChild(); node != nil; node = node.NextSibling() {
		value, err := captionNodeText(node, source)
		if err != nil {
			return "", err
		}
		output.WriteString(value)
	}
	return output.String(), nil
}

func captionNodeText(node ast.Node, source []byte) (string, error) {
	switch node := node.(type) {
	case *ast.Text:
		return decodeDestination(node.Value(source)), nil
	case *ast.String:
		return string(node.Value), nil
	case *ast.RawHTML:
		var output strings.Builder
		for index := 0; index < node.Segments.Len(); index++ {
			segment := node.Segments.At(index)
			output.Write(segment.Value(source))
		}
		return output.String(), nil
	default:
		return "", fmt.Errorf("caption supports only text and links, not %s", node.Kind())
	}
}

func captionHTML(parts []page.CaptionPart) string {
	var output strings.Builder
	for _, part := range parts {
		if part.URL == "" {
			output.WriteString(html.EscapeString(part.Text))
		} else {
			output.WriteString(`<a href="` + html.EscapeString(part.URL) + `">` + html.EscapeString(part.Text) + `</a>`)
		}
	}
	return output.String()
}

func readCaptionAttribute(value string) (string, int, error) {
	if !strings.HasPrefix(value, `caption="`) {
		return "", 0, fmt.Errorf(`caption attribute must use caption="..."`)
	}
	var caption strings.Builder
	for index := len(`caption="`); index < len(value); index++ {
		switch value[index] {
		case '"':
			return caption.String(), index + 1, nil
		case '\\':
			if index+1 < len(value) && (value[index+1] == '\\' || value[index+1] == '"') {
				index++
			}
			caption.WriteByte(value[index])
		default:
			caption.WriteByte(value[index])
		}
	}
	return "", 0, fmt.Errorf("caption attribute is missing its closing quote")
}

type imageCaptionParser struct {
	captions map[*ast.Image]string
	markers  []ast.Node
	err      error
}

func (captions *imageCaptionParser) Trigger() []byte { return []byte{'{'} }

func (captions *imageCaptionParser) Parse(parent ast.Node, reader text.Reader, _ parser.Context) ast.Node {
	image, ok := parent.LastChild().(*ast.Image)
	if !ok {
		return nil
	}
	line, _ := reader.PeekLine()
	if !strings.HasPrefix(string(line), "{caption") {
		return nil
	}
	caption, consumed, err := readCaptionAttribute(string(line[1:]))
	if err == nil && (consumed+1 >= len(line) || line[consumed+1] != '}') {
		err = fmt.Errorf("image caption attribute must end with }")
	}
	if err != nil {
		captions.err = err
		return nil
	}
	captions.captions[image] = caption
	reader.Advance(consumed + 2)
	marker := ast.NewString(nil)
	captions.markers = append(captions.markers, marker)
	return marker
}
