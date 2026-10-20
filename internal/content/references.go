package content

import (
	"fmt"
	"net/url"
	"path"
	"strings"
	"unicode"

	"github.com/yuin/goldmark/util"
)

type references struct {
	prefix string
	source string
	files  map[string]reference
}

type reference struct {
	url  string
	page bool
}

func (references references) resolve(value string) (string, error) {
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("invalid reference %q: %w", value, err)
	}
	if parsed.IsAbs() || parsed.Host != "" || strings.HasPrefix(value, "//") {
		return value, nil
	}
	if strings.HasPrefix(value, "/") {
		return ResolveURL(value, references.prefix)
	}
	if parsed.Path == "" {
		return value, nil
	}
	if strings.ContainsAny(parsed.Path, "\\\x00") ||
		strings.IndexFunc(parsed.Path, unicode.IsControl) >= 0 {
		return "", fmt.Errorf("unsafe source reference %q", value)
	}
	source := path.Join(path.Dir(references.source), parsed.Path)
	if !strings.HasPrefix(source, "content/") && !strings.HasPrefix(source, "static/") {
		return "", fmt.Errorf("source reference %q escapes content and static boundaries", value)
	}
	target, exists := references.files[source]
	if !exists {
		return "", fmt.Errorf("source reference %q does not map to a published page or asset", value)
	}
	resolved := target.url
	if parsed.RawQuery != "" || parsed.ForceQuery {
		resolved += "?" + parsed.RawQuery
	}
	if fragment := strings.IndexByte(value, '#'); fragment >= 0 {
		resolved += value[fragment:]
	}
	return resolved, nil
}

func (references references) resolveImage(value string) (string, error) {
	parsed, err := url.Parse(value)
	if err != nil {
		return "", err
	}
	if !parsed.IsAbs() && parsed.Host == "" && !strings.HasPrefix(value, "/") {
		target := references.files[path.Join(path.Dir(references.source), parsed.Path)]
		if target.page {
			return "", fmt.Errorf("image reference %q must name a published asset", value)
		}
	}
	return references.resolve(value)
}

func decodeDestination(value []byte) string {
	return string(util.ResolveEntityNames(util.ResolveNumericReferences(util.UnescapePunctuations(value))))
}
