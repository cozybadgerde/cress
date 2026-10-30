package content

import (
	"encoding/hex"
	"fmt"
	"math"

	"github.com/cozybadgerde/cress/internal/page"
)

func (input fields) readBrand(prefix string) (page.Brand, error) {
	var brand page.Brand
	table, err := input.readTable("brand")
	if err != nil {
		return brand, err
	}
	table.warnUnknown("color", "logo", "favicon")
	brand.Color, err = table.readString("color", "")
	if err != nil {
		return brand, err
	}
	if _, exists := table.values["color"]; exists {
		if len(brand.Color) != 7 || brand.Color[0] != '#' {
			return brand, table.invalidField("color", "must be a six-digit #RRGGBB color")
		}
		rgb, err := hex.DecodeString(brand.Color[1:])
		if err != nil {
			return brand, table.invalidField("color", "must be a six-digit #RRGGBB color")
		}
		if ratio := contrastWithWhite(rgb); ratio < 4.5 {
			*input.warnings = append(*input.warnings, fmt.Sprintf(
				"%s: %scolor %s has contrast %.2f:1 against white (below 4.5:1); this reference comparison does not assess the selected theme; check contrast against its actual text and background colors",
				input.source, table.prefix, brand.Color, ratio))
		}
	}
	if _, exists := table.values["favicon"]; exists {
		value, err := table.readString("favicon", "")
		if err != nil {
			return brand, err
		}
		brand.Favicon, err = ResolveURL(value, prefix)
		if err != nil {
			return brand, table.invalidField("favicon", err.Error())
		}
	}
	if _, exists := table.values["logo"]; exists {
		logo, err := table.readTable("logo")
		if err != nil {
			return brand, err
		}
		logo.warnUnknown("src", "alt")
		for _, required := range []string{"src", "alt"} {
			if _, exists := logo.values[required]; !exists {
				return brand, logo.invalidField(required, "is required")
			}
		}
		src, err := logo.readString("src", "")
		if err != nil {
			return brand, err
		}
		url, err := ResolveURL(src, prefix)
		if err != nil {
			return brand, logo.invalidField("src", err.Error())
		}
		alt, err := logo.readString("alt", "")
		if err != nil {
			return brand, err
		}
		brand.Logo = &page.Image{URL: url, Alt: alt}
	}
	return brand, nil
}

// contrastWithWhite returns the WCAG 2 contrast ratio of an sRGB color against white.
func contrastWithWhite(rgb []byte) float64 {
	linear := func(component byte) float64 {
		channel := float64(component) / 255
		if channel <= 0.04045 {
			return channel / 12.92
		}
		return math.Pow((channel+0.055)/1.055, 2.4)
	}
	luminance := .2126*linear(rgb[0]) + .7152*linear(rgb[1]) + .0722*linear(rgb[2])
	return 1.05 / (luminance + .05)
}
