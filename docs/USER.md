# Using Cress

Cress currently provides a development walking skeleton, not the complete
1.0 feature set. It builds ordinary pages through an embedded developer
theme skeleton. That skeleton is not a selectable built-in theme.

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
Unknown fields produce warnings and have no effect. Page types and site
metadata defaults are not implemented yet.

The [frontmatter schema](../schemas/frontmatter.schema.json) describes
parsed TOML values for this development version. External tools must
extract and parse the frontmatter before applying the schema. Unknown
fields are permitted by the schema; Cress reports their warnings.
Schemas are versioned with the source and will track release versions.

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

Only deployment at `/` is supported in this slice. Markdown destinations
remain as authored until URL resolution arrives in phase 3. For now, use
published URLs such as `[About](/about/)`, not source filenames.

## Markdown and trust

CommonMark and tables are supported. Raw inline and block HTML is trusted
and passes through unchanged, including its URLs. Code spans and fenced
blocks remain escaped literal text. Metadata remains escaped separately.

Review imported or contributed HTML before building or opening it.
Generation does not execute embedded code or fetch referenced resources,
but browsers can do both. Local files are not a content sandbox.

Syntax highlighting, captions, asset copying, navigation, publication
metadata, and theme selection are not implemented yet. `cress.toml`,
`static/`, and `themes/` are reserved for later phases and are not loaded.

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
