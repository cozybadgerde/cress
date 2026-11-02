package theme

import (
	"bytes"
	"context"
	"html/template"
	"io/fs"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/page"
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
		for _, background := range []string{palette["background"], palette["surface"]} {
			if ratio := contrast(palette["muted"], background); ratio < 4.5 {
				t.Errorf("%s metadata contrast %.2f < 4.5", name, ratio)
			}
		}
		for _, invariant := range []string{"a:focus-visible", ".skip-link", "flex-wrap: wrap", "max-width: 100%", "white-space: pre-wrap"} {
			if !strings.Contains(string(data), invariant) {
				t.Errorf("%s missing %s", name, invariant)
			}
		}
	}
}

func TestBuiltinOptionalHeadingDates(t *testing.T) {
	for _, name := range []string{"poppy", "clover", "anemone"} {
		t.Run(name, func(t *testing.T) { testOptionalHeadingDates(t, name) })
	}
}

func testOptionalHeadingDates(t *testing.T, name string) {
	t.Helper()
	source, err := fs.Sub(embedded, "builtin/"+name)
	if err != nil {
		t.Fatal(err)
	}
	pkg, _, err := loadPackage(context.Background(), source, Request{Language: "en", Prefix: "/"})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := template.New("shared").Funcs(template.FuncMap{
		"asset": pkg.AssetURL, "label": pkg.Label,
	}).Parse(pkg.Templates["shared.gohtml"])
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, published, updated string
		count                    int
	}{
		{"absent", "", "", 0},
		{"published", "2026-10-21", "", 1},
		{"updated", "", "2026-10-23", 1},
		{"both", "2026-10-21", "2026-10-23", 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			model := page.Page{Title: "<Services>", PublishedAt: test.published, UpdatedAt: test.updated}
			if err := parsed.ExecuteTemplate(&output, "page-heading", model); err != nil {
				t.Fatal(err)
			}
			html := output.String()
			if !strings.Contains(html, "<h1>&lt;Services&gt;</h1>") ||
				strings.Count(html, "<time ") != test.count {
				t.Fatalf("heading lost escaped title or optional dates: %s", html)
			}
			if strings.Contains(html, `class="dates"`) != (test.count > 0) ||
				strings.Contains(html, `class="date-separator"`) != (test.count == 2) {
				t.Fatalf("empty metadata or orphaned separator: %s", html)
			}
			if test.count > 0 && strings.Index(html, "<time ") > strings.Index(html, "<h1>") {
				t.Fatalf("dates follow the heading: %s", html)
			}
			for _, date := range []string{test.published, test.updated} {
				if date != "" && !strings.Contains(html, `<time datetime="`+date+`">`+date+`</time>`) {
					t.Fatalf("lost machine-readable date %s: %s", date, html)
				}
			}
		})
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
