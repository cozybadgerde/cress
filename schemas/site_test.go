package schemas_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/santhosh-tekuri/jsonschema/v6"

	"github.com/cozybadgerde/cress/internal/content"
)

func TestSiteSchemaMatchesRuntime(t *testing.T) {
	schema := compileSchema(t, "site.schema.json")
	tests := []struct {
		name, source string
		valid        bool
	}{
		{"empty", "", true},
		{"site", "[site]\ntitle = 'Business'\ndescription = ''\ncopyright = ''\nlanguage = 'de-DE'\nhost = 'https://example.org/'\npath = '/company/'", true},
		{"private language", "[site]\nlanguage = 'x-private'", true},
		{"title type", "[site]\ntitle = 1", false},
		{"title empty", "[site]\ntitle = '  '", false},
		{"site type", "site = false", false},
		{"description type", "[site]\ndescription = false", false},
		{"copyright type", "[site]\ncopyright = 1", false},
		{"language type", "[site]\nlanguage = 1", false},
		{"language empty", "[site]\nlanguage = ''", false},
		{"language underscore", "[site]\nlanguage = 'en_US'", false},
		{"language space", "[site]\nlanguage = 'en US'", false},
		{"origin type", "[site]\nhost = false", false},
		{"origin empty", "[site]\nhost = ''", false},
		{"origin relative", "[site]\nhost = 'example.org'", false},
		{"origin path", "[site]\nhost = 'https://example.org/path/'", false},
		{"origin query", "[site]\nhost = 'https://example.org?query=1'", false},
		{"origin credentials", "[site]\nhost = 'https://user@example.org/'", false},
		{"origin localhost", "[site]\nhost = 'http://localhost:8080/'", true},
		{"origin uppercase", "[site]\nhost = 'HTTPS://EXAMPLE.ORG/'", true},
		{"origin IPv6", "[site]\nhost = 'http://[::1]:8080/'", true},
		{"prefix type", "[site]\npath = false", false},
		{"prefix leading slash", "[site]\npath = 'company/'", false},
		{"prefix trailing slash", "[site]\npath = '/company'", false},
		{"prefix repeated slash", "[site]\npath = '/company//'", false},
		{"prefix dot", "[site]\npath = '/company/../'", false},
		{"prefix escape", "[site]\npath = '/%20/'", false},
		{"prefix query", "[site]\npath = '/company/?q=1'", false},
		{"prefix space", "[site]\npath = '/company name/'", false},
		{"prefix unicode", "[site]\npath = '/é/'", true},
		{"image", "[site.image]\nsrc = '/team.jpg'\nalt = ''", true},
		{"image empty", "[site.image]\nsrc = ''", true},
		{"image type", "[site]\nimage = false", false},
		{"image missing source", "[site.image]\nalt = ''", false},
		{"image missing alternative", "[site.image]\nsrc = '/team.jpg'", false},
		{"image relative source", "[site.image]\nsrc = 'team.jpg'\nalt = ''", false},
		{"image source type", "[site.image]\nsrc = false", false},
		{"image alternative type", "[site.image]\nsrc = ''\nalt = false", false},
		{"image caption", "[site.image]\nsrc = '/team.jpg'\nalt = ''\ncaption = 'By [Alex](/credits/)'", true},
		{"image caption type", "[site.image]\nsrc = ''\ncaption = false", false},
		{"highlight enabled", "[syntax]\nhighlight = true", true},
		{"highlight disabled", "[syntax]\nhighlight = false", true},
		{"highlight type", "[syntax]\nhighlight = 'yes'", false},
		{"syntax type", "syntax = 1", false},
		{"navigation", "[[navigation.main]]\nlabel = 'About'\nurl = '/about/'\n[[navigation.footer]]\nlabel = 'External'\nurl = 'https://example.net/'", true},
		{"empty navigation", "[navigation]\nmain = []\nfooter = []", true},
		{"navigation type", "navigation = 1", false},
		{"group type", "[navigation]\nmain = false", false},
		{"entry type", "[navigation]\nmain = [1]", false},
		{"missing label", "[[navigation.main]]\nurl = '/about/'", false},
		{"empty label", "[[navigation.main]]\nlabel = '  '\nurl = '/about/'", false},
		{"missing URL", "[[navigation.main]]\nlabel = 'About'", false},
		{"invalid URL type", "[[navigation.main]]\nlabel = 'About'\nurl = 1", false},
		{"invalid scheme", "[[navigation.main]]\nlabel = 'About'\nurl = 'mailto:hello@example.org'", false},
		{"protocol relative", "[[navigation.main]]\nlabel = 'About'\nurl = '//example.org/'", false},
		{"relative page", "[[navigation.main]]\nlabel = 'About'\nurl = 'about.md'", false},
		{"open customizations", "[theme.custom]\nanything = { draft = 1, list = ['a', 'b'] }", true},
		{"theme type", "theme = false", false},
		{"custom type", "[theme]\ncustom = 1", false},
		{"unknown fields", "future = 1\n[site]\ndraft = true\n[theme]\nfuture = 'unknown'", true},
		{"theme name", "[theme]\nname = 'my-theme-2'", true},
		{"theme name empty", "[theme]\nname = ''", false},
		{"theme name traversal", "[theme]\nname = '../other'", false},
		{"theme name case", "[theme]\nname = 'MyTheme'", false},
		{"brand", "[brand]\ncolor = '#123AbC'\nfavicon = '/icon.svg'\n[brand.logo]\nsrc = '/logo.svg'\nalt = ''", true},
		{"brand type", "brand = false", false},
		{"brand color type", "[brand]\ncolor = 1", false},
		{"brand color empty", "[brand]\ncolor = ''", false},
		{"brand color short", "[brand]\ncolor = '#fff'", false},
		{"brand color digits", "[brand]\ncolor = '#ffffff;'", false},
		{"brand favicon empty", "[brand]\nfavicon = ''", false},
		{"brand favicon relative", "[brand]\nfavicon = 'icon.svg'", false},
		{"brand logo alternative", "[brand.logo]\nsrc = '/logo.svg'", false},
		{"brand logo type", "[brand]\nlogo = false", false},
		{"brand logo empty", "[brand.logo]\nsrc = ''\nalt = ''", false},
		{"profile", "[profile]\ntype = 'NGO'\nname = 'Club'\ntelephone = '+49 1'\nemail = 'a@example.org'\n" +
			"[profile.address]\ncity = 'Berlin'\n[[profile.hours]]\ndays = ['mo', 'su']\nopens = '08:00'\ncloses = '23:59'", true},
		{"profile type", "profile = 1", false},
		{"profile type name", "[profile]\ntype = 'local business'\nname = 'A'", false},
		{"profile name empty", "[profile]\nname = ' '", false},
		{"profile address type", "[profile]\nname = 'A'\naddress = 1", false},
		{"profile hours type", "[profile]\nname = 'A'\nhours = 1", false},
		{"profile hours day", "[profile]\nname = 'A'\n[[profile.hours]]\ndays = ['monday']\nopens = '08:00'\ncloses = '18:00'", false},
		{"profile hours duplicate", "[profile]\nname = 'A'\n[[profile.hours]]\ndays = ['mo', 'mo']\nopens = '08:00'\ncloses = '18:00'", false},
		{"profile hours empty", "[profile]\nname = 'A'\n[[profile.hours]]\ndays = []\nopens = '08:00'\ncloses = '18:00'", false},
		{"profile hours time", "[profile]\nname = 'A'\n[[profile.hours]]\ndays = ['mo']\nopens = '8:00'\ncloses = '18:00'", false},
		{"list per page", "[list]\nper_page = 0", true},
		{"list per page negative", "[list]\nper_page = -1", false},
		{"list per page type", "[list]\nper_page = '10'", false},
		{"list type", "list = 1", false},
		{"profile hours missing", "[profile]\nname = 'A'\n[[profile.hours]]\ndays = ['mo']\nopens = '08:00'", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fields := make(map[string]any)
			if err := toml.Unmarshal([]byte(test.source), &fields); err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			instance, err := jsonschema.UnmarshalJSON(strings.NewReader(string(encoded)))
			if err != nil {
				t.Fatal(err)
			}
			schemaErr := schema.Validate(instance)
			_, _, runtimeErr := content.ParseConfig([]byte(test.source))
			if (schemaErr == nil) != test.valid || (runtimeErr == nil) != test.valid {
				t.Fatalf("want valid %v; schema error = %v; runtime error = %v",
					test.valid, schemaErr, runtimeErr)
			}
		})
	}
	for _, invalid := range []any{nil, "configuration", []any{"site"}} {
		if err := schema.Validate(invalid); err == nil {
			t.Errorf("schema accepted non-object %#v", invalid)
		}
	}
}
