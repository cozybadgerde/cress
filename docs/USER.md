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

Create a `content/` directory. Each lowercase `.md` file outside an asset
directory is a page. Other regular files outside designated asset roots
are not published. Symlinks and special files inside
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
Sources are site-root paths, absolute HTTP/HTTPS URLs, or source-relative
assets on a page. Site defaults cannot use source-relative paths.
See [content and assets](#content-and-assets-contract) for asset publication
and caption syntax.

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
its fields and rendering contract are phase 5 work. Selectable theme
packages also arrive in phase 5. These unimplemented settings warn as
unknown rather than silently taking effect.

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

Markdown source references resolve to published destinations, and
site-root URLs receive the configured prefix. Authored HTML URLs always
pass through unchanged.

## Markdown and trust

CommonMark and tables are supported. Raw inline and block HTML is trusted
and passes through unchanged, including its URLs. Code spans and fenced
blocks remain escaped literal text. Metadata remains escaped separately.

Review imported or contributed HTML before building or opening it.
Generation does not execute embedded code or fetch referenced resources,
but browsers can do both. Local files are not a content sandbox.

List generation, publication metadata, and theme selection are not
implemented yet. `themes/` remains reserved and is not loaded.

## Content and assets contract

Use explicit page-owned asset directories named after the source
stem: `about.md` owns `about.assets/`; `about/index.md` owns
`about/index.assets/`. Publish these beside the owning page's `index.html`,
omitting the `.assets` directory name. Preserve nested asset paths and
contents. Orphan asset directories fail. Draft-owned assets stay private.
Files in asset directories are opaque assets, not content pages.

Shared `static/` files publish without the `static/` prefix. Files outside
designated asset roots remain private. Reject symlinks and special files
in either kind of asset tree. All published files share collision checks,
including case-folded and canonically equivalent destinations.

Resolve source-relative `.md` links and asset references relative to the
Markdown file, preserving query strings and fragments. A source reference
must map to a published page or asset; missing sources, boundary escapes,
and excluded draft destinations fail. This is source-to-output resolution,
not a general broken-link check. Site-root URLs need not name an existing
file; external URLs remain unchanged. Fragment-only and query-only links
stay local. Title images may use source-relative assets on a page; site
image defaults remain site-root or external.

For example, `content/about.assets/team.jpg` publishes as
`public/about/team.jpg`. Link to it from `content/about.md` with
`![Team](about.assets/team.jpg)`. A shared `static/images/logo.svg`
publishes as `public/images/logo.svg`; reference it with `/images/logo.svg`
or a source-relative path such as `../static/images/logo.svg` from
`content/index.md`. CSS and other asset contents are copied verbatim;
Cress does not rewrite their URLs.

### Captions

Use an immediate caption attribute after a Markdown image:

```markdown
![Our team](about.assets/team.jpg){caption="Photo by [Alex](https://example.org/)"}
```

Use the same attribute in a code fence's opening info line:

````markdown
```go caption="Example from [the guide](https://example.org/)"
fmt.Println("Hello")
```
````

For title images, add a `caption` string to `[image]` or `[site.image]`.
Image overrides still replace the entire table, including its caption.
An omitted or empty caption adds no caption markup.

Captions support text and Markdown links only. Raw HTML is escaped, not
trusted; other Markdown formatting and media are rejected. Caption links
support HTTP/HTTPS, site-root paths, and source-relative destinations,
including local queries and fragments. Inherited site image captions use
site-root or HTTP/HTTPS destinations. Image alternative text and hover
titles remain separate. Quoted attributes support `\"` and `\\` escapes.
Malformed explicit caption attributes fail with source context. Captions
must fit on one line. Captioned images inside Markdown links or image
alternatives fail rather than producing nested or inaccessible links.

Standalone captioned images and fenced code use `figure` and `figcaption`.
Images inside prose use inline markup with an associated caption, not a
figure inside a paragraph. No creator, license, or disclosure is inferred.

### Highlighting

Highlighting is enabled by default. Disable it in `cress.toml` for plain
code:

```toml
[markdown]
highlight = false
```

Recognized language names and aliases are highlighted during generation.
Untagged and unknown languages stay plain without guessing. No browser JavaScript
or resource fetching is used. Code always remains literal escaped text.

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
