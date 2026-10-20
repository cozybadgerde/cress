package content

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/page"
)

func TestPublicationMetadata(t *testing.T) {
	for _, prefix := range []string{"/", "/company/"} {
		config, _, err := ParseConfig([]byte("[site]\nhost = 'https://example.org:8443/'\npath = '" + prefix + "'"))
		if err != nil {
			t.Fatal(err)
		}
		pages := []Page{
			{Data: page.Page{Canonical: config.Host + prefix + "z/", PublishedAt: "2026-01-01"}},
			{Data: page.Page{Canonical: config.Host + prefix + "a%20%26%20b/", UpdatedAt: "2026-02-03"}},
			{Draft: true, Data: page.Page{Canonical: "https://example.org/private/"}},
		}
		files, warnings, err := preparePublication(pages, config)
		if err != nil || len(warnings) != 0 || len(files) != 2 {
			t.Fatalf("files = %+v, warnings = %v, error = %v", files, warnings, err)
		}
		robots := string(files[0].Data)
		if robots != "User-agent: *\nAllow: /\n\nSitemap: "+config.Host+prefix+"sitemap.xml\n" {
			t.Fatalf("incorrect robots: %q", robots)
		}
		var document sitemap
		if err := xml.Unmarshal(files[1].Data, &document); err != nil {
			t.Fatal(err)
		}
		if document.XMLName.Space != "http://www.sitemaps.org/schemas/sitemap/0.9" ||
			len(document.URLs) != 2 || document.URLs[0].Modified != "2026-02-03" ||
			document.URLs[1].Modified != "" || document.URLs[0].Location != pages[1].Data.Canonical {
			t.Fatalf("incorrect sitemap: %+v", document)
		}
		if !strings.HasPrefix(string(files[1].Data), xml.Header) {
			t.Fatal("XML declaration missing")
		}
	}
}

func TestMissingOriginAndEmptySitemap(t *testing.T) {
	config, _, err := ParseConfig(nil)
	if err != nil {
		t.Fatal(err)
	}
	files, warnings, err := preparePublication(nil, config)
	if err != nil || len(warnings) != 1 || !strings.Contains(warnings[0], "site.host") ||
		len(files) != 1 || string(files[0].Data) != "User-agent: *\nAllow: /\n" {
		t.Fatalf("files = %+v, warnings = %v, error = %v", files, warnings, err)
	}
	config.Host = "https://example.org"
	files, warnings, err = preparePublication(nil, config)
	if err != nil || len(warnings) != 0 || len(files) != 2 ||
		!strings.HasPrefix(string(files[1].Data), xml.Header+"<urlset") {
		t.Fatalf("empty sitemap = %+v, warnings = %v, error = %v", files, warnings, err)
	}
}

func TestCanonicalURLPreparation(t *testing.T) {
	for _, test := range []struct{ config, source, want string }{
		{"", "about.md", ""},
		{"[site]\nhost = 'https://example.org/'", "index.md", "https://example.org/"},
		{"[site]\nhost = 'https://example.org'\npath = '/company/'", "about/index.md", "https://example.org/company/about/"},
		{"[site]\nhost = 'https://example.org'\npath = '/é/'", "Our & team.md",
			"https://example.org/%C3%A9/Our%20&%20team/"},
	} {
		t.Run(test.source+test.config, func(t *testing.T) {
			config, _, err := ParseConfig([]byte(test.config))
			if err != nil {
				t.Fatal(err)
			}
			item, _, err := Prepare(test.source, []byte("+++\ntitle = 'Page'\npublished_at = 2099-01-01\n+++\nBody"), config)
			if err != nil || item.Draft || item.Data.Canonical != test.want {
				t.Fatalf("canonical = %q, draft = %v, error = %v; want %q",
					item.Data.Canonical, item.Draft, err, test.want)
			}
			if config.Host != "" {
				files, _, err := preparePublication([]Page{item}, config)
				if err != nil {
					t.Fatal(err)
				}
				var document sitemap
				if err := xml.Unmarshal(files[1].Data, &document); err != nil {
					t.Fatal(err)
				}
				if len(document.URLs) != 1 || document.URLs[0].Location != test.want {
					t.Fatalf("sitemap URL changed: %+v", document)
				}
			}
		})
	}
}
