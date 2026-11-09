package content

import (
	"reflect"
	"strings"
	"testing"
)

func TestProfilePreparation(t *testing.T) {
	config, _, err := ParseConfig([]byte("[site]\ntitle = 'Club'\ndescription = 'Sports'\n" +
		"host = 'https://example.org'\npath = '/club/'\n[brand.logo]\nsrc = '/logo.svg'\nalt = ''\n" +
		"[profile]\ntype = 'SportsClub'\nemail = 'hi@example.org'\ntelephone = ' '\n" +
		"[profile.address]\ncity = 'Berlin'\n[[profile.hours]]\ndays = ['sa', 'su']\nopens = '10:00'\ncloses = '02:00'\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"@context": "https://schema.org", "@type": "SportsClub", "name": "Club",
		"url": "https://example.org/club/", "description": "Sports", "logo": "https://example.org/club/logo.svg",
		"email":   "hi@example.org",
		"address": map[string]any{"@type": "PostalAddress", "addressLocality": "Berlin"},
		"openingHoursSpecification": []any{map[string]any{
			"@type": "OpeningHoursSpecification", "dayOfWeek": []string{"Saturday", "Sunday"},
			"opens": "10:00", "closes": "02:00",
		}},
	}
	if !reflect.DeepEqual(config.Profile, want) {
		t.Fatalf("profile = %#v", config.Profile)
	}
	for source, wantData := range map[string]bool{"index.md": true, "about.md": false} {
		item, _, err := Prepare(source, []byte("+++\ntitle = 'Page'\n+++\n"), config)
		if err != nil || (item.Data.StructuredData != nil) != wantData {
			t.Fatalf("%s structured data = %v, error = %v", source, item.Data.StructuredData, err)
		}
	}
	config, _, err = ParseConfig([]byte("[profile]\nname = 'Club'\n"))
	if err != nil || config.Profile != nil {
		t.Fatalf("profile without host = %v, error = %v", config.Profile, err)
	}
}

func TestProfileErrors(t *testing.T) {
	for _, test := range []struct{ source, want string }{
		{"[profile]\ntype = 'Bakery'", "profile.name is required"},
		{"[site]\ntitle = 'A'\n[profile]\nemail = 'Alex <a@example.org>'", "profile.email must be a plain email"},
		{"[site]\ntitle = 'A'\n[profile]\nemail = 'nobody'", "profile.email must be a plain email"},
		{"[site]\ntitle = 'A'\n[[profile.hours]]\ndays = ['mo']\nopens = '24:00'\ncloses = '18:00'", "profile.hours[0].opens"},
	} {
		_, _, err := ParseConfig([]byte(test.source))
		if err == nil || !strings.Contains(err.Error(), test.want) {
			t.Errorf("%q: error = %v; want %q", test.source, err, test.want)
		}
	}
}
