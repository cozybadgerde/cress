package content

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/page"
)

func TestDateMetadata(t *testing.T) {
	config, _, err := ParseConfig(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"2024-02-29", "'2024-02-29'", "'0001-01-01'", "'9999-12-31'"} {
		t.Run(value, func(t *testing.T) {
			item, warnings, err := Prepare("page.md", []byte("+++\ntitle = 'Page'\npublished_at = "+value+
				"\nupdated_at = "+value+"\nexcerpt = '<escaped>'\n+++\nBody"), config)
			if err != nil || len(warnings) != 0 {
				t.Fatalf("warnings = %v, error = %v", warnings, err)
			}
			if item.Data.PublishedAt != strings.Trim(value, "'") ||
				item.Data.UpdatedAt != item.Data.PublishedAt || item.Data.Excerpt != "<escaped>" {
				t.Fatalf("metadata = %+v", item.Data)
			}
		})
	}
	for _, value := range []string{
		"''", "'2023-02-29'", "'2026-04-31'", "'2026-13-01'", "'2026-01-00'",
		"'2026-1-01'", "'0000-01-01'", "'2026-10-20T12:00:00Z'", "2026-10-20T12:00:00Z",
		"2026-10-20T12:00:00", "12:00:00", "true", "20261020", "[]", "{}",
	} {
		for _, field := range []string{"published_at", "updated_at"} {
			t.Run(field+"/"+value, func(t *testing.T) {
				_, _, err := Prepare("draft.md", []byte("+++\ndraft = true\n"+field+" = "+value+"\n+++"), config)
				if err == nil || !strings.Contains(err.Error(), "content/draft.md: "+field) {
					t.Fatalf("invalid draft metadata accepted: %v", err)
				}
			})
		}
	}
}

func TestListFields(t *testing.T) {
	config, _, err := ParseConfig(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, fields := range []string{
		"type = 'ordinary'\n[list]",
		"type = 'landing'\n[list]",
		"type = 'list'\nlist = false",
		"type = 'list'\n[list]\norder = 'updated_at'",
		"type = 'list'\n[list]\norder = false",
		"type = 'list'\n[list]\nreverse = 'true'",
		"excerpt = []",
	} {
		t.Run(fields, func(t *testing.T) {
			if _, _, err := Prepare("draft.md", []byte("+++\ndraft = true\n"+fields+"\n+++"), config); err == nil {
				t.Fatal("invalid list metadata accepted on draft")
			}
		})
	}
	item, warnings, err := Prepare("list.md", []byte("+++\ntitle = 'List'\ntype = 'list'\n[list]\nextra = true\n+++\nIntro"), config)
	if err != nil || !slices.Equal(warnings, []string{`content/list.md: unknown frontmatter field "list.extra"`}) ||
		item.list.order != "published_at" || item.list.reverse {
		t.Fatalf("item = %+v, warnings = %v, error = %v", item, warnings, err)
	}
}

func TestPreparedLists(t *testing.T) {
	makePage := func(source, fields string) Page {
		t.Helper()
		config, _, err := ParseConfig([]byte("[site]\npath = '/company/'"))
		if err != nil {
			t.Fatal(err)
		}
		item, _, err := prepareMetadata(source, []byte("+++\n"+fields+"\n+++\nNo automatic excerpt"), config)
		if err != nil {
			t.Fatal(err)
		}
		return item
	}
	for _, source := range []string{"blog/list.md", "blog/list/index.md"} {
		for _, order := range []string{"published_at", "title"} {
			for _, reverse := range []bool{false, true} {
				t.Run(source+"/"+order+map[bool]string{true: "/reverse", false: "/forward"}[reverse], func(t *testing.T) {
					list := makePage(source, "title = 'List'\ntype = 'list'")
					list.list = listOptions{order: order, reverse: reverse}
					pages := []Page{
						list,
						makePage("blog/index.md", "title = 'Section'\npublished_at = 2026-10-20"),
						makePage("blog/a.md", "title = 'zebra'\npublished_at = 2026-10-20\nupdated_at = 2020-01-01\nexcerpt = '<Excerpt>'"),
						makePage("blog/b/index.md", "title = 'ALPHA'\npublished_at = 2026-10-20"),
						makePage("blog/c.md", "title = 'alpha'\npublished_at = 2026-09-20\nupdated_at = 2027-01-01"),
						makePage("blog/undated.md", "title = 'Undated'"),
						makePage("blog/landing.md", "title = 'Landing'\ntype = 'landing'"),
						makePage("blog/other-list.md", "title = 'Other'\ntype = 'list'\n[list]\norder = 'title'"),
						makePage("blog/draft.md", "title = 'Draft'\ndraft = true"),
						makePage("blog/archive/old.md", "title = 'Deep'"),
						makePage("outside.md", "title = 'Outside'"),
					}
					warnings := prepareLists(pages, Config{})
					want := []string{"/company/blog/a/", "/company/blog/b/", "/company/blog/c/"}
					if order == "title" {
						want = []string{"/company/blog/b/", "/company/blog/c/", "/company/blog/undated/", "/company/blog/a/"}
						if len(warnings) != 0 {
							t.Fatalf("title mode warned: %v", warnings)
						}
					} else if len(warnings) != 1 || !strings.Contains(warnings[0], "content/blog/undated.md") ||
						!strings.Contains(warnings[0], list.Source) {
						t.Fatalf("missing-date warning = %v", warnings)
					}
					if reverse {
						slices.Reverse(want)
					}
					var actual []string
					for _, entry := range pages[0].Data.Entries {
						actual = append(actual, entry.URL)
						if entry.URL == "/company/blog/a/" && (entry.Excerpt != "<Excerpt>" || entry.UpdatedAt != "2020-01-01") {
							t.Fatalf("entry metadata lost: %+v", entry)
						}
						if entry.URL != "/company/blog/a/" && entry.Excerpt != "" {
							t.Fatalf("excerpt inferred: %+v", entry)
						}
					}
					if !slices.Equal(actual, want) {
						t.Fatalf("entries = %v, want %v", actual, want)
					}
				})
			}
		}
	}
	for _, source := range []string{"blog.md", "blog/index.md", "index.md"} {
		pages := []Page{
			makePage(source, "title = 'List'\ntype = 'list'\n[list]\norder = 'title'"),
			makePage("about/index.md", "title = 'About'"),
			makePage("nested/child.md", "title = 'Deep'"),
		}
		prepareLists(pages, Config{})
		if len(pages[0].Data.Entries) != 1 || pages[0].Data.Entries[0].URL != "/company/about/" {
			t.Fatalf("%s selected incorrect logical scope: %+v", source, pages[0].Data.Entries)
		}
	}
}

func TestTitleOrderingAndEmptyLists(t *testing.T) {
	pages := []Page{
		{Destination: "list/index.html", Data: page.Page{Type: page.List}, list: listOptions{order: "title"}},
		{Destination: "a/index.html", Data: page.Page{
			Type: page.Ordinary, Title: "É", URL: "/a/", Description: "Description",
			Image: &page.Image{URL: "/image.svg", Alt: "Alternative", Caption: []page.CaptionPart{{Text: "Credit", URL: "/credit/"}}},
		}},
		{Destination: "b/index.html", Data: page.Page{Type: page.Ordinary, Title: "e\u0301", URL: "/b/"}},
		{Destination: "c/index.html", Data: page.Page{Type: page.Ordinary, Title: "Straße", URL: "/c/"}},
		{Destination: "d/index.html", Data: page.Page{Type: page.Ordinary, Title: "STRASSE", URL: "/d/"}},
	}
	prepareLists(pages, Config{})
	var urls []string
	for _, entry := range pages[0].Data.Entries {
		urls = append(urls, entry.URL)
	}
	if !reflect.DeepEqual(urls, []string{"/c/", "/d/", "/a/", "/b/"}) {
		t.Fatalf("Unicode keys or route ties incorrect: %v", urls)
	}
	entry := pages[0].Data.Entries[2]
	if entry.Description != "Description" || !reflect.DeepEqual(entry.Image, pages[1].Data.Image) {
		t.Fatalf("prepared entry metadata changed: %+v", entry)
	}
	empty := []Page{{Data: page.Page{Type: page.List}, Destination: "index.html"}}
	if warnings := prepareLists(empty, Config{}); len(warnings) != 0 || len(empty[0].Data.Entries) != 0 {
		t.Fatalf("empty list failed: %+v, %v", empty, warnings)
	}
}
