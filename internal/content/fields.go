package content

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/cozybadgerde/cress/internal/page"
)

type fields struct {
	values   map[string]any
	source   string
	kind     string
	prefix   string
	warnings *[]string
}

func (input fields) warnUnknown(known ...string) {
	for _, key := range slices.Sorted(maps.Keys(input.values)) {
		if !slices.Contains(known, key) {
			*input.warnings = append(*input.warnings, fmt.Sprintf(
				"%s: unknown %s field %q", input.source, input.kind, input.prefix+key))
		}
	}
}

func (input fields) readTable(key string) (fields, error) {
	nested := input
	nested.prefix += key + "."
	nested.values = nil
	if value, exists := input.values[key]; exists {
		table, ok := value.(map[string]any)
		if !ok {
			return nested, input.invalidField(key, "must be a table")
		}
		nested.values = table
	}
	return nested, nil
}

func (input fields) readString(key, fallback string) (string, error) {
	value, exists := input.values[key]
	if !exists {
		return fallback, nil
	}
	text, ok := value.(string)
	if !ok {
		return "", input.invalidField(key, "must be a string")
	}
	return text, nil
}

func (input fields) readNonblank(key, fallback string) (string, error) {
	text, err := input.readString(key, fallback)
	if err != nil {
		return "", input.invalidField(key, "must be a nonempty string")
	}
	if _, exists := input.values[key]; exists && strings.TrimSpace(text) == "" {
		return "", input.invalidField(key, "must be a nonempty string")
	}
	return text, nil
}

func (input fields) invalidField(key, message string) error {
	return fmt.Errorf("%s: %s%s %s", input.source, input.prefix, key, message)
}

func (input fields) readImage(prefix string) (*page.Image, error) {
	if _, exists := input.values["image"]; !exists {
		return nil, nil
	}
	image, err := input.readTable("image")
	if err != nil {
		return nil, err
	}
	image.warnUnknown("src", "alt", "caption", "focus")
	if _, exists := image.values["src"]; !exists {
		return nil, image.invalidField("src", "is required")
	}
	src, err := image.readString("src", "")
	if err != nil {
		return nil, err
	}
	alt, err := image.readString("alt", "")
	if err != nil {
		return nil, err
	}
	caption, err := image.readString("caption", "")
	if err != nil {
		return nil, err
	}
	focus, err := image.readString("focus", "")
	if err != nil {
		return nil, err
	}
	if _, exists := image.values["focus"]; exists && !slices.Contains(imageFocuses, focus) {
		return nil, image.invalidField("focus", "must be one of "+strings.Join(imageFocuses, ", "))
	}
	resolve := func(value string) (string, error) {
		if prefix == "" {
			parsed, err := parseURL(value)
			if err != nil {
				return "", err
			}
			if strings.HasPrefix(value, "//") || strings.HasPrefix(parsed.Path, "//") {
				return "", fmt.Errorf("protocol-relative URLs are not allowed")
			}
			if strings.HasPrefix(value, "/") {
				return ResolveURL(value, "/")
			}
			return value, nil
		}
		return ResolveURL(value, prefix)
	}
	parts, err := parseCaption(caption, resolve)
	if err != nil {
		return nil, image.invalidField("caption", err.Error())
	}
	if src == "" {
		return nil, nil
	}
	if _, exists := image.values["alt"]; !exists {
		return nil, image.invalidField("alt", "is required for a nonempty image source")
	}
	resolved, err := resolve(src)
	if err != nil {
		return nil, image.invalidField("src", err.Error())
	}
	return &page.Image{URL: resolved, Alt: alt, Caption: parts, Focus: focus}, nil
}

var imageFocuses = []string{"center", "top", "bottom", "left", "right",
	"top-left", "top-right", "bottom-left", "bottom-right"}
