package content

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"

	"github.com/cozybadgerde/cress/internal/page"
)

// Config contains validated site settings and prepared shared presentation data.
type Config struct {
	Site        page.Site
	Description string
	Image       *page.Image
	Host        string
	Highlight   bool
}

// ParseConfig validates TOML site configuration and resolves shared URLs.
func ParseConfig(source []byte) (Config, []string, error) {
	config := Config{Site: page.Site{Language: "en", URL: "/"}, Highlight: true}
	var warnings []string
	if !utf8.Valid(source) {
		return config, warnings, fmt.Errorf("cress.toml: configuration must be UTF-8")
	}
	var values map[string]any
	if err := toml.Unmarshal(source, &values); err != nil {
		return config, warnings, fmt.Errorf("cress.toml: parse configuration: %w", err)
	}
	input := fields{values: values, source: "cress.toml", kind: "configuration", warnings: &warnings}
	input.warnUnknown("site", "navigation", "theme", "markdown")
	site, err := input.readTable("site")
	if err != nil {
		return config, warnings, err
	}
	site.warnUnknown("title", "description", "copyright", "language", "host", "path", "image")
	config.Site.Title, err = site.readNonblank("title", "")
	if err != nil {
		return config, warnings, err
	}
	config.Description, err = site.readString("description", "")
	if err != nil {
		return config, warnings, err
	}
	config.Site.Copyright, err = site.readString("copyright", "")
	if err != nil {
		return config, warnings, err
	}
	config.Site.Language, err = site.readString("language", "en")
	if err != nil {
		return config, warnings, err
	}
	if config.Site.Language == "" || strings.Contains(config.Site.Language, "_") {
		return config, warnings, site.invalidField("language", "must be a valid BCP 47 language tag")
	}
	if _, err := language.Parse(config.Site.Language); err != nil {
		return config, warnings, site.invalidField("language", "must be a valid BCP 47 language tag")
	}
	prefix, err := site.readString("path", "/")
	if err != nil {
		return config, warnings, err
	}
	config.Site.URL, err = validatePath(prefix)
	if err != nil {
		return config, warnings, site.invalidField("path", err.Error())
	}
	if _, exists := site.values["host"]; exists {
		host, err := site.readString("host", "")
		if err != nil {
			return config, warnings, err
		}
		config.Host, err = validateOrigin(host)
		if err != nil {
			return config, warnings, site.invalidField("host", err.Error())
		}
	}
	config.Image, err = site.readImage(config.Site.URL)
	if err != nil {
		return config, warnings, err
	}
	config.Site.Navigation, err = input.readNavigation(config.Site.URL)
	if err != nil {
		return config, warnings, err
	}
	theme, err := input.readTable("theme")
	if err != nil {
		return config, warnings, err
	}
	theme.warnUnknown("custom")
	custom, err := theme.readTable("custom")
	if err != nil {
		return config, warnings, err
	}
	config.Site.Custom = custom.values
	markdown, err := input.readTable("markdown")
	if err != nil {
		return config, warnings, err
	}
	markdown.warnUnknown("highlight")
	if value, exists := markdown.values["highlight"]; exists {
		enabled, ok := value.(bool)
		if !ok {
			return config, warnings, markdown.invalidField("highlight", "must be a boolean")
		}
		config.Highlight = enabled
	}
	return config, warnings, nil
}

func (input fields) readNavigation(prefix string) (page.Navigation, error) {
	var prepared page.Navigation
	navigation, err := input.readTable("navigation")
	if err != nil {
		return prepared, err
	}
	navigation.warnUnknown("main", "footer")
	prepared.Main, err = navigation.readLinks("main", prefix)
	if err != nil {
		return prepared, err
	}
	prepared.Footer, err = navigation.readLinks("footer", prefix)
	return prepared, err
}

func (input fields) readLinks(key, prefix string) ([]page.Link, error) {
	value, exists := input.values[key]
	if !exists {
		return nil, nil
	}
	items, ok := value.([]any)
	if !ok {
		return nil, input.invalidField(key, "must be an array of tables")
	}
	var links []page.Link
	for index, item := range items {
		values, ok := item.(map[string]any)
		if !ok {
			return nil, input.invalidField(fmt.Sprintf("%s[%d]", key, index), "must be a table")
		}
		entry := fields{
			values: values, source: input.source, kind: input.kind,
			prefix: fmt.Sprintf("%s%s[%d].", input.prefix, key, index), warnings: input.warnings,
		}
		entry.warnUnknown("label", "url")
		label, err := entry.readNonblank("label", "")
		if err != nil {
			return nil, err
		}
		if label == "" {
			return nil, entry.invalidField("label", "is required")
		}
		destination, err := entry.readString("url", "")
		if err != nil {
			return nil, err
		}
		resolved, err := ResolveURL(destination, prefix)
		if err != nil {
			return nil, entry.invalidField("url", err.Error())
		}
		links = append(links, page.Link{Label: label, URL: resolved})
	}
	return links, nil
}

// LoadConfig loads optional cress.toml without treating read failures as absence.
func LoadConfig(ctx context.Context, directory string) (config Config, warnings []string, loadErr error) {
	if err := ctx.Err(); err != nil {
		return config, nil, err
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return config, nil, fmt.Errorf("open site directory: %w", err)
	}
	defer func() {
		if err := root.Close(); err != nil {
			loadErr = errors.Join(loadErr, fmt.Errorf("close site directory: %w", err))
		}
	}()
	info, err := root.Lstat("cress.toml")
	if errors.Is(err, fs.ErrNotExist) {
		return ParseConfig(nil)
	}
	if err != nil {
		return config, nil, fmt.Errorf("inspect cress.toml: %w", err)
	}
	if !info.Mode().IsRegular() {
		return config, nil, fmt.Errorf("cress.toml must be a regular file, not a symlink or directory")
	}
	source, err := root.ReadFile("cress.toml")
	if err != nil {
		return config, nil, fmt.Errorf("read cress.toml: %w", err)
	}
	return ParseConfig(source)
}
