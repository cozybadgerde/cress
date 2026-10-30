package theme

import (
	"context"
	"io/fs"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestBuiltinPackagesAndExactLabels(t *testing.T) {
	for _, name := range []string{"poppy", "clover", "anemone"} {
		source, err := fs.Sub(embedded, "builtin/"+name)
		if err != nil {
			t.Fatal(err)
		}
		for _, language := range []string{"en", "de", "en-US"} {
			pkg, warnings, err := loadPackage(context.Background(), source, Request{
				Name: name, Language: language, Prefix: "/company/",
			})
			if err != nil || len(warnings) != 0 || len(pkg.Assets) != 1 {
				t.Fatalf("%s/%s: %v, %v", name, language, err, warnings)
			}
			label, err := pkg.Label("skip")
			if language == "en-US" {
				if err == nil {
					t.Fatal("regional language fell back")
				}
			} else if err != nil || label == "" {
				t.Fatalf("missing translation: %v", err)
			}
			if value, err := pkg.AssetURL("site.css"); err != nil || value != "/company/themes/"+name+"/site.css" {
				t.Fatalf("asset: %s, %v", value, err)
			}
		}
	}
}

func TestBuiltinDefaultPaletteContrast(t *testing.T) {
	variable := regexp.MustCompile(`--([a-z-]+):\s*(#[0-9a-f]{6})`)
	for _, name := range []string{"poppy", "clover", "anemone"} {
		data, err := embedded.ReadFile("builtin/" + name + "/public/site.css")
		if err != nil {
			t.Fatal(err)
		}
		palette := make(map[string]string)
		for _, match := range variable.FindAllStringSubmatch(string(data), -1) {
			palette[match[1]] = match[2]
		}
		for _, background := range []string{palette["background"], palette["surface"]} {
			for _, foreground := range []string{palette["text"], palette["link"]} {
				if ratio := contrast(foreground, background); ratio < 4.5 {
					t.Errorf("%s text %s/%s contrast %.2f < 4.5", name, foreground, background, ratio)
				}
			}
			if ratio := contrast(palette["border"], background); ratio < 3 {
				t.Errorf("%s boundary contrast %.2f < 3", name, ratio)
			}
		}
		for _, token := range []string{"syntax-keyword", "syntax-string", "syntax-comment"} {
			if ratio := contrast(palette[token], palette["surface"]); ratio < 4.5 {
				t.Errorf("%s syntax contrast %.2f < 4.5", name, ratio)
			}
		}
		for _, invariant := range []string{"a:focus-visible", ".skip-link", "flex-wrap: wrap", "max-width: 100%", "white-space: pre-wrap"} {
			if !strings.Contains(string(data), invariant) {
				t.Errorf("%s missing %s", name, invariant)
			}
		}
	}
}

func contrast(first, second string) float64 {
	luminance := func(value string) float64 {
		rgb, err := strconv.ParseUint(strings.TrimPrefix(value, "#"), 16, 32)
		if err != nil || len(value) != 7 {
			panic("invalid test palette: " + value)
		}
		linear := func(component uint64) float64 {
			channel := float64(component) / 255
			if channel <= 0.04045 {
				return channel / 12.92
			}
			return math.Pow((channel+0.055)/1.055, 2.4)
		}
		return .2126*linear((rgb>>16)&255) + .7152*linear((rgb>>8)&255) + .0722*linear(rgb&255)
	}
	light, dark := luminance(first), luminance(second)
	if light < dark {
		light, dark = dark, light
	}
	return (light + .05) / (dark + .05)
}
