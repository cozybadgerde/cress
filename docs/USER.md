# Using Cress

Cress currently provides a development walking skeleton, not the complete
1.0 feature set. It builds ordinary and landing pages through an embedded
developer theme skeleton. That skeleton is not a selectable built-in theme.

## Commands

```sh
cress --help
cress --version
cress site --help
cress theme --help
cress site build
```

`cress site build [directory]` builds the selected site directory, defaulting
to the current directory. It reads `content/` and writes only to `public/`
inside that directory. An unsuccessful build exits with status 1.
Warnings alone do not fail a build.

The `theme` scope currently provides help only. Initialization and preview
commands arrive in later phases.

## Writing pages

Create a `content/` directory. Each lowercase `.md` file is a page.
Other regular files are not published. Symlinks and special files inside
the content tree are rejected, as is a symlink at `content/` or `public/`.
The selected site directory is resolved to its canonical filesystem path.

Use optional TOML frontmatter delimited by standalone `+++` lines at the
start of the file. Both LF and CRLF line endings are supported.

```markdown
+++
title = "About us"
draft = false
+++

We help small businesses.
```

| Field | Type | Behavior |
|-------|------|----------|
| `title` | String | Required for published pages; must contain non-whitespace text when provided. Rendered as escaped text. |
| `draft` | Boolean | Defaults to `false`. Draft pages are excluded from builds and need no title. |

All pages must have valid frontmatter, including drafts. Invalid TOML,
incorrect known field types, and invalid known values fail the build.
Unknown fields produce warnings and have no effect. Known fields are
validated even on excluded drafts. No directory-level inheritance occurs.

| Additional field | Type | Behavior |
|------------------|------|----------|
| `description` | String | Inherits `site.description` when absent; `""` clears it. |
| `type` | String | `ordinary` by default; also accepts `landing` and `list`. Published lists fail explicitly until phase 4. |
| `image` | Table | Replaces the complete inherited `site.image`; `{ src = "" }` clears it. |

An image table requires a string `src`. A nonempty source also requires
an explicit string `alt`; an empty alternative marks a decorative image.
Sources are site-root paths or absolute HTTP/HTTPS URLs. Source-relative
image references, asset copying, and captions arrive in phase 3.

## Site configuration

An optional `cress.toml` in the site directory configures the site.
Only absence uses defaults; malformed, unreadable, symlinked, or non-file
configuration fails. Each supplied known field is validated. Unknown keys
warn at every core table level, except inside the protected open
`theme.custom` table, where arbitrary TOML data is preserved.

```toml
[site]
title = "Example business"
description = "Practical help for small businesses."
copyright = "Copyright 2026 Example business"
language = "en"
host = "https://example.org"
path = "/company/"

[site.image]
src = "/images/team.jpg"
alt = "Our team"

[[navigation.main]]
label = "About"
url = "/about/"

[[navigation.footer]]
label = "Partner"
url = "https://example.net/"

[theme.custom]
# Theme-specific data; not interpreted by the developer skeleton.
```

| Site field | Default | Meaning |
|------------|---------|---------|
| `title` | Omitted | Optional site identity, distinct from required page titles; supplied values must contain non-whitespace text. |
| `description` | `""` | Default page description; explicit page values override it. |
| `copyright` | `""` | Plain-text footer line; no inferred year or legal wording. |
| `language` | `"en"` | Valid BCP 47 language tag; preserve spelling for future exact theme-label matching. |
| `host` | Omitted | Public origin, including HTTP/HTTPS scheme and optional port; not merely a hostname. |
| `path` | `"/"` | Web URL prefix, not a filesystem destination or deployment operation. |
| `image` | Omitted | Default title image, using the same complete-table contract as frontmatter. |

`site.title`, `site.copyright`, and `site.language` are site-only settings.
Only the description and title image inherit into page metadata. Page
titles, types, and draft status do not inherit.

`host` cannot contain credentials, a query, a fragment, or a path other
than an optional final `/`. The prepared origin omits that final slash.
An explicitly empty host is invalid; omit it when unknown. Canonical
URLs, sitemap generation, and missing-origin warnings arrive in phase 4.
Ordinary browsing links do not depend on the configured hostname.

`path` must start and end with `/`. Reject dot segments, repeated slashes,
percent escapes, whitespace, backslashes, queries, and fragments rather
than repairing them. The URL prefix applies to prepared page, navigation,
and title-image URLs. Files still go directly under `public/`, not under
a prefix-named subdirectory.

Navigation uses flat, ordered `main` and `footer` arrays of tables with
nonblank `label` and `url` strings. URLs are published site-root paths
with a single leading slash, or absolute HTTP/HTTPS URLs. Queries and
fragments are allowed. Protocol-relative URLs, other schemes, whitespace,
backslashes, and credentials are invalid. Target existence is not checked.
Both groups preserve configuration order; empty groups are omitted.

Internal URLs always describe paths beneath the site URL root. With
`path = "/company/"`, `/about/` becomes `/company/about/`. An authored
`/company/about/` deliberately becomes `/company/company/about/`; Cress
does not guess whether a prefix has already been included. External URLs
remain unchanged.

The future `[brand]` table will hold brand color, logo, and favicon;
its fields and rendering contract are phase 5 work. `[markdown]` options
arrive in phase 3, and selectable theme packages in phase 5. These
unimplemented settings warn as unknown rather than silently taking effect.

The [frontmatter schema](../schemas/frontmatter.schema.json) describes
parsed page values; the [configuration schema](../schemas/site.schema.json)
describes parsed site values. External tools must parse TOML and extract
frontmatter before applying them. Unknown fields are permitted by schemas;
Cress reports their warnings. Runtime also validates URL semantics and
language tags. Schemas are versioned with source and will track releases.

## Routes

| Source | Output file | Public URL |
|--------|-------------|------------|
| `content/index.md` | `public/index.html` | `/` |
| `content/about.md` | `public/about/index.html` | `/about/` |
| `content/about/index.md` | `public/about/index.html` | `/about/` |
| `content/Our team.md` | `public/Our team/index.html` | `/Our%20team/` |

The last two `about` layouts are alternatives, not two distinct pages.
Duplicate output files and file/directory conflicts fail with both sources
named. Collision checks use Unicode case folding and canonical equivalence
to conservatively reject names that can alias across Linux and macOS.
They do not change the emitted names. Drafts do not claim output destinations.

Names retain their case and are not slugified. URL path segments are
escaped separately from filesystem paths. Empty page names, backslashes,
and path traversal are rejected.

Markdown destinations remain as authored until URL resolution arrives in
phase 3, even when a URL prefix is configured. For now, supply complete
published destinations for body links. Authored HTML URLs always pass
through unchanged.

## Markdown and trust

CommonMark and tables are supported. Raw inline and block HTML is trusted
and passes through unchanged, including its URLs. Code spans and fenced
blocks remain escaped literal text. Metadata remains escaped separately.

Review imported or contributed HTML before building or opening it.
Generation does not execute embedded code or fetch referenced resources,
but browsers can do both. Local files are not a content sandbox.

Syntax highlighting, captions, asset copying, list generation, publication
metadata, and theme selection are not implemented yet. `static/` and
`themes/` remain reserved and are not loaded.

## Output protection

Build only into a new or empty `public/` directory. Cress refuses nonempty
directories, symlinks, and non-directory targets without changing them.
All pages are prepared and rendered, and destinations checked, before
output writing begins.

A filesystem failure while writing can leave incomplete output. Cress
reports that failure; inspect the output before removing it manually.
To rebuild, remove only the generated `public/` directory after verifying
its contents. Existing-output ownership, replacement, recovery, and
per-site writer locking arrive in phase 6. Do not run concurrent builds
for the same site in this development slice.

For local source setup, see the [developer guide](DEVELOPER.md).
