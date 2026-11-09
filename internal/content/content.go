package content

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/url"
	"path"
	"strings"
	"unicode/utf8"

	"github.com/pelletier/go-toml/v2"

	"github.com/cozybadgerde/cress/internal/page"
)

// Page associates prepared data with its source and output destination.
type Page struct {
	Source      string
	Destination string
	Draft       bool
	Data        page.Page
	body        []byte
	pageImage   bool
	list        listOptions
}

// Prepare validates a content-relative Markdown source and prepares its page.
func Prepare(name string, source []byte, config Config) (Page, []string, error) {
	prepared, warnings, err := prepareMetadata(name, source, config)
	if err != nil || (prepared.Draft && !config.IncludeDrafts) {
		return prepared, warnings, err
	}
	if err := renderContent(&prepared, config, nil); err != nil {
		return prepared, warnings, err
	}
	return prepared, warnings, nil
}

func prepareMetadata(name string, source []byte, config Config) (Page, []string, error) {
	prepared := Page{Source: path.Join("content", name)}
	destination, route, err := resolveRoute(name)
	if err != nil {
		return prepared, nil, fmt.Errorf("%s: %w", prepared.Source, err)
	}
	prepared.Destination = destination
	prepared.Data = page.Page{
		URL:         strings.TrimSuffix(config.Site.URL, "/") + route,
		Type:        page.Ordinary,
		Description: config.Description,
		Site:        config.Site,
	}

	if !utf8.Valid(source) {
		return prepared, nil, fmt.Errorf("%s: content must be UTF-8", prepared.Source)
	}
	frontmatter, body, err := splitFrontmatter(source)
	if err != nil {
		return prepared, nil, fmt.Errorf("%s: %w", prepared.Source, err)
	}
	var values map[string]any
	if err := toml.Unmarshal(frontmatter, &values); err != nil {
		return prepared, nil, fmt.Errorf("%s: parse frontmatter: %w", prepared.Source, err)
	}

	var warnings []string
	input := fields{values: values, source: prepared.Source, kind: "frontmatter", warnings: &warnings}
	input.warnUnknown("title", "draft", "description", "type", "image",
		"published_at", "updated_at", "excerpt", "list")

	if value, exists := values["draft"]; exists {
		draft, ok := value.(bool)
		if !ok {
			return prepared, warnings, fmt.Errorf("%s: draft must be a boolean", prepared.Source)
		}
		prepared.Draft = draft
	}
	prepared.Data.Title, err = input.readNonblank("title", "")
	if err != nil {
		return prepared, warnings, err
	}
	prepared.Data.Description, err = input.readString("description", config.Description)
	if err != nil {
		return prepared, warnings, err
	}
	pageType, err := input.readString("type", string(page.Ordinary))
	if err != nil {
		return prepared, warnings, err
	}
	switch page.Type(pageType) {
	case page.Ordinary, page.Landing, page.List:
		prepared.Data.Type = page.Type(pageType)
	default:
		return prepared, warnings, input.invalidField("type", "must be ordinary, landing, or list")
	}
	if prepared.Data.Type == page.Landing {
		prepared.Data.Image = config.Image
	}
	prepared.Data.PublishedAt, err = input.readDate("published_at")
	if err != nil {
		return prepared, warnings, err
	}
	prepared.Data.UpdatedAt, err = input.readDate("updated_at")
	if err != nil {
		return prepared, warnings, err
	}
	prepared.Data.Excerpt, err = input.readString("excerpt", "")
	if err != nil {
		return prepared, warnings, err
	}
	prepared.list, err = input.readList(prepared.Data.Type, config.ListPerPage)
	if err != nil {
		return prepared, warnings, err
	}
	if _, exists := values["image"]; exists {
		prepared.pageImage = true
		prepared.Data.Image, err = input.readImage("")
		if err != nil {
			return prepared, warnings, err
		}
	}
	if prepared.Draft && !config.IncludeDrafts {
		return prepared, warnings, nil
	}
	if prepared.Data.Title == "" {
		return prepared, warnings, fmt.Errorf("%s: published pages require a title", prepared.Source)
	}
	if config.Host != "" {
		prepared.Data.Canonical = config.Host + prepared.Data.URL
	}
	if prepared.Data.URL == config.Site.URL {
		prepared.Data.StructuredData = config.Profile
	}

	prepared.body = body
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
