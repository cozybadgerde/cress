package theme

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want semver
		ok   bool
	}{
		{name: "major minor", in: "1.1", want: semver{1, 1, 0}, ok: true},
		{name: "major minor patch", in: "1.0.1", want: semver{1, 0, 1}, ok: true},
		{name: "major only", in: "2", want: semver{2, 0, 0}, ok: true},
		{name: "v prefix", in: "v1.1", want: semver{1, 1, 0}, ok: true},
		{name: "surrounding space", in: "  1.2.3  ", want: semver{1, 2, 3}, ok: true},
		{name: "zero", in: "0.0.0", want: semver{0, 0, 0}, ok: true},
		{name: "prerelease is ignored", in: "1.0.0-alpha.3", want: semver{1, 0, 0}, ok: true},
		{name: "build metadata is ignored", in: "1.2.0+20260809", want: semver{1, 2, 0}, ok: true},
		{name: "leading zeroes", in: "01.02", want: semver{1, 2, 0}, ok: true},
		{name: "big numbers", in: "10.20.30", want: semver{10, 20, 30}, ok: true},

		{name: "untagged local build", in: "dev"},
		{name: "empty", in: ""},
		{name: "blank", in: "   "},
		{name: "v alone", in: "v"},
		{name: "words", in: "next"},
		{name: "four components", in: "1.2.3.4"},
		{name: "trailing dot", in: "1.2."},
		{name: "leading dot", in: ".1"},
		{name: "empty component", in: "1..2"},
		{name: "negative", in: "-1"},
		{name: "not a number", in: "1.x"},
		{name: "inner space", in: "1. 2"},
		{name: "a lizard", in: "🦎"},
		{name: "a toilet", in: "1.0.0; DROP TABLE themes"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseVersion(tc.in)
			if ok != tc.ok {
				t.Fatalf("parseVersion(%q) ok = %v, want %v", tc.in, ok, tc.ok)
			}
			if ok && got != tc.want {
				t.Errorf("parseVersion(%q) = %+v, want %+v", tc.in, got, tc.want)
			}
		})
	}
}

func TestSemverLess(t *testing.T) {
	tests := []struct {
		name string
		a, b semver
		want bool
	}{
		{name: "equal", a: semver{1, 1, 0}, b: semver{1, 1, 0}},
		{name: "lower major", a: semver{1, 9, 9}, b: semver{2, 0, 0}, want: true},
		{name: "higher major", a: semver{2, 0, 0}, b: semver{1, 9, 9}},
		{name: "lower minor", a: semver{1, 0, 9}, b: semver{1, 1, 0}, want: true},
		{name: "higher minor", a: semver{1, 1, 0}, b: semver{1, 0, 9}},
		{name: "lower patch", a: semver{1, 1, 0}, b: semver{1, 1, 1}, want: true},
		{name: "higher patch", a: semver{1, 1, 1}, b: semver{1, 1, 0}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.a.less(tc.b); got != tc.want {
				t.Errorf("%+v.less(%+v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestLoadMetaFields(t *testing.T) {
	fsys := fstest.MapFS{MetaFile: {Data: []byte(`
name = "Birch"
description = "Clean black on white."
author = "Cozy Badger"
license = "MIT"
homepage = "https://example.com/birch"
cress = "1.1"
`)}}

	meta, err := loadMeta(fsys)
	if err != nil {
		t.Fatalf("loadMeta: %v", err)
	}
	if meta == nil {
		t.Fatal("loadMeta = nil, want the parsed metadata")
	}

	for _, field := range []struct{ name, got, want string }{
		{"Name", meta.Name, "Birch"},
		{"Description", meta.Description, "Clean black on white."},
		{"Author", meta.Author, "Cozy Badger"},
		{"License", meta.License, "MIT"},
		{"Homepage", meta.Homepage, "https://example.com/birch"},
		{"Cress", meta.Cress, "1.1"},
	} {
		if field.got != field.want {
			t.Errorf("%s = %q, want %q", field.name, field.got, field.want)
		}
	}
	if meta.declared != (semver{1, 1, 0}) || !meta.declaredOK {
		t.Errorf("declared = %+v (ok %v), want {1 1 0}", meta.declared, meta.declaredOK)
	}
	if warnings := meta.Warnings("1.1.0"); len(warnings) != 0 {
		t.Errorf("Warnings = %v, want none", warnings)
	}
}

func TestLoadMetaAbsentIsNotAnError(t *testing.T) {
	meta, err := loadMeta(fstest.MapFS{"templates/page.html": {Data: []byte(`<main></main>`)}})
	if err != nil {
		t.Fatalf("loadMeta: %v", err)
	}
	if meta != nil {
		t.Errorf("loadMeta = %+v, want nil for a theme with no %s", meta, MetaFile)
	}
	// A theme with no metadata has nothing to warn about, and asking must not
	// depend on the caller checking for nil first.
	if warnings := meta.Warnings("1.1.0"); warnings != nil {
		t.Errorf("Warnings = %v, want none", warnings)
	}
}

func TestLoadMetaMalformedIsAnError(t *testing.T) {
	_, err := loadMeta(fstest.MapFS{MetaFile: {Data: []byte("name = \nthis is not toml")}})
	if err == nil {
		t.Fatal("loadMeta = nil error, want a parse failure")
	}
	if !strings.Contains(err.Error(), MetaFile) {
		t.Errorf("error %q does not name %s", err, MetaFile)
	}
}

func TestLoadMetaUnknownKeyWarnsRatherThanFailing(t *testing.T) {
	meta, err := loadMeta(fstest.MapFS{MetaFile: {Data: []byte(`
name = "Birch"
screenshot = "preview.png"
`)}})
	if err != nil {
		t.Fatalf("loadMeta: %v", err)
	}
	if meta.Name != "Birch" {
		t.Errorf("Name = %q, want the file to load anyway", meta.Name)
	}

	warnings := meta.Warnings("1.1.0")
	if len(warnings) != 1 {
		t.Fatalf("Warnings = %v, want one", warnings)
	}
	if !strings.Contains(warnings[0], "screenshot") {
		t.Errorf("warning %q does not name the unknown key", warnings[0])
	}
}

// warningsFor loads a theme declaring the given contract and asks what it has
// to say to the cress named by running.
func warningsFor(t *testing.T, declared, running string) []string {
	t.Helper()
	meta, err := loadMeta(fstest.MapFS{MetaFile: {Data: []byte("cress = \"" + declared + "\"\n")}})
	if err != nil {
		t.Fatalf("loadMeta: %v", err)
	}
	return meta.Warnings(running)
}

func TestMetaWarningsStaysSilent(t *testing.T) {
	tests := []struct{ name, declared, running string }{
		{name: "same version", declared: "1.1", running: "1.1.0"},
		{name: "newer patch", declared: "1.1", running: "1.1.4"},
		{name: "newer minor", declared: "1.1", running: "1.9.0"},
		{name: "prerelease of the declared version", declared: "1.1", running: "1.1.0-alpha.1"},
		{name: "no declaration", declared: "", running: "1.1.0"},
		{name: "untagged local build", declared: "1.1", running: "dev"},
		{name: "no running version", declared: "1.1", running: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if warnings := warningsFor(t, tc.declared, tc.running); len(warnings) != 0 {
				t.Errorf("Warnings = %v, want none", warnings)
			}
		})
	}
}

func TestMetaWarningsReports(t *testing.T) {
	tests := []struct{ name, declared, running, want string }{
		{
			name: "older minor", declared: "1.1", running: "1.0.1",
			want: "may read fields this version does not provide",
		},
		{
			name: "older major", declared: "2.0", running: "1.9.9",
			want: "may read fields this version does not provide",
		},
		{
			name: "newer major", declared: "1.1", running: "2.0.0",
			want: "not promised across a major version",
		},
		{
			name: "unparseable declaration", declared: "next", running: "1.1.0",
			want: "is not a version",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			warnings := warningsFor(t, tc.declared, tc.running)
			if len(warnings) != 1 {
				t.Fatalf("Warnings = %v, want one", warnings)
			}
			if !strings.Contains(warnings[0], tc.want) {
				t.Errorf("warning %q does not mention %q", warnings[0], tc.want)
			}
		})
	}
}

// An unparseable declaration is reported once, when the file is read, and does
// not then also produce a comparison against a version that could not be read.
func TestMetaWarningsUnparseableDeclarationIsReportedOnce(t *testing.T) {
	meta, err := loadMeta(fstest.MapFS{MetaFile: {Data: []byte(`cress = "next"`)}})
	if err != nil {
		t.Fatalf("loadMeta: %v", err)
	}
	for _, running := range []string{"dev", "1.0.0", "9.9.9"} {
		if warnings := meta.Warnings(running); len(warnings) != 1 {
			t.Errorf("Warnings(%q) = %v, want one", running, warnings)
		}
	}
}

// Warnings returns a fresh slice each time, so a caller appending to one result
// cannot reach the warnings the next caller gets.
func TestMetaWarningsDoesNotShareItsSlice(t *testing.T) {
	meta, err := loadMeta(fstest.MapFS{MetaFile: {Data: []byte("cress = \"1.1\"\nscreenshot = \"x.png\"\n")}})
	if err != nil {
		t.Fatalf("loadMeta: %v", err)
	}

	first := meta.Warnings("1.0.0")
	first[0] = "clobbered"
	if second := meta.Warnings("1.0.0"); second[0] == "clobbered" {
		t.Error("Warnings returned a slice aliasing the stored warnings")
	}
}
