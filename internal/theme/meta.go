package theme

import (
	"fmt"
	"io/fs"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// MetaFile is the optional metadata file at the root of a theme directory,
// beside templates/ and static/.
const MetaFile = "theme.toml"

// metaVersionKey names the contract field in error messages, so a theme author
// reads back the key they wrote rather than a Go field name.
const metaVersionKey = "cress"

// Meta is a theme's self-description, read from MetaFile. Every field is
// optional, and so is the file: a theme without one loads exactly as it did
// before the file existed.
//
// Only Cress is load-bearing. The rest is provenance, for a person reading the
// directory and for `cress theme validate` to report.
type Meta struct {
	// Name is the theme's display name. It is deliberately not checked against
	// the directory, because the directory is what cress.toml resolves: the two
	// cannot disagree in any way that changes what gets loaded.
	Name string `toml:"name"`
	// Description is a one-line summary of what the theme is for.
	Description string `toml:"description"`
	// Author names whoever wrote the theme.
	Author string `toml:"author"`
	// License is the license the theme is published under, ideally an SPDX
	// identifier ("MIT").
	License string `toml:"license"`
	// Homepage is where the theme lives.
	Homepage string `toml:"homepage"`
	// Cress is the template contract the theme was written against, as a version
	// ("1.1"). It reads as a statement rather than a range, which is what lets
	// one field catch a mismatch in either direction: a cress older than this
	// may not offer every field the theme reads, and a cress of a later major
	// version no longer promises the contract the theme was built on.
	Cress string `toml:"cress"`

	// declared is Cress parsed, and declaredOK whether it parsed at all. Holding
	// the result keeps the parse at load time, where the file being read is what
	// an error can name.
	declared   semver
	declaredOK bool
	// warnings is what reading the file found worth reporting: everything here
	// is independent of which cress is running, so it is settled once.
	warnings []string
}

// loadMeta reads MetaFile from the root of fsys. A theme without one yields a
// nil Meta and no error, which is the ordinary case rather than a failure.
//
// A file that does not parse is an error, because the theme author wrote
// something that is not TOML and nothing can be recovered from it. An unknown
// key is only a warning, and the difference is deliberate: an unknown key is
// also what a theme looks like when it declares something a later cress added,
// and refusing to load a theme over optional metadata would turn this file's own
// forward compatibility into a broken site. Nothing in here decides how a page
// renders, so nothing in here should be able to stop a build.
func loadMeta(fsys fs.FS) (*Meta, error) {
	raw, err := fs.ReadFile(fsys, MetaFile)
	if err != nil {
		// The read error is dropped rather than wrapped: the file is optional, so
		// not having one and not being able to read one are the same answer, and
		// reporting the second would make an unreadable file fail a theme that
		// declares nothing.
		return nil, nil
	}

	var meta Meta
	md, err := toml.Decode(string(raw), &meta)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", MetaFile, err)
	}
	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, len(undecoded))
		for i, key := range undecoded {
			keys[i] = key.String()
		}
		meta.warnings = append(meta.warnings, fmt.Sprintf("%s has unknown key(s): %s", MetaFile, strings.Join(keys, ", ")))
	}

	// Every message quotes meta.Cress rather than a trimmed copy, so a theme
	// author reads back the text they wrote, the same way config.Load quotes an
	// author's own base_url.
	if strings.TrimSpace(meta.Cress) != "" {
		meta.declared, meta.declaredOK = parseVersion(meta.Cress)
		if !meta.declaredOK {
			meta.warnings = append(meta.warnings, fmt.Sprintf("%s declares %s = %q, which is not a version, so the theme's contract cannot be checked", MetaFile, metaVersionKey, meta.Cress))
		}
	}
	return &meta, nil
}

// Warnings reports what is worth saying about the theme's metadata to somebody
// running the cress version named by running, which is empty or unparseable for
// a build that carries no version. Reading the file is settled at load; only the
// comparison needs to know what is running, and it is skipped rather than
// guessed when there is nothing to compare against.
//
// Every result is a warning and never an error. A theme whose author was merely
// conservative about the version they claimed still builds the site.
func (m *Meta) Warnings(running string) []string {
	if m == nil {
		return nil
	}
	warnings := append([]string(nil), m.warnings...)
	if mismatch := m.compatibility(running); mismatch != "" {
		warnings = append(warnings, mismatch)
	}
	return warnings
}

// compatibility compares the contract the theme declared against the cress
// running it, returning an empty string when there is nothing to say.
//
// The two cases it reports are the two ways one declaration can be wrong, and
// they cannot both hold: a cress older than the theme asks for may not offer
// every field the templates read, and a cress a whole major version newer no
// longer promises the contract the theme was written against. Both fail the same
// quiet way, because a template renders a field it does not recognize as an
// empty string rather than refusing.
func (m *Meta) compatibility(running string) string {
	if !m.declaredOK {
		return ""
	}
	current, ok := parseVersion(running)
	if !ok {
		return ""
	}
	switch {
	case current.less(m.declared):
		return fmt.Sprintf("%s declares %s = %q, but this is cress %s; the theme may read fields this version does not provide", MetaFile, metaVersionKey, m.Cress, running)
	case current.major > m.declared.major:
		return fmt.Sprintf("%s declares %s = %q, but this is cress %s; the template contract the theme was written against is not promised across a major version", MetaFile, metaVersionKey, m.Cress, running)
	}
	return ""
}

// semver is a version reduced to the three numbers worth comparing.
type semver struct {
	major, minor, patch int
}

// less reports whether v orders before other.
func (v semver) less(other semver) bool {
	if v.major != other.major {
		return v.major < other.major
	}
	if v.minor != other.minor {
		return v.minor < other.minor
	}
	return v.patch < other.patch
}

// parseVersion reads a version leniently enough to compare: an optional "v"
// prefix, one to three dot-separated numbers, and any pre-release or build
// metadata after them.
//
// Omitted components are zero, so a theme may write the "1.1" its contract is
// actually about rather than padding it to a patch release it does not care
// about. Pre-release ordering is deliberately ignored, so cress 1.1.0-alpha.1
// satisfies a theme asking for 1.1: the alpha of a version is where its contract
// gets built, and warning the people testing it would be noise.
//
// Anything else fails, which is how the untagged local build ("dev") ends up
// with no version to compare rather than a special case to remember.
func parseVersion(s string) (semver, bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	if cut := strings.IndexAny(s, "-+"); cut >= 0 {
		s = s[:cut]
	}
	if s == "" {
		return semver{}, false
	}

	parts := strings.Split(s, ".")
	if len(parts) > 3 {
		return semver{}, false
	}
	var numbers [3]int
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil {
			return semver{}, false
		}
		numbers[i] = n
	}
	return semver{major: numbers[0], minor: numbers[1], patch: numbers[2]}, true
}
