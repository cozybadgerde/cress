package content_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/content"
	"github.com/cozybadgerde/cress/internal/page"
)

func TestMetadataResolution(t *testing.T) {
	config, _, err := content.ParseConfig([]byte(`
[site]
title = "Site identity"
description = "Shared description"
language = "de"
path = "/company/"
[site.image]
src = "/shared.jpg"
alt = "Shared alternative"
`))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, frontmatter, description string
		image                          *page.Image
		pageType                       page.Type
	}{
		{"inherit", "", "Shared description", &page.Image{URL: "/company/shared.jpg", Alt: "Shared alternative"}, page.Ordinary},
		{"override", "description = 'Page description'\n[image]\nsrc = '/page.jpg'\nalt = 'Page alternative'",
			"Page description", &page.Image{URL: "/company/page.jpg", Alt: "Page alternative"}, page.Ordinary},
		{"clear", "description = ''\n[image]\nsrc = ''", "", nil, page.Ordinary},
		{"landing", "type = 'landing'", "Shared description",
			&page.Image{URL: "/company/shared.jpg", Alt: "Shared alternative"}, page.Landing},
		{"external image", "[image]\nsrc = 'https://example.net/photo.jpg'\nalt = ''",
			"Shared description", &page.Image{URL: "https://example.net/photo.jpg", Alt: ""}, page.Ordinary},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := "+++\ntitle = 'Page title'\n" + test.frontmatter + "\n+++\nBody"
			prepared, _, err := content.Prepare("nested/page.md", []byte(source), config)
			if err != nil {
				t.Fatal(err)
			}
			if prepared.Data.Title != "Page title" || prepared.Data.Description != test.description ||
				prepared.Data.Type != test.pageType || !reflect.DeepEqual(prepared.Data.Image, test.image) {
				t.Fatalf("unexpected metadata: %+v", prepared.Data)
			}
			if prepared.Data.URL != "/company/nested/page/" ||
				prepared.Destination != "nested/page/index.html" || prepared.Data.Site.Language != "de" {
				t.Fatalf("unexpected site and route data: %+v", prepared)
			}
		})
	}
	if config.Image.URL != "/company/shared.jpg" || config.Image.Alt != "Shared alternative" {
		t.Fatal("page overrides changed shared defaults")
	}
}

func TestPageValidationAndTypes(t *testing.T) {
	tests := []struct {
		name, fields, wantError string
		draft                   bool
	}{
		{"published list", "title = 'List'\ntype = 'list'", "", false},
		{"excluded list", "draft = true\ntype = 'list'", "", true},
		{"invalid type", "title = 'Page'\ntype = 'article'", "type must be ordinary, landing, or list", false},
		{"invalid draft type", "draft = true\ntype = false", "type must be a string", true},
		{"invalid description", "title = 'Page'\ndescription = 1", "description must be a string", false},
		{"invalid draft description", "draft = true\ndescription = false", "description must be a string", true},
		{"incomplete override", "title = 'Page'\n[image]\nalt = 'Missing source'", "image.src is required", false},
		{"no inherited alternative", "title = 'Page'\n[image]\nsrc = '/page.jpg'", "image.alt is required", false},
		{"relative draft image", "draft = true\n[image]\nsrc = 'relative.jpg'\nalt = ''", "", true},
		{"invalid draft image", "draft = true\n[image]\nsrc = '//example.org/image.jpg'\nalt = ''", "image.src", true},
		{"missing page title", "description = 'Not a title'", "published pages require a title", false},
	}
	config := defaultConfig(t)
	config.Site.Title = "Not a page title fallback"
	config.Image = &page.Image{URL: "/shared.jpg", Alt: "Not a page alternative fallback"}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			prepared, _, err := content.Prepare("page.md", []byte("+++\n"+test.fields+"\n+++\nBody"), config)
			if test.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("error = %v, want %q", err, test.wantError)
				}
			} else if err != nil || prepared.Draft != test.draft {
				t.Fatalf("draft = %v, error = %v", prepared.Draft, err)
			}
		})
	}
}
