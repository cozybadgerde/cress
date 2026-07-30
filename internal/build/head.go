package build

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

// headTemplate renders the metadata block a theme emits with {{ .Page.Head }}.
//
// It is an html/template rather than string concatenation because every value in
// it is the author's text landing in an attribute: a quotation mark in a
// description would otherwise close the attribute early and put the rest of the
// sentence into the markup.
//
// Each tag is omitted when it has nothing to say. An empty description is worse
// than none, and a canonical link needs an absolute URL, which a site without a
// base_url cannot produce.
var headTemplate = template.Must(template.New("head").Parse(
	`{{ with .Description }}<meta name="description" content="{{ . }}" />
{{ end }}` +
		`{{ with .Canonical }}<link rel="canonical" href="{{ . }}" />
{{ end }}` +
		`{{ if .NoIndex }}<meta name="robots" content="noindex" />
{{ end }}`))

// headData is what headTemplate renders. It is not the template contract: a
// theme sees the rendered result as PageView.Head, never these fields.
type headData struct {
	Description string
	Canonical   string
	// NoIndex keeps a page out of search results. Only the synthesized 404 sets
	// it, and nothing configures it: a page that does not exist is the one case
	// cress can decide on the author's behalf. Crawl policy in general belongs to
	// robots.txt rather than to a key here.
	NoIndex bool
}

// renderHead builds the metadata block for one page. The result carries no
// trailing newline, so a theme's own indentation decides where the next line
// starts and a block with nothing to say leaves no blank line behind.
func renderHead(data headData) (template.HTML, error) {
	var buf bytes.Buffer
	if err := headTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("rendering head metadata: %w", err)
	}
	// #nosec G203 -- every value went through html/template's escaping above.
	return template.HTML(strings.TrimRight(buf.String(), "\n")), nil
}
