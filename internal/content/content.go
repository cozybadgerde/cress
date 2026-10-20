package content

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"maps"
	"net/url"
	"path"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/pelletier/go-toml/v2"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/cozybadgerde/cress/internal/page"
)

// Page associates prepared data with its source and output destination.
type Page struct {
	Source      string
	Destination string
	Draft       bool
	Data        page.Page
}

// Prepare validates a content-relative Markdown source and prepares its page.
func Prepare(name string, source []byte) (Page, []string, error) {
	prepared := Page{Source: path.Join("content", name)}
	destination, route, err := resolveRoute(name)
	if err != nil {
		return prepared, nil, fmt.Errorf("%s: %w", prepared.Source, err)
	}
	prepared.Destination = destination
	prepared.Data.URL = route

	if !utf8.Valid(source) {
		return prepared, nil, fmt.Errorf("%s: content must be UTF-8", prepared.Source)
	}
	frontmatter, body, err := splitFrontmatter(source)
	if err != nil {
		return prepared, nil, fmt.Errorf("%s: %w", prepared.Source, err)
	}
	var fields map[string]any
	if err := toml.Unmarshal(frontmatter, &fields); err != nil {
		return prepared, nil, fmt.Errorf("%s: parse frontmatter: %w", prepared.Source, err)
	}

	var warnings []string
	for _, key := range slices.Sorted(maps.Keys(fields)) {
		if key != "title" && key != "draft" {
			warnings = append(warnings, fmt.Sprintf("%s: unknown frontmatter field %q", prepared.Source, key))
		}
	}

	if value, exists := fields["draft"]; exists {
		draft, ok := value.(bool)
		if !ok {
			return prepared, warnings, fmt.Errorf("%s: draft must be a boolean", prepared.Source)
		}
		prepared.Draft = draft
	}
	if value, exists := fields["title"]; exists {
		title, ok := value.(string)
		if !ok || strings.TrimSpace(title) == "" {
			return prepared, warnings, fmt.Errorf("%s: title must be a nonempty string", prepared.Source)
		}
		prepared.Data.Title = title
	}
	if prepared.Draft {
		return prepared, warnings, nil
	}
	if prepared.Data.Title == "" {
		return prepared, warnings, fmt.Errorf("%s: published pages require a title", prepared.Source)
	}

	markdown := goldmark.New(
		goldmark.WithExtensions(extension.Table),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	var rendered bytes.Buffer
	if err := markdown.Convert(body, &rendered); err != nil {
		return prepared, warnings, fmt.Errorf("%s: render Markdown: %w", prepared.Source, err)
	}
	prepared.Data.Body = template.HTML(rendered.String())
	return prepared, warnings, nil
}

func splitFrontmatter(source []byte) ([]byte, []byte, error) {
	first, rest, _ := bytes.Cut(source, []byte("\n"))
	if string(bytes.TrimSuffix(first, []byte("\r"))) != "+++" {
		return nil, source, nil
	}
	offset := len(source) - len(rest)
	start := offset
	for {
		line, remaining, newline := bytes.Cut(source[offset:], []byte("\n"))
		if string(bytes.TrimSuffix(line, []byte("\r"))) == "+++" {
			return source[start:offset], remaining, nil
		}
		if !newline {
			return nil, nil, fmt.Errorf("frontmatter is missing its closing +++ delimiter")
		}
		offset += len(line) + 1
	}
}

func resolveRoute(name string) (string, string, error) {
	if !fs.ValidPath(name) || strings.ContainsAny(name, "\\\x00") || !utf8.ValidString(name) ||
		path.Ext(name) != ".md" {
		return "", "", fmt.Errorf("invalid Markdown source path %q", name)
	}
	section := path.Dir(name)
	stem := strings.TrimSuffix(path.Base(name), ".md")
	if stem == "" || stem == "." || stem == ".." {
		return "", "", fmt.Errorf("invalid page name %q", name)
	}
	if stem != "index" {
		section = path.Join(section, stem)
	}
	if section == "." {
		section = ""
	}
	destination := path.Join(section, "index.html")
	route := "/"
	if section != "" {
		route = (&url.URL{Path: "/" + section + "/"}).EscapedPath()
	}
	return destination, route, nil
}
