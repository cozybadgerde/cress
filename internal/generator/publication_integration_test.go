package generator_test

import (
	"context"
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/generator"
)

func TestBuildListsAndPublication(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, prefix := range []string{"/", "/company/"} {
		t.Run(prefix, func(t *testing.T) {
			site := createSite(t, map[string]string{
				"cress.toml": "[site]\nhost = 'https://example.org:8443/'\npath = '" + prefix + "'",
				"content/index.md": "+++\ntitle = 'Home'\ntype = 'landing'\n+++\n" +
					"[Journal](blog/list.md)",
				"content/blog/list.md": "+++\ntitle = 'Journal'\ntype = 'list'\n+++\nJournal intro",
				"content/blog/alphabetical/index.md": "+++\ntitle = 'Alphabetical'\ntype = 'list'\n" +
					"[list]\norder = 'title'\nreverse = true\n+++\nAlphabetical intro",
				"content/blog/first.md": "+++\ntitle = '<First>'\npublished_at = 2026-10-20\n" +
					"updated_at = '2026-10-21'\nexcerpt = '<Excerpt>'\n[image]\n" +
					"src = 'first.assets/seed.svg'\nalt = 'Seed'\ncaption = '[Credits](../contact.md)'\n" +
					"+++\nArticle body must not appear in a list",
				"content/blog/first.assets/seed.svg": "<svg/>",
				"content/blog/second/index.md":       "+++\ntitle = 'Second'\npublished_at = '2026-09-01'\n+++\nSecond body",
				"content/blog/undated.md":            "+++\ntitle = 'Undated'\n+++\nUndated body",
				"content/blog/index.md":              "+++\ntitle = 'Section root'\n+++\nNot an entry",
				"content/blog/landing.md":            "+++\ntitle = 'Section landing'\ntype = 'landing'\n+++",
				"content/blog/deep/article.md":       "+++\ntitle = 'Deep article'\n+++",
				"content/blog/draft.md":              "+++\ndraft = true\n+++\nPrivate",
				"content/blog/empty/list.md":         "+++\ntitle = 'Empty'\ntype = 'list'\n+++\nEmpty intro",
				"content/contact.md":                 "+++\ntitle = 'Contact'\n+++",
			})
			result, err := generator.Build(context.Background(), site)
			if err != nil || result.Pages != 12 || len(result.Warnings) != 2 ||
				!strings.Contains(result.Warnings[0], "content/blog/undated.md: missing published_at; omitted from the feed of content/blog/alphabetical/index.md") ||
				!strings.Contains(result.Warnings[1], "content/blog/undated.md") ||
				!strings.Contains(result.Warnings[1], "content/blog/list.md") {
				t.Fatalf("result = %+v, error = %v", result, err)
			}
			read := func(name string) string {
				t.Helper()
				data, err := os.ReadFile(filepath.Join(site, "public", name))
				if err != nil {
					t.Fatal(err)
				}
				return string(data)
			}
			list := read("blog/list/index.html")
			for _, want := range []string{
				`<link rel="canonical" href="https://example.org:8443` + prefix + `blog/list/">`,
				"<p>Journal intro</p>", `<a href="` + prefix + `blog/first/">&lt;First&gt;</a>`,
				`<time datetime="2026-10-20">2026-10-20</time>`,
				`<time datetime="2026-10-21">2026-10-21</time>`, "<p>&lt;Excerpt&gt;</p>",
			} {
				if !strings.Contains(list, want) {
					t.Errorf("missing %q in %s", want, list)
				}
			}
			if strings.Index(list, "Journal intro") > strings.Index(list, `<ul class="entries">`) ||
				strings.Index(list, "&lt;First&gt;</a>") > strings.Index(list, ">Second</a>") {
				t.Fatalf("intro or chronological order changed: %s", list)
			}
			for _, forbidden := range []string{"Undated", "Not an entry", "Section landing", "Deep article", "Private",
				"Article body must not appear", "Second body", "<Excerpt>"} {
				if strings.Contains(list, forbidden) {
					t.Errorf("unexpected %q in list", forbidden)
				}
			}
			alphabetical := read("blog/alphabetical/index.html")
			if strings.Index(alphabetical, ">Undated</a>") > strings.Index(alphabetical, ">Second</a>") ||
				strings.Index(alphabetical, ">Second</a>") > strings.Index(alphabetical, "&lt;First&gt;</a>") {
				t.Fatalf("reverse title ordering failed: %s", alphabetical)
			}
			empty := read("blog/empty/list/index.html")
			if !strings.Contains(empty, "<p>Empty intro</p>") || strings.Contains(empty, "<ul>") {
				t.Fatalf("empty list intro changed: %s", empty)
			}
			var sitemap struct {
				XMLName xml.Name
				URLs    []struct {
					Location string `xml:"loc"`
					Modified string `xml:"lastmod"`
				} `xml:"url"`
			}
			if err := xml.Unmarshal([]byte(read("sitemap.xml")), &sitemap); err != nil {
				t.Fatal(err)
			}
			if len(sitemap.URLs) != result.Pages-1 ||
				sitemap.XMLName.Space != "http://www.sitemaps.org/schemas/sitemap/0.9" {
				t.Fatalf("sitemap = %+v", sitemap)
			}
			var locations []string
			for _, entry := range sitemap.URLs {
				locations = append(locations, entry.Location)
				if !strings.HasPrefix(entry.Location, "https://example.org:8443"+prefix) ||
					strings.Contains(entry.Location, "draft") {
					t.Fatalf("incorrect publication URL: %s", entry.Location)
				}
				if strings.HasSuffix(entry.Location, "/blog/first/") {
					if entry.Modified != "2026-10-21" {
						t.Fatalf("explicit update date missing: %+v", entry)
					}
				} else if entry.Modified != "" {
					t.Fatalf("update date inferred: %+v", entry)
				}
			}
			if !slices.IsSorted(locations) || !slices.Contains(locations, "https://example.org:8443"+prefix+"blog/undated/") {
				t.Fatalf("sitemap ordering or undated standalone page missing: %v", locations)
			}
			if !strings.Contains(read("robots.txt"), "Sitemap: https://example.org:8443"+prefix+"sitemap.xml\n") {
				t.Fatal("sitemap not advertised")
			}
			origin := "https://example.org:8443" + prefix
			if !strings.Contains(list, `<link rel="alternate" type="application/atom+xml" href="`+prefix+`blog/list/feed.xml" title="Journal">`) {
				t.Errorf("feed not advertised: %s", list)
			}
			var feed struct {
				XMLName  xml.Name
				Language string `xml:"lang,attr"`
				ID       string `xml:"id"`
				Updated  string `xml:"updated"`
				Author   string `xml:"author>name"`
				Links    []struct {
					Rel  string `xml:"rel,attr"`
					Href string `xml:"href,attr"`
				} `xml:"link"`
				Entries []struct {
					ID        string `xml:"id"`
					Title     string `xml:"title"`
					Published string `xml:"published"`
					Updated   string `xml:"updated"`
					Summary   string `xml:"summary"`
					Content   struct {
						Type string `xml:"type,attr"`
						Base string `xml:"base,attr"`
						Body string `xml:",chardata"`
					} `xml:"content"`
				} `xml:"entry"`
			}
			if err := xml.Unmarshal([]byte(read("blog/list/feed.xml")), &feed); err != nil {
				t.Fatal(err)
			}
			if feed.XMLName.Space != "http://www.w3.org/2005/Atom" || feed.Language != "en" || feed.ID != origin+"blog/list/" ||
				feed.Updated != "2026-10-21T00:00:00Z" || feed.Author != "example.org:8443" || len(feed.Links) != 2 ||
				feed.Links[0].Href != origin+"blog/list/feed.xml" || len(feed.Entries) != 2 {
				t.Fatalf("feed = %+v", feed)
			}
			entry := feed.Entries[0]
			if entry.ID != origin+"blog/first/" || entry.Title != "<First>" || entry.Published != "2026-10-20T00:00:00Z" ||
				entry.Updated != "2026-10-21T00:00:00Z" || entry.Summary != "<Excerpt>" || entry.Content.Type != "html" ||
				entry.Content.Base != origin+"blog/first/" || !strings.Contains(entry.Content.Body, "<p>Article body must not appear in a list</p>") ||
				feed.Entries[1].Title != "Second" {
				t.Fatalf("feed entries = %+v", feed.Entries)
			}
			if alphabetical := read("blog/alphabetical/feed.xml"); strings.Contains(alphabetical, "Undated") {
				t.Fatalf("undated page in feed: %s", alphabetical)
			}
			if _, err := os.Stat(filepath.Join(site, "public/blog/empty/list/feed.xml")); !os.IsNotExist(err) ||
				strings.Contains(empty, "atom+xml") {
				t.Fatalf("empty list has a feed: %v", err)
			}
			first := read("blog/first/index.html")
			if !strings.Contains(first, `src="`+prefix+`blog/first/seed.svg"`) ||
				!strings.Contains(first, `href="`+prefix+`contact/"`) {
				t.Fatalf("metadata URLs not prepared: %s", first)
			}
		})
	}
}

func TestBuildPublicationWithoutOrigin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := createSite(t, map[string]string{
		"cress.toml":       "[site]\npath = '/company/'",
		"content/index.md": "+++\ntitle = 'Home'\n+++\n[About](about.md)",
		"content/about.md": "+++\ntitle = 'About'\n+++",
	})
	result, err := generator.Build(context.Background(), site)
	if err != nil || len(result.Warnings) != 1 || !strings.Contains(result.Warnings[0], "site.host") {
		t.Fatalf("result = %+v, error = %v", result, err)
	}
	home, err := os.ReadFile(filepath.Join(site, "public", "index.html"))
	if err != nil || strings.Contains(string(home), `rel="canonical"`) ||
		!strings.Contains(string(home), `href="/company/about/"`) {
		t.Fatalf("missing origin broke browsing or invented canonical: %s, %v", home, err)
	}
	if _, err := os.Lstat(filepath.Join(site, "public", "sitemap.xml")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("sitemap generated without origin: %v", err)
	}
	robots, err := os.ReadFile(filepath.Join(site, "public", "robots.txt"))
	if err != nil || string(robots) != "User-agent: *\nAllow: /\n" {
		t.Fatalf("robots = %s, error = %v", robots, err)
	}
}

func TestBuildPublicationCollisions(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, generated := range []string{"robots.txt", "sitemap.xml"} {
		t.Run(generated, func(t *testing.T) {
			site := createSite(t, map[string]string{
				"cress.toml":          "[site]\nhost = 'https://example.org'",
				"content/index.md":    "+++\ntitle = 'Home'\n+++",
				"static/" + generated: "authored file",
				"public/keep.txt":     "previous output",
			})
			_, err := generator.Build(context.Background(), site)
			if err == nil || !strings.Contains(err.Error(), "collision") ||
				!strings.Contains(err.Error(), "static/"+generated) ||
				!strings.Contains(err.Error(), "generated "+generated) {
				t.Fatalf("collision missing competing sources: %v", err)
			}
			previous, err := os.ReadFile(filepath.Join(site, "public", "keep.txt"))
			if err != nil || string(previous) != "previous output" {
				t.Fatalf("previous output changed: %q, %v", previous, err)
			}
		})
	}
	site := createSite(t, map[string]string{
		"content/index.md":   "+++\ntitle = 'Home'\n+++",
		"static/sitemap.xml": "authored sitemap",
	})
	if _, err := generator.Build(context.Background(), site); err != nil {
		t.Fatalf("omitted sitemap still claims a destination: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(site, "public", "sitemap.xml"))
	if err != nil || string(data) != "authored sitemap" {
		t.Fatalf("authored sitemap changed: %s, %v", data, err)
	}
}
