// Package content discovers Markdown pages under a site's content directory and
// parses each file's YAML front matter, producing the page model the builder
// renders. It performs no Markdown-to-HTML conversion; that is the render
// package's job.
package content

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v2"
)

// yamlFrontMatter is the one front-matter format cress accepts: YAML fenced by
// lines containing only "---".
//
// It has to be named explicitly. frontmatter.Parse falls back to a default set
// of seven formats when given none, TOML (+++) and JSON (;;;) among them, so
// taking that default meant cress silently parsed syntaxes its own guides rule
// out, and a page written in one of them worked until it did not.
//
// The decoder is yaml.v2 because that is the one the library uses for its own
// YAML format: passing it leaves a "---" block decoding exactly as it always
// has, and adds no module the build did not already contain.
var yamlFrontMatter = frontmatter.NewFormat("---", "---", yaml.Unmarshal)

const (
	mdExt   = ".md"
	htmlExt = ".html"
	// indexName is the basename (without extension) treated as a directory
	// index: content/guide/index.md serves the /guide/ URL.
	indexName = "index"
)

// Page is a single source document ready to be rendered.
type Page struct {
	// SourcePath is the file path relative to the content root, in slash form
	// (e.g. "guide/setup.md").
	SourcePath string
	// OutputPath is where the rendered file lands relative to the output root,
	// in slash form (e.g. "guide/setup.html").
	OutputPath string
	// URL is the site-root-relative URL for the page (e.g. "/guide/setup.html",
	// or "/guide/" for an index file).
	URL string
	// Title is the page title: front-matter "title", else the first H1, else the
	// file's base name.
	Title string
	// Description is the front-matter "description", empty when the page sets
	// none. The site-wide fallback is the builder's to apply: this package knows
	// the file and nothing else.
	Description string
	// Language is the front-matter "language" tag, empty when the page sets none.
	// It overrides the site's language for this page, which the builder resolves.
	Language string
	// Image is the front-matter "image": this page's lead image, empty when the
	// page names none. It is a field of its own rather than a key left in Meta
	// because cress resolves it against the site-wide value, which a theme
	// reading Meta could not do. The fallback is the builder's to apply: this
	// package knows the file and nothing else.
	Image string
	// ImageAlt is the front-matter "image_alt", describing Image for a reader who
	// cannot see it. Empty leaves the image decorative.
	ImageAlt string
	// ImageCaption is the front-matter "image_caption", the line shown under
	// Image: an attribution, a licence, or a disclosure. Empty renders none.
	//
	// Both belong to whichever image is used, so the builder resolves the three
	// together rather than one key at a time.
	ImageCaption string
	// Layout is the front-matter "layout" name, empty when the page sets none.
	// It names a template the theme may define; whether one does is the
	// builder's question, since this package knows the file and not the theme.
	Layout string
	// Draft marks a page excluded from a normal build.
	Draft bool
	// Meta is the parsed front matter, passed through verbatim to templates.
	Meta map[string]any
	// Body is the Markdown body with front matter stripped.
	Body []byte
}

// Collect walks root and parses every Markdown file into a Page. The returned
// pages are ordered by their source path for deterministic output.
//
// Only regular files become pages. The walk itself does not follow symlinks,
// but reading one would, so a link is skipped rather than resolved: without
// that, content/leak.md pointing at a file outside the site would be published
// as a page of it. That matters most where it is least visible, in a build run
// unattended over content somebody else can add to.
func Collect(root string) ([]*Page, error) {
	var pages []*Page
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !d.Type().IsRegular() {
			return nil
		}
		if !strings.EqualFold(filepath.Ext(path), mdExt) {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("resolving %s: %w", path, err)
		}
		page, err := parseFile(path, filepath.ToSlash(rel))
		if err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// parseFile reads one Markdown file at absPath, whose path relative to the
// content root is relSlash.
func parseFile(absPath, relSlash string) (*Page, error) {
	// #nosec G304 -- absPath comes from walking the site's own content tree.
	src, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	meta, body, err := splitFrontMatter(src)
	if err != nil {
		return nil, err
	}

	out := strings.TrimSuffix(relSlash, mdExt) + htmlExt
	page := &Page{
		SourcePath:   relSlash,
		OutputPath:   out,
		URL:          urlFor(out),
		Title:        titleFor(meta, body, relSlash),
		Description:  stringField(meta, "description"),
		Language:     stringField(meta, "language"),
		Image:        stringField(meta, "image"),
		ImageAlt:     stringField(meta, "image_alt"),
		ImageCaption: stringField(meta, "image_caption"),
		Layout:       stringField(meta, "layout"),
		Draft:        draftFor(meta),
		Meta:         meta,
		Body:         body,
	}
	return page, nil
}

// urlFor maps a slash-form output path to a site-root-relative URL. An index
// file collapses to its directory URL ("index.html" -> "/", "a/index.html" ->
// "/a/").
func urlFor(outSlash string) string {
	url := "/" + outSlash
	if strings.HasSuffix(outSlash, indexName+htmlExt) {
		return strings.TrimSuffix(url, indexName+htmlExt)
	}
	return url
}

// titleFor resolves a page title, preferring front-matter title, then the first
// H1 heading, then the file's base name.
func titleFor(meta map[string]any, body []byte, relSlash string) string {
	if t, ok := meta["title"].(string); ok && strings.TrimSpace(t) != "" {
		return t
	}
	if h := firstHeading(body); h != "" {
		return h
	}
	base := filepath.Base(relSlash)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// stringField reads a string front-matter key, trimmed. It yields "" for a key
// that is absent, blank, or written as some other type, so a caller can treat
// "the page said nothing" as one case however the author got there.
func stringField(meta map[string]any, key string) string {
	value, _ := meta[key].(string)
	return strings.TrimSpace(value)
}

// draftFor reads the boolean "draft" front-matter flag, defaulting to false.
func draftFor(meta map[string]any) bool {
	draft, _ := meta["draft"].(bool)
	return draft
}

// firstHeading returns the text of the first ATX H1 ("# ...") in the body,
// ignoring lines inside fenced code blocks. It returns "" when none is found.
func firstHeading(body []byte) string {
	inFence := false
	for _, raw := range bytes.Split(body, []byte("\n")) {
		line := strings.TrimRight(string(raw), "\r")
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		if rest, ok := strings.CutPrefix(line, "# "); ok {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

// splitFrontMatter separates a leading YAML front-matter block, delimited by
// lines containing only "---", from the Markdown body. When the source has no
// front matter it returns an empty meta map and the source unchanged; a
// malformed block is reported as an error. A block fenced any other way is not
// front matter and stays part of the body, where it renders as the text it is.
func splitFrontMatter(src []byte) (map[string]any, []byte, error) {
	meta := map[string]any{}
	body, err := frontmatter.Parse(bytes.NewReader(src), &meta, yamlFrontMatter)
	if err != nil {
		return nil, nil, fmt.Errorf("front matter: %w", err)
	}
	return meta, body, nil
}
