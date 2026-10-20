package content

import (
	"encoding/hex"

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
		if _, err := hex.DecodeString(brand.Color[1:]); err != nil {
			return brand, table.invalidField("color", "must be a six-digit #RRGGBB color")
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
