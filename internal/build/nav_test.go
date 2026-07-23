package build

import (
	"testing"

	"github.com/cozybadgerde/cress/internal/config"
	"github.com/cozybadgerde/cress/internal/content"
)

func TestResolveNav(t *testing.T) {
	pages := []*content.Page{
		{SourcePath: "index.md", URL: "/", Title: "Home"},
		{SourcePath: "about.md", URL: "/about.html", Title: "About page"},
	}

	items := []config.NavItem{
		{Title: "Home", Path: "index.md"},
		{Title: "", Path: "about.md"},       // empty title falls back to page title
		{Title: "Gone", Path: "missing.md"}, // dropped with a warning
	}

	links, warnings := resolveNav(items, pages)

	if len(links) != 2 {
		t.Fatalf("got %d links, want 2", len(links))
	}
	if links[0] != (navLink{Title: "Home", URL: "/"}) {
		t.Errorf("link[0] = %+v", links[0])
	}
	if links[1] != (navLink{Title: "About page", URL: "/about.html"}) {
		t.Errorf("link[1] = %+v, want title from page", links[1])
	}
	if len(warnings) != 1 {
		t.Fatalf("got %d warnings, want 1: %v", len(warnings), warnings)
	}
}

func TestActiveNav(t *testing.T) {
	nav := []navLink{{Title: "Home", URL: "/"}, {Title: "About", URL: "/about.html"}}

	got := activeNav(nav, "/about.html")

	if got[0].Active {
		t.Error("home should not be active")
	}
	if !got[1].Active {
		t.Error("about should be active")
	}
	// The source slice must not be mutated.
	if nav[1].Active {
		t.Error("activeNav mutated its input")
	}
}
