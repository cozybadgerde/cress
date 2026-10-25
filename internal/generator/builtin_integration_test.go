package generator_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/generator"
	"github.com/cozybadgerde/cress/internal/scaffold"
	"github.com/cozybadgerde/cress/internal/theme"
)

func TestBuiltinCompletePageStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	for _, name := range []string{"poppy", "clover"} {
		site := t.TempDir()
		if err := scaffold.Init(context.Background(), site, name); err != nil {
			t.Fatal(err)
		}
		if _, err := generator.Build(context.Background(), site); err != nil {
			t.Fatal(err)
		}
		for _, file := range []string{"index.html", "services/index.html", "about/index.html", "contact/index.html", "journal/list/index.html", "journal/welcome/index.html"} {
			data, err := os.ReadFile(filepath.Join(site, "public", file))
			if err != nil {
				t.Fatal(err)
			}
			html := string(data)
			for _, invariant := range []string{`<html lang="en">`, `href="#main"`, `id="main" tabindex="-1"`, `aria-label="Main"`, `aria-label="Footer"`} {
				if file == "index.html" && invariant == `id="main" tabindex="-1"` {
					invariant = `id="main" class="landing" tabindex="-1"`
				}
				if !strings.Contains(html, invariant) {
					t.Errorf("%s/%s missing %s", name, file, invariant)
				}
			}
			if strings.Count(html, "<h1") != 1 || strings.Count(html, "<main ") != 1 ||
				strings.Contains(html, "<script") {
				t.Errorf("%s/%s invalid landmarks/headings or unexpected scripts", name, file)
			}
		}
		if err := theme.Init(context.Background(), site, name); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(site, "themes", name, "templates/page.gohtml")
		if err := os.WriteFile(target, []byte("Local package: {{.Title}}"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := generator.Build(context.Background(), site); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(site, "public/about/index.html"))
		if err != nil || string(data) != "Local package: About" {
			t.Fatalf("local package did not override built-in: %s, %v", data, err)
		}
		if err := os.WriteFile(target, []byte("{{missing}}"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := generator.Build(context.Background(), site); err == nil {
			t.Fatal("invalid local package fell back")
		}
	}
}
