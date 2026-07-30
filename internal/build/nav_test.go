package build

import (
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/config"
	"github.com/cozybadgerde/cress/internal/content"
	"github.com/cozybadgerde/cress/internal/theme"
)

func TestResolveNav(t *testing.T) {
	pages := []*content.Page{
		{SourcePath: "index.md", URL: "/", Title: "Home"},
		{SourcePath: "about.md", URL: "/about.html", Title: "About page"},
		{SourcePath: "imprint.md", URL: "/imprint.html", Title: "Imprint"},
	}

	nav := config.Nav{
		Main: []config.NavItem{
			{Title: "Home", Path: "index.md"},
			{Title: "", Path: "about.md"},       // empty title falls back to page title
			{Title: "Gone", Path: "missing.md"}, // dropped with a warning
		},
		Footer: []config.NavItem{
			{Title: "Imprint", Path: "imprint.md"},
			{Title: "Privacy", Path: "privacy.md"}, // dropped with a warning
		},
	}

	got, warnings := resolveNav(nav, pages, "")

	if len(got.Main) != 2 {
		t.Fatalf("got %d main links, want 2", len(got.Main))
	}
	if got.Main[0] != (theme.NavLink{Title: "Home", URL: "/"}) {
		t.Errorf("main[0] = %+v", got.Main[0])
	}
	if got.Main[1] != (theme.NavLink{Title: "About page", URL: "/about.html"}) {
		t.Errorf("main[1] = %+v, want title from page", got.Main[1])
	}
	if len(got.Footer) != 1 {
		t.Fatalf("got %d footer links, want 1", len(got.Footer))
	}
	if got.Footer[0] != (theme.NavLink{Title: "Imprint", URL: "/imprint.html"}) {
		t.Errorf("footer[0] = %+v", got.Footer[0])
	}

	// Each warning names its group, so the user knows which table to fix.
	if len(warnings) != 2 {
		t.Fatalf("got %d warnings, want 2: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0], "nav.main") {
		t.Errorf("warning[0] = %q, want it to name nav.main", warnings[0])
	}
	if !strings.Contains(warnings[1], "nav.footer") {
		t.Errorf("warning[1] = %q, want it to name nav.footer", warnings[1])
	}
}

func TestActiveNav(t *testing.T) {
	nav := theme.NavView{
		Main:   []theme.NavLink{{Title: "Home", URL: "/"}, {Title: "About", URL: "/about.html"}},
		Footer: []theme.NavLink{{Title: "Imprint", URL: "/imprint.html"}},
	}

	got := activeNav(nav, "/about.html")

	if got.Main[0].Active {
		t.Error("home should not be active")
	}
	if !got.Main[1].Active {
		t.Error("about should be active")
	}
	if got.Footer[0].Active {
		t.Error("imprint should not be active")
	}
	// The source slices must not be mutated.
	if nav.Main[1].Active {
		t.Error("activeNav mutated its input")
	}
}

func TestActiveNavFlagsFooter(t *testing.T) {
	// A footer link to the current page is current too.
	nav := theme.NavView{Footer: []theme.NavLink{{Title: "Imprint", URL: "/imprint.html"}}}

	got := activeNav(nav, "/imprint.html")

	if !got.Footer[0].Active {
		t.Error("imprint should be active on its own page")
	}
}
