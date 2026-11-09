package content_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/content"
	"github.com/cozybadgerde/cress/internal/page"
)

func defaultConfig(t *testing.T) content.Config {
	t.Helper()
	config, _, err := content.ParseConfig(nil)
	if err != nil {
		t.Fatal(err)
	}
	return config
}

func TestParseConfig(t *testing.T) {
	config, warnings, err := content.ParseConfig([]byte(`
[site]
title = "Business"
description = "Shared description"
copyright = "<Copyright>"
language = "de-CH"
host = "https://example.org:8443/"
path = "/company/"

[site.image]
src = "/team.jpg"
alt = "Our team"

[[navigation.main]]
label = "About"
url = "/about/?view=full#team"

[[navigation.main]]
label = "External"
url = "https://example.net/"

[[navigation.footer]]
label = "Contact"
url = "/contact/"

[theme.custom]
title = { nested = ["kept", "as data"] }

[list]
per_page = 0
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	want := content.Config{
		Site: page.Site{
			Title: "Business", Copyright: "<Copyright>", Language: "de-CH", URL: "/company/",
			Navigation: page.Navigation{
				Main: []page.Link{
					{Label: "About", URL: "/company/about/?view=full#team"},
					{Label: "External", URL: "https://example.net/"},
				},
				Footer: []page.Link{{Label: "Contact", URL: "/company/contact/"}},
			},
			Custom: map[string]any{"title": map[string]any{"nested": []any{"kept", "as data"}}},
			Syntax: page.Syntax{Highlight: true},
		},
		Description: "Shared description", Host: "https://example.org:8443",
		Image: &page.Image{URL: "/company/team.jpg", Canonical: "https://example.org:8443/company/team.jpg", Alt: "Our team"},
	}
	if !reflect.DeepEqual(config, want) {
		t.Fatalf("config = %#v, want %#v", config, want)
	}
}

func TestConfigDefaults(t *testing.T) {
	config := defaultConfig(t)
	want := content.Config{Site: page.Site{Language: "en", URL: "/", Syntax: page.Syntax{Highlight: true}}, ListPerPage: 10}
	if !reflect.DeepEqual(config, want) {
		t.Fatalf("config = %#v, want %#v", config, want)
	}
}

func TestSyntaxHighlightingSetting(t *testing.T) {
	for _, test := range []struct {
		name   string
		source string
		want   bool
	}{
		{"default", "", true},
		{"enabled", "[syntax]\nhighlight = true", true},
		{"disabled", "[syntax]\nhighlight = false", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			config, warnings, err := content.ParseConfig([]byte(test.source))
			if err != nil || len(warnings) != 0 {
				t.Fatalf("configuration failed: %v, %v", err, warnings)
			}
			prepared, _, err := content.Prepare("code.md",
				[]byte("+++\ntitle = 'Code'\n+++\n```go\nvar value = true\n```\n"), config)
			if err != nil {
				t.Fatal(err)
			}
			if prepared.Data.Site.Syntax.Highlight != test.want ||
				strings.Contains(string(prepared.Data.Body), "<span") != test.want {
				t.Fatalf("theme setting and token output disagree: %+v", prepared.Data)
			}
		})
	}
}

func TestInvalidConfig(t *testing.T) {
	tests := []struct{ source, diagnostic string }{
		{"site = 'wrong'", "site must be a table"},
		{"[site]\ntitle = ''", "site.title"},
		{"[site]\ntitle = false", "site.title"},
		{"[site]\ndescription = 1", "site.description"},
		{"[site]\ncopyright = false", "site.copyright"},
		{"[site]\nlanguage = ''", "site.language"},
		{"[site]\nlanguage = 'en_US'", "site.language"},
		{"[site]\nlanguage = 'not a tag'", "site.language"},
		{"[site]\npath = false", "site.path"},
		{"[site]\npath = '/company'", "site.path"},
		{"[site]\npath = 'company/'", "site.path"},
		{"[site]\npath = '/company//'", "site.path"},
		{"[site]\npath = '/./'", "site.path"},
		{"[site]\npath = '/../'", "site.path"},
		{"[site]\npath = '/%20/'", "site.path"},
		{"[site]\npath = '/space here/'", "site.path"},
		{"[site]\nhost = ''", "site.host"},
		{"[site]\nhost = 'example.org'", "site.host"},
		{"[site]\nhost = 'https://example.org/path/'", "site.host"},
		{"[site]\nhost = 'https://example.org?'", "site.host"},
		{"[site]\nhost = 'https://example.org#'", "site.host"},
		{"[site]\nhost = 'https://user:pass@example.org'", "site.host"},
		{"[site]\nhost = 'https://example.org:65536'", "site.host"},
		{"[site]\nhost = 'https://example.org:'", "site.host"},
		{"[site]\nhost = 'https://[not-ip]/'", "site.host"},
		{"[site]\nimage = false", "site.image"},
		{"[site.image]\nalt = 'missing source'", "site.image.src"},
		{"[site.image]\nsrc = '/image.jpg'", "site.image.alt"},
		{"[site.image]\nsrc = '/image.jpg'\nalt = 1", "site.image.alt"},
		{"[site.image]\nsrc = 'image.jpg'\nalt = ''", "site.image.src"},
		{"[navigation]\nmain = false", "navigation.main"},
		{"[navigation]\nmain = [1]", "navigation.main[0]"},
		{"[[navigation.main]]\nurl = '/about/'", "navigation.main[0].label"},
		{"[[navigation.main]]\nlabel = '  '\nurl = '/about/'", "navigation.main[0].label"},
		{"[[navigation.main]]\nlabel = 'About'", "navigation.main[0].url"},
		{"[[navigation.main]]\nlabel = 'About'\nurl = '//example.org/'", "navigation.main[0].url"},
		{"theme = false", "theme"},
		{"[theme]\ncustom = 1", "theme.custom"},
		{"[theme]\ncustom = []", "theme.custom"},
		{"syntax = false", "syntax must be a table"},
		{"[syntax]\nhighlight = 'true'", "syntax.highlight must be a boolean"},
		{"[site.image]\nsrc = ''\ncaption = false", "site.image.caption"},
		{"[site.image]\nsrc = ''\ncaption = '[Credit](relative.md)'", "site.image.caption"},
		{"[site", "parse configuration"},
		{"\xff", "UTF-8"},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			_, _, err := content.ParseConfig([]byte(test.source))
			if err == nil || !strings.Contains(err.Error(), "cress.toml:") ||
				!strings.Contains(err.Error(), test.diagnostic) {
				t.Fatalf("error = %v, want source and %q", err, test.diagnostic)
			}
		})
	}
}

func TestConfigWarningsAndOpenCustomizations(t *testing.T) {
	_, warnings, err := content.ParseConfig([]byte(`
brand = { future = "unknown" }
[site]
unknown = 1
[site.image]
src = ""
caption = "future"
[navigation]
extra = []
[[navigation.main]]
label = "About"
url = "/about/"
target = "future"
[theme]
future = "unknown"
[theme.custom]
site = { nonsense = 1 }
`))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		`cress.toml: unknown configuration field "site.unknown"`,
		`cress.toml: unknown configuration field "navigation.extra"`,
		`cress.toml: unknown configuration field "navigation.main[0].target"`,
		`cress.toml: unknown configuration field "theme.future"`,
		`cress.toml: unknown configuration field "brand.future"`,
	}
	if !reflect.DeepEqual(warnings, want) {
		t.Fatalf("warnings = %v, want %v", warnings, want)
	}
}

func TestBrandColorContrastWarning(t *testing.T) {
	tests := []struct {
		color string
		warn  bool
	}{{"#164b81", false}, {"#767676", false}, {"#777777", true}, {"#F5D000", true}}
	for _, test := range tests {
		t.Run(test.color, func(t *testing.T) {
			config, warnings, err := content.ParseConfig([]byte("[brand]\ncolor = '" + test.color + "'"))
			if err != nil || config.Site.Brand.Color != test.color {
				t.Fatalf("color = %q, error = %v", config.Site.Brand.Color, err)
			}
			if warned := len(warnings) == 1 && strings.Contains(warnings[0], "brand.color "+test.color+" has contrast"); warned != test.warn || len(warnings) > 1 {
				t.Fatalf("warnings = %v, want warning %v", warnings, test.warn)
			}
			if test.warn && (!strings.Contains(warnings[0], "against white (below 4.5:1)") ||
				!strings.Contains(warnings[0], "this reference comparison does not assess the selected theme") ||
				!strings.Contains(warnings[0], "check contrast against its actual text and background colors")) {
				t.Fatalf("warning must explain its reference-only scope and theme-specific review: %q", warnings[0])
			}
		})
	}
}

func TestSupportedLanguagesAndOrigins(t *testing.T) {
	for _, tag := range []string{"en", "de-DE", "zh-Hant-TW", "EN-us", "x-private", "sr-Latn"} {
		t.Run(tag, func(t *testing.T) {
			config, _, err := content.ParseConfig([]byte("[site]\nlanguage = '" + tag + "'"))
			if err != nil || config.Site.Language != tag {
				t.Fatalf("language = %q, error = %v", config.Site.Language, err)
			}
		})
	}
	for _, origin := range []string{"https://example.org", "http://localhost:8080", "https://[::1]:443", "https://example.org/"} {
		t.Run(origin, func(t *testing.T) {
			config, _, err := content.ParseConfig([]byte("[site]\nhost = '" + origin + "'"))
			if err != nil || config.Host != strings.TrimSuffix(origin, "/") {
				t.Fatalf("host = %q, error = %v", config.Host, err)
			}
		})
	}
}
