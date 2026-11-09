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
					_, warnings := prepareLists(pages, Config{})
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
	if _, warnings := prepareLists(empty, Config{}); len(warnings) != 0 || len(empty[0].Data.Entries) != 0 {
		t.Fatalf("empty list failed: %+v, %v", empty, warnings)
	}
}

func TestPagination(t *testing.T) {
	makePages := func(section string, perPage, count int) []Page {
		pages := []Page{{Source: "list.md", Destination: section + "index.html",
			Data: page.Page{Type: page.List, URL: "/" + section}, list: listOptions{order: "title", perPage: perPage}}}
		for index := range count {
			name := string(rune('a' + index))
			pages = append(pages, Page{Destination: name + "/index.html",
				Data: page.Page{Type: page.Ordinary, Title: name, URL: "/" + name + "/"}})
		}
		return pages
	}
	pages := makePages("blog/", 2, 5)
	all, _ := prepareLists(pages, Config{Host: "https://example.com"})
	all = slices.DeleteFunc(all, func(item Page) bool { return item.Source != "list.md" })
	if len(all) != 3 {
		t.Fatalf("pages = %d, want 3", len(all))
	}
	for index, want := range []struct {
		destination, url, canonical, previous, next string
		entries                                     []string
	}{
		{"blog/index.html", "/blog/", "", "", "/blog/2/", []string{"/a/", "/b/"}},
		{"blog/2/index.html", "/blog/2/", "https://example.com/blog/2/", "/blog/", "/blog/3/", []string{"/c/", "/d/"}},
		{"blog/3/index.html", "/blog/3/", "https://example.com/blog/3/", "/blog/2/", "", []string{"/e/"}},
	} {
		item := all[index]
		var entries []string
		for _, entry := range item.Data.Entries {
			entries = append(entries, entry.URL)
		}
		pagination := item.Data.Pagination
		if item.Destination != want.destination || item.Data.URL != want.url || item.Data.Canonical != want.canonical ||
			item.Source != "list.md" || !slices.Equal(entries, want.entries) || pagination.Number != index+1 ||
			pagination.Total != 3 || pagination.Previous != want.previous || pagination.Next != want.next {
			t.Fatalf("page %d = %+v, pagination = %+v", index+1, item, pagination)
		}
		for linkIndex, link := range pagination.Pages {
			if link.Current != (linkIndex == index) {
				t.Fatalf("page %d links = %+v", index+1, pagination.Pages)
			}
		}
	}
	home := makePages("", 1, 2)
	if all, _ := prepareLists(home, Config{}); len(all) != 4 || all[3].Destination != "2/index.html" {
		t.Fatalf("home pagination = %+v", all)
	}
	for _, perPage := range []int{0, 5} {
		pages := makePages("blog/", perPage, 5)
		if all, _ := prepareLists(pages, Config{}); len(all) != 6 || all[0].Data.Pagination != nil ||
			len(all[0].Data.Entries) != 5 {
			t.Fatalf("per_page %d split the list: %+v", perPage, all)
		}
	}
}

func TestListPerPage(t *testing.T) {
	config, _, err := ParseConfig([]byte("[list]\nper_page = 3"))
	if err != nil || config.ListPerPage != 3 {
		t.Fatalf("config = %+v, error = %v", config, err)
	}
	for fields, want := range map[string]int{"": 3, "[list]\nper_page = 0": 0, "[list]\nper_page = 7": 7} {
		item, _, err := Prepare("list.md", []byte("+++\ntitle = 'List'\ntype = 'list'\n"+fields+"\n+++"), config)
		if err != nil || item.list.perPage != want {
			t.Fatalf("%q: per_page = %d, error = %v", fields, item.list.perPage, err)
		}
	}
	for _, value := range []string{"-1", "'10'", "1.5", "2147483648"} {
		if _, _, err := ParseConfig([]byte("[list]\nper_page = " + value)); err == nil {
			t.Fatalf("site per_page %s accepted", value)
		}
		if _, _, err := Prepare("list.md", []byte("+++\ntype = 'list'\n[list]\nper_page = "+value+"\n+++"), config); err == nil {
			t.Fatalf("page per_page %s accepted", value)
		}
	}
}
