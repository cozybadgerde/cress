package generator_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cozybadgerde/cress/internal/generator"
	"github.com/cozybadgerde/cress/internal/output"
)

func TestPreviewDraftIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := createSite(t, map[string]string{
		"cress.toml":                       "[site]\nhost = 'https://example.org'\npath = '/company/'",
		"content/index.md":                 "+++\ntitle = 'Home'\n+++\nWelcome",
		"content/list.md":                  "+++\ntitle = 'List'\ntype = 'list'\n[list]\norder = 'title'\n+++",
		"content/private.md":               "+++\ntitle = 'Private'\ndraft = true\n+++\n![Private](private.assets/image.svg)\n[Home](index.md)",
		"content/private.assets/image.svg": "private image bytes",
		"public/keep":                      "unrecognized deployable output",
	})
	for _, drafts := range []bool{false, true} {
		session, _, err := output.NewSession(context.Background(), site)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := session.Close(); err != nil {
				t.Error(err)
			}
		})
		result, err := generator.BuildPreview(context.Background(), generator.PreviewRequest{
			Version: "dev", Drafts: drafts, Session: session,
		})
		if err != nil || result.Token == "" || result.Prefix != "/company/" {
			t.Fatalf("preview = %+v, %v", result, err)
		}
		for _, name := range []string{"private/index.html", "private/image.svg"} {
			_, err := os.Stat(filepath.Join(session.Directory(), "public", name))
			if (err == nil) != drafts {
				t.Fatalf("draft file %s visibility %v: %v", name, drafts, err)
			}
		}
		for _, name := range []string{"list/index.html", "sitemap.xml"} {
			data, err := os.ReadFile(filepath.Join(session.Directory(), "public", name))
			if err != nil || strings.Contains(string(data), "/company/private/") != drafts {
				t.Fatalf("draft reference in %s: %s, %v", name, data, err)
			}
		}
		data, err := os.ReadFile(filepath.Join(site, "public/keep"))
		if err != nil || string(data) != "unrecognized deployable output" {
			t.Fatalf("preview touched deployable output: %s, %v", data, err)
		}
	}
}

func TestIncludedDraftRequiresTitle(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping filesystem integration test")
	}
	site := createSite(t, map[string]string{"content/private.md": "+++\ndraft = true\n+++\nPrivate"})
	if _, err := generator.Build(context.Background(), site); err != nil {
		t.Fatal(err)
	}
	session, _, err := output.NewSession(context.Background(), site)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := session.Close(); err != nil {
			t.Error(err)
		}
	}()
	result, err := generator.BuildPreview(context.Background(), generator.PreviewRequest{
		Version: "dev", Drafts: true, Session: session,
	})
	if err == nil || !strings.Contains(err.Error(), "title") || result.Token != "" {
		t.Fatalf("untitled included draft accepted: %+v, %v", result, err)
	}
}
