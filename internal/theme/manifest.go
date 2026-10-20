package theme

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"golang.org/x/mod/semver"
)

const developmentVersion = "1.0.0"

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// Manifest identifies a theme and its supported runtime.
type Manifest struct {
	Name    string        `json:"name"`
	Version string        `json:"version"`
	Cress   Compatibility `json:"cress"`
}

// Compatibility declares a minimum runtime and exactly one supported major.
type Compatibility struct {
	Minimum string `json:"minimum"`
	Major   int    `json:"major"`
}

// ValidName reports whether a name is a portable single theme directory.
func ValidName(name string) bool { return namePattern.MatchString(name) }

func validVersion(version string) bool {
	return strings.Count(strings.SplitN(strings.SplitN(version, "-", 2)[0], "+", 2)[0], ".") == 2 &&
		!strings.HasPrefix(version, "v") && semver.IsValid("v"+version)
}

// ParseManifest validates known fields and warns about unknown ones.
func ParseManifest(source []byte) (Manifest, []string, error) {
	var manifest Manifest
	var warnings []string
	if !utf8.Valid(source) {
		return manifest, nil, fmt.Errorf("theme.json must be UTF-8")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(source, &fields); err != nil || fields == nil {
		return manifest, nil, fmt.Errorf("theme.json must be a JSON object")
	}
	for _, key := range slices.Sorted(maps.Keys(fields)) {
		if !slices.Contains([]string{"name", "version", "cress"}, key) {
			warnings = append(warnings, "theme.json: unknown field "+key)
		}
	}
	for _, key := range []string{"name", "version", "cress"} {
		if len(fields[key]) == 0 || bytes.Equal(fields[key], []byte("null")) {
			return manifest, warnings, fmt.Errorf("theme.json: %s is required and must not be null", key)
		}
	}
	var runtime map[string]json.RawMessage
	if err := json.Unmarshal(fields["cress"], &runtime); err != nil || runtime == nil {
		return manifest, warnings, fmt.Errorf("theme.json: cress must be an object")
	}
	for _, key := range slices.Sorted(maps.Keys(runtime)) {
		if key != "minimum" && key != "major" {
			warnings = append(warnings, "theme.json: unknown field cress."+key)
		}
	}
	for _, key := range []string{"minimum", "major"} {
		if len(runtime[key]) == 0 || bytes.Equal(runtime[key], []byte("null")) {
			return manifest, warnings, fmt.Errorf("theme.json: cress.%s is required and must not be null", key)
		}
	}
	for _, field := range []struct {
		name   string
		source json.RawMessage
		target any
	}{
		{"name", fields["name"], &manifest.Name},
		{"version", fields["version"], &manifest.Version},
		{"cress.minimum", runtime["minimum"], &manifest.Cress.Minimum},
		{"cress.major", runtime["major"], &manifest.Cress.Major},
	} {
		if err := json.Unmarshal(field.source, field.target); err != nil {
			return manifest, warnings, fmt.Errorf("theme.json: invalid %s: %w", field.name, err)
		}
	}
	if !ValidName(manifest.Name) || !validVersion(manifest.Version) || !validVersion(manifest.Cress.Minimum) ||
		manifest.Cress.Major < 0 {
		return manifest, warnings, fmt.Errorf("theme.json: invalid name, semantic version, or cress.major")
	}
	if semver.Major("v"+manifest.Cress.Minimum) != fmt.Sprintf("v%d", manifest.Cress.Major) {
		return manifest, warnings, fmt.Errorf("theme.json: cress.minimum must belong to cress.major")
	}
	return manifest, warnings, nil
}

// CheckCompatibility validates a runtime version without bypassing development checks.
func CheckCompatibility(manifest Manifest, version string) ([]string, error) {
	var warnings []string
	if version == "dev" {
		version = developmentVersion
		warnings = append(warnings, "development runtime: checking named theme against compatibility baseline "+developmentVersion)
	}
	version = strings.TrimPrefix(version, "v")
	if !validVersion(version) {
		return warnings, fmt.Errorf("invalid Cress runtime version %q", version)
	}
	if semver.Major("v"+version) != fmt.Sprintf("v%d", manifest.Cress.Major) ||
		semver.Compare("v"+version, "v"+manifest.Cress.Minimum) < 0 {
		return warnings, fmt.Errorf("theme %s requires Cress major %d, minimum %s; runtime is %s",
			manifest.Name, manifest.Cress.Major, manifest.Cress.Minimum, version)
	}
	return warnings, nil
}
