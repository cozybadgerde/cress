# Using Cress

Cress builds ordinary, landing, and list pages offline using the built-in
`poppy`, `clover`, and `anemone` themes or complete local theme packages.
Poppy is the default when no theme is selected. Release readiness and human
accessibility review are tracked in the [work plan](../TODO.md).

## Commands

```sh
cress --help
cress --version
cress site --help
cress theme --help
cress site init my-site
cress site build
cress site preview --port 0
```

`cress site build [directory]` builds the selected site directory, defaulting
to the current directory. It reads `content/`, writes generated files to
`public/`, and keeps private writer state in `.cress/` inside that directory.
An unsuccessful build exits with status 1.
Warnings alone do not fail a build.

`cress theme init <name> [site-directory]` creates `themes/<name>/` from the
embedded skeleton; the site directory defaults to `.`. The target must not
exist. Partial initialization failures leave files for inspection, not
automatic removal or overwriting. The command does not change `cress.toml`.
Select the result with `[theme] name = "<name>"`.

`cress site init [directory] [--theme poppy|clover|anemone]` creates a small
English sample site in a new or completely empty directory; the default is `.`.
The parent directory must exist. Targets that are files, symlinks, or
nonempty directories are refused, including hidden files. No merge, Git
initialization, automatic build, or network access occurs. Partial failures
are reported and left for inspection.

```sh
cress site init my-site
cress site build my-site
cress site preview --port 0 my-site
```

The scaffold includes `cress.toml`, `.gitignore` for `public/` and `.cress/`,
seven example pages covering all page types, and an original local SVG.
The generated configuration documents every supported site setting, with
commented examples for optional images, branding, and theme customizations.
Uncomment optional tables and their fields together when needed.
The content tour's frontmatter holds the main authoring hints; other sample
pages explain only distinct features, such as list settings.
The footer links to this tour of Markdown, captions, title images, and metadata;
compare its preview with `content/tour.md`. Remove that page and its footer
link when you no longer need the reference.
Edit the sample identity, content, contact details, and navigation before
publication. Set `site.host` to your public HTTP/HTTPS origin for canonicals
and sitemap output; it is intentionally unset initially and causes a warning.
Use `--theme clover` during initialization for the small-business layout.
Use `--theme anemone` for the paper/Scandi editorial layout.

## Testing a site

Use the same initialization command for a disposable test site as for a
real project. From a directory where you keep projects, choose a new name:

```sh
cress site init test-site
cress site build test-site
cress site preview --port 0 test-site
```

Open the printed URL and compare the content tour with `content/tour.md`.
Edit `test-site/cress.toml` and the sample content to try your own settings.
Preview rebuilds on changes; refresh the browser manually. Stop with Ctrl-C.
Omitting `site.host` is fine for local testing; the warning reminds you to
set it before publishing.

To compare built-ins, initialize another new directory with
`cress site init --theme clover test-clover` or
`cress site init --theme anemone test-anemone`. Initialization never resets
or overwrites an existing nonempty project. Keep real project sources
outside disposable directories, and deploy only their built `public/` output.

## Built-in themes

Poppy uses a chalk off-white background, dark green palette, calm sans headings,
and generous single-column spacing for freelancer sites. Clover uses a
white/slate and dark blue layout with a compact solid title block, measured
sans-serif headings, and fine section rules for an established small-business
presentation. A wide outer frame gives title blocks and title images more
space, while a centered reading column keeps prose narrow and text left-aligned.
Filled primary links keep the next action clear.
Clover centers ordinary and list-page headings on the reading column.

Anemone uses warm paper surfaces, dark ink, and editorial serif headings.
On wide screens a side masthead sits beside the reading column; on narrow
screens it sits above it. Framed images have soft shadows. Body images rotate
slightly on wide screens; title images, text, and captions never rotate.
On wide screens a landing title image sits beside the title, height-capped.
Page title images retain their proportions without cropping. Pages work without
photographs. It suits makers, independent shops, cafes, and creative businesses.

All three use system fonts, local CSS, no scripts, and no remote assets.
All support every page type and captioned content.
New sites enable gray logo and favicon placeholders in `static/images/`.
Replace them before publishing, or remove their branding settings.
Each theme shows the logo beside the site title: in Poppy's header, Clover's
tinted header, and Anemone's masthead. A compact square box contains square,
circular, or other artwork without cropping. The placeholder has empty
alternative text because the adjacent site title supplies the identity.
The logo and title form one home link with a shared keyboard-focus outline.
When both are present, themes treat the logo as decorative to avoid repeating
the identity. Without a title, the link uses the logo's alternative text,
or a localized "Home" label when that text is empty. Long titles wrap beside
the logo. Square or circular artwork works best; wide or tall artwork fits
inside the same box and may appear smaller.
Optional publication and update dates share a compact, muted row above
ordinary and list-page titles. Landing-page and list-entry dates remain
below their titles.

On landing pages, all three themes enlarge the first paragraph. The paragraph
directly after it is shown as the primary action when it contains only
one direct Markdown link: outlined in Poppy, filled in Clover. Formatting
inside the link is allowed. Anemone instead gives that link a bold underline
and more space, without a filled button. Paragraphs with surrounding text or
additional links stay ordinary prose. This is a theme choice using Cress's neutral
`standalone-link` paragraph class, not a generator-selected primary action.

In Poppy and Clover, a landing page's title image is cropped so the
introduction stays near the top. Poppy shows a wide 5:2 band below the title.
Clover shows the same band under the title block on narrow screens; on wide
screens the title block and image sit side by side, and any caption stays
below the image. Crops keep the centre unless the image sets `focus`. A
`[site.image]` default also appears there unless the page overrides it.

On ordinary and list pages, Poppy crops title images to 16:9 and Clover to
3:2. Clover also caps their height at the smaller of 36rem and 70% of the
viewport height. These crops reserve space before images load. Anemone
keeps title images uncropped, so their space can change as they load.

List entries crop their images to a fixed shape for easy scanning: a 5:2
band in Poppy, a 4:3 thumbnail left of the text in Clover, and a framed 4:3
print beside the text in Anemone. On narrow screens Clover and Anemone stack
the image below the date. Entry images load lazily; the list's own title
image loads immediately.

Select a theme with `[theme] name = "poppy"`, `"clover"`, or `"anemone"`.
A complete same-name local package takes priority, including over the default
Poppy.
Invalid local packages never fall back. The embedded development skeleton
is used only by `theme init`, not as a selectable theme.

Built-in labels support exact `en` and `de`. Regional tags such as `en-US`
or `de-CH` do not fall back; use a local theme with the matching locale.
Page content is not automatically translated. No built-in consumes
`theme.custom` keys. All use the brand color as their accent. See the
[accessibility assessment](ACCESSIBILITY.md) for review status and publisher
responsibilities.

## Local preview

```sh
cress site preview
cress site preview --drafts --port 0 my-site
```

`cress site preview [directory]` defaults to the current directory and port
8080. It generates its own private temporary output, never requiring or
modifying deployable `public/`. Multiple sessions keep separate outputs.
The server binds only to `127.0.0.1`; `--port 0` chooses a free port and
prints it. An occupied port fails rather than silently selecting another.
There is no remote-binding flag or automatic browser launch.

Drafts are excluded unless `--drafts` is explicitly supplied. This choice
belongs to the running session and is preserved on every rebuild, including
draft assets and list membership. Included drafts need a title and other
publishable metadata; Cress does not invent missing values. Normal builds
always exclude drafts. The terminal prints the session's draft mode.

Open the printed URL. Preview preserves `site.path`, serving only beneath
that prefix and redirecting `/` to it. Canonicals and sitemap URLs retain
the configured `site.host`, not the local server origin. A successful
change to `site.path` prints the new preview URL. No authored HTML or copied
asset contents are rewritten. Trust local preview content just as you
would published content; raw HTML and theme templates remain trusted.

Native filesystem events watch `cress.toml`, `content/`, `static/`, and
`themes/`, including nested and newly created directories. Atomic editor
saves are supported. Generated `public/` and `.cress/` files do not trigger
rebuilds. A short debounce coalesces edits; edits made during a build queue
another full build. Native watching requires a filesystem that provides
events; network and virtual filesystems may not do so. Watcher failures,
including event overflow, stop unsuccessfully rather than silently losing
edits. No polling fallback is provided.

Rebuild outcomes appear in the terminal. Refresh the browser manually.
No reload client or error overlay is injected. Invalid inputs leave the
last verified preview available and watching continues. A rebuild blocked
by another site writer stays pending and retries about every 500 ms without
needing another edit. Other failures are reported but do not automatically
retry. Explicit builds and initial preview startup remain fail-fast.
Read-only serving does not hold the site writer lock. During a rebuild,
HTTP requests may wait until publication or restoration finishes.

Verified previous output remains available after a failed publication
that restores it. Confirmed new output remains available after cleanup
failure. Both rebuilds are reported as unsuccessful. If usable output
cannot be verified, preview stops serving and exits unsuccessfully,
preserving evidence and naming its temporary session directory.

Ctrl-C or SIGTERM shuts down the server and cleans verified session output.
Each session has owner-only temporary storage, a site-bound identity, and
an OS lifetime lock. On startup, Cress reclaims only recognized sessions for
the same canonical site whose lifetime lock can be acquired. Active sessions
and unrelated entries are preserved. Ambiguous recovery state is preserved
and reported, not guessed away. Cleanup errors fail the command and may
leave a session for later inspection or verified reclamation.
Do not delete session locks to force cleanup. There is no persistent
preview cache or guarantee against OS temporary-storage cleanup.

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
| `type` | String | `ordinary` by default; also accepts `landing` and `list`. |
| `image` | Table | Sets the title image. On landing pages it replaces the complete inherited `site.image`; `{ src = "" }` clears it. |

An image table requires a string `src`. A nonempty source also requires
an explicit string `alt`; an empty alternative marks a decorative image.
An optional `focus` names the part to keep in view when a theme crops the
image: `center` (default), `top`, `bottom`, `left`, `right`, `top-left`,
`top-right`, `bottom-left`, or `bottom-right`. For example, use
`focus = "top"` for a portrait whose face sits in the upper third.
Sources are site-root paths, absolute HTTP/HTTPS URLs, or source-relative
assets on a page. Site defaults cannot use source-relative paths.
See [content and assets](#content-and-assets-contract) for asset publication
and caption syntax.

## Dates, excerpts, and lists

`published_at` and `updated_at` are optional page-specific calendar dates.
Use TOML local dates such as `published_at = 2026-10-20`, or equivalent
`"YYYY-MM-DD"` strings. Empty strings, invalid calendar dates, times, and
timestamps fail, including on drafts. Dates never schedule publication.
No ordering relationship between the two fields is required.

`excerpt` is optional plain text, escaped in templates. An absent or empty
excerpt has no output; Cress never extracts body text or uses the description
as an excerpt. Dates and excerpts do not inherit from site configuration.

Create a list with explicit `type = "list"`:

```toml
title = "Journal"
type = "list"

[list]
order = "published_at"
reverse = false
```

The `[list]` table is optional on lists and invalid on other page types.
`order` is `"published_at"` by default, newest first, or `"title"` for
case-insensitive Unicode-normalized lexical order. This is not
language-specific dictionary ordering. Equal keys sort by published URL
ascending. `reverse = true` reverses the complete sequence, including ties.

A list selects visible ordinary pages directly beneath its containing
logical section. For `blog/list.md` or `blog/list/index.md`, the section
is `/blog/`; for `blog.md` or `blog/index.md`, it is `/`. A list at `/`
selects direct children of `/`. Exclude the section's
own root, the list itself, landing pages, other lists, excluded drafts,
and deeper descendants. Explicit draft-inclusive preview includes eligible
draft entries. Equivalent source layouts select the same entries.

Chronological lists warn about and omit eligible pages without
`published_at`; those pages still publish normally. Title-ordered lists
include undated pages without that warning. `updated_at` never affects
ordering.

The Markdown body appears above the prepared entries. Each entry exposes
its title, URL, description, image/caption, dates, and optional excerpt,
but not its body. All built-ins render linked titles, supplied images,
explicit dates, and excerpts. Empty lists preserve their introduction without
prescribed empty-state wording.

## Publication metadata

With `site.host`, every visible page receives an absolute canonical URL.
`public/sitemap.xml` includes all visible pages, ordered by published URL;
`lastmod` is present only when `updated_at` is explicitly supplied. No
filesystem timestamps or build dates are inferred. Canonicals and sitemap
URLs combine the public origin and already-prefixed browsing URL once.

`public/robots.txt` permits crawling and advertises the absolute sitemap
URL. Without `site.host`, Cress issues one actionable warning, omits
canonicals and the sitemap, and writes `robots.txt` without a sitemap line.
These generated files share collision checks with pages and assets;
conflicting `static/robots.txt` or `static/sitemap.xml` fails rather than
overriding either source when the corresponding file is generated.

Crawlers discover `robots.txt` only at a host's root. For subpath
deployments, the host-root robots policy and any sitemap advertisement
must be coordinated through external deployment; Cress does not write
outside the selected output directory.

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
# Theme-specific data; no built-in consumes custom keys.
```

| Site field | Default | Meaning |
|------------|---------|---------|
| `title` | Omitted | Optional site identity, distinct from required page titles; supplied values must contain non-whitespace text. |
| `description` | `""` | Default page description; explicit page values override it. |
| `copyright` | `""` | Plain-text footer line; no inferred year or legal wording. |
| `language` | `"en"` | Valid BCP 47 language tag; preserve spelling for exact theme-label matching. Built-ins support `en` and `de`. |
| `host` | Omitted | Public origin, including HTTP/HTTPS scheme and optional port; not merely a hostname. |
| `path` | `"/"` | Web URL prefix, not a filesystem destination or deployment operation. |
| `image` | Omitted | Default title image for landing pages, using the same complete-table contract as frontmatter. |

`site.title`, `site.copyright`, and `site.language` are site-only settings.
Only the description and title image inherit into page metadata. The
description reaches every page; the title image reaches landing pages only,
so ordinary pages, list pages, and list entries show only their own images.
Page titles, types, and draft status do not inherit.

`host` cannot contain credentials, a query, a fragment, or a path other
than an optional final `/`. The prepared origin omits that final slash.
An explicitly empty host is invalid; omit it when unknown.
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

### Branding and theme selection

Optional brand settings stay separate from `site` and `theme.custom`:

```toml
[brand]
color = "#346b38"
favicon = "/images/favicon.svg"

[brand.logo]
src = "/images/logo.svg"
alt = "Example business"

[theme]
name = "my-theme"
```

`color` is exactly six hexadecimal digits after `#`; its authored case is
preserved. Logo and favicon URLs are site-root or HTTP/HTTPS references.
They receive the site URL prefix like navigation; targets are not checked.
A logo requires explicit string `src` and `alt`; `alt = ""` is decorative.
Omit optional settings instead of supplying empty colors or asset URLs.
Themes choose how to present brand data. All built-ins render the logo,
favicon, and theme-color metadata, and use `color` as their accent for
links, buttons, and Clover's title block. Cress warns when `color` has less
than 4.5:1 contrast against white as a reference comparison only. The build
continues. This does not assess the selected theme: dark or tinted themes
may intentionally use colors with low contrast against white. Check the
actual text/background pairs in the rendered theme, whether or not Cress
warns; a warning does not necessarily mean the color needs changing.

Theme names start with a lowercase ASCII letter and contain only lowercase
letters, digits, and hyphens. A selected `themes/<name>/` must be a complete,
readable package. Invalid local packages fail without merging or fallback.
Only designated `public/` files publish, under `themes/<name>/` in output;
all outputs share collision checks. Only a genuinely absent local package
permits built-in lookup for `poppy`, `clover`, or `anemone`.
No selection means `poppy`, including a same-name complete local override.

Named themes declare runtime compatibility in `theme.json`. Development
builds reporting `dev` check against an explicit `1.0.0` compatibility
baseline and warn; they do not skip minimum or major validation. This is
development policy, not a release claim. See the
[theme reference](THEMES.md) for manifests, templates, and translations.

The [frontmatter schema](../schemas/frontmatter.schema.json) describes
parsed page values; the [configuration schema](../schemas/site.schema.json)
describes parsed site values. External tools must parse TOML and extract
frontmatter before applying them. Normalize TOML local dates to date-only
strings; enable date-format assertions for calendar validation.
Unknown fields are permitted by schemas;
Cress reports their warnings. Runtime also validates URL semantics and
language tags. Schemas are versioned with source; select them from the same
release tag as the executable, not from a moving branch. See
[external schema validation](#external-schema-validation).

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
They do not change the emitted names. Excluded drafts do not claim output
destinations; explicitly included preview drafts do.

Names retain their case and are not slugified. URL path segments are
escaped separately from filesystem paths. Empty page names, backslashes,
and path traversal are rejected.

Markdown source references resolve to published destinations, and
site-root URLs receive the configured prefix. Authored HTML URLs always
pass through unchanged.

## Markdown and trust

CommonMark, tables, strikethrough, and static task lists are supported.
Raw inline and block HTML is trusted and passes through unchanged,
including its URLs. Code spans and fenced blocks remain escaped literal
text. Metadata remains escaped separately.

Review imported or contributed HTML before building or opening it.
Generation does not execute embedded code or fetch referenced resources,
but browsers can do both. Local files are not a content sandbox.

Only the selected theme's designated public assets are published.

### Strikethrough and task lists

Use `~~old wording~~` to render deleted text as `<del>old wording</del>`.
Goldmark also accepts matching single tildes (`~old wording~`); prefer
double tildes for clarity. Inline emphasis and links still work inside
strikethrough. Code spans and fenced blocks remain literal.

Do not communicate an important change through strikethrough alone:
screen readers may not announce it. Explain the current information in
words, for example, "Previously Saturday; now Friday."

Begin a list item's text with `[ ]` for incomplete or `[x]` / `[X]` for
complete, followed by a space and meaningful task text:

```markdown
- [x] Agree the project scope.
- [ ] Review the [services](services.md).
  - [ ] Choose a starting date.
```

Ordered lists also support task markers. Each item renders a native
disabled checkbox wrapped with its first paragraph's inline content in
a `<label>`. Subsequent paragraphs and nested lists stay outside that
label. Links remain usable. Ordinary list items remain unchanged.

These are static checklists, not forms. To change completion state, edit
the Markdown and rebuild; visitors cannot toggle or save it. No setting,
script, or extra dependency is required. Native checkbox appearance varies
by browser. Raw HTML remains unchanged and publisher-controlled.

### Heading links

Every Markdown heading receives an ID, including hash-prefixed headings
and headings underlined with `=` or `-`. Link within the page with
`[Opening hours](#opening-hours)`, or to another source page with
`[Opening hours](contact.md#opening-hours)`. Existing URL-resolution rules
preserve fragments; Cress does not check whether a fragment target exists.

For a stable target independent of the heading wording, add an explicit ID:

```markdown
## When to visit {#opening-hours}

[See our opening hours](#opening-hours).
```

Only heading ID attributes are supported. IDs must be nonempty text without
ASCII whitespace or null characters. Explicit IDs retain their spelling,
including case and non-ASCII letters. Duplicate explicit IDs on one page
fail generation, even if they occur after an automatic heading.

Automatic IDs use Goldmark's built-in rules on the heading's final source
line, without its heading markers or attribute block: trim surrounding
whitespace, lowercase ASCII letters, keep ASCII letters and digits, replace
each space, hyphen, or underscore with a hyphen, and drop other characters.
These are source-based IDs, not slugs of rendered text. Non-ASCII characters
are dropped: `Über uns` becomes `ber-uns`. An empty result becomes `heading`.
Use an explicit ID when those rules produce an unsuitable target.

Explicit IDs are reserved first. Automatic headings then receive IDs in
document order, adding `-1`, `-2`, and so on until the ID is unused. For
example, two `## Services` headings produce `services` and `services-1`;
if another heading explicitly uses `{#services}`, they become `services-1`
and `services-2`, regardless of where that explicit heading appears.

`main`, `page-title`, and all IDs starting with `cress-` are reserved for
theme and generated markup. Explicit use fails. Automatic `main` and
`page-title` targets receive numeric suffixes; automatic IDs starting with
`cress-` receive a `heading-` prefix before uniqueness suffixes are assigned.

No visible permalink controls or automatic table of contents are added.
Raw HTML remains unchanged, including its IDs. Publishers must avoid
collisions involving raw HTML and coordinate any custom theme IDs with
their content.

## Content and assets contract

Use explicit page-owned asset directories named after the source
stem: `about.md` owns `about.assets/`; `about/index.md` owns
`about/index.assets/`. Publish these beside the owning page's `index.html`,
omitting the `.assets` directory name. Preserve nested asset paths and
contents. Orphan asset directories fail. Excluded draft-owned assets stay
private; draft-inclusive preview publishes them only inside its session.
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
[syntax]
highlight = false
```

Recognized language names and aliases are highlighted during generation.
Untagged and unknown languages stay plain without guessing. No browser JavaScript
or resource fetching is used. Code always remains literal escaped text.
Themes control syntax colors and code-block styling, but must respect this
switch: disabled means no syntax coloring, including from an alternative
highlighter. Fonts, backgrounds, padding, and scrolling remain styled.

Development configuration migration: replace the former `[markdown]` table
with `[syntax]`. The old table is no longer recognized and warns as an unknown
setting; it does not disable highlighting.

## Output protection

Cress accepts new or empty `public/` directories. Later builds replace
recognized Cress output completely, removing stale files and manual
additions. Put public inputs in the source project, not generated output.
Unrecognized nonempty directories, symlinks, and non-directory targets are
refused without modification.

Ownership pairs the reserved `public/.cress-output.json` marker with private
`.cress/output.json` state. The public marker contains only a format version
and an opaque token, not source paths or secrets. Authored output cannot
claim that marker or a case-equivalent path. Keep `.cress/` out of version
control and deployment; it is created with owner-only directory access.
Do not edit the ownership records or remove the persistent writer lock.
Copying a deployed `public/` directory alone does not establish ownership
in another site project.

Output generated before ownership support is unrecognized. Inspect it and
remove only the old generated `public/` once, keeping source files. There
is no automatic adoption or force switch. Subsequent builds need no manual
output removal.

Each build takes a site-wide OS advisory lock before reading inputs or
recovering output and holds it until the attempt finishes. Concurrent builds
fail immediately with a busy diagnostic, including attempts through path
aliases. Process exit releases ownership; lock-file existence and PID age
are not evidence of an active writer. Do not delete the lock file to bypass
contention. Coordination covers cooperating local processes on the
supported Linux and macOS systems, not distributed writers or manual edits.

Pages and assets are fully prepared, rendered, validated, and written to
`.cress/stage-<token>/` before `public/` changes. Previous output moves to
`.cress/backup-<token>/` during replacement. Input, rendering, or staging
failures preserve previous output. Failed promotion restores verified
previous output where possible; errors report failed restoration as well.
After confirmed publication, cleanup failure keeps valid new output in
place but still fails the command.

The next build, under the writer lock, automatically restores or cleans
only states verified against the transaction and paired ownership tokens.
Ambiguous state stops without guessed promotion or deletion. Diagnostics
name preserved paths. A leftover `.cress/output.next`, an altered marker,
or an unrecognized stage requires inspection rather than automatic cleanup.
Before manual recovery, stop Cress writers and copy the preserved evidence
to a safe location. Compare the transaction's previous/current tokens with
the output markers; restore only a verified backup. Do not simply delete
`.cress/` to bypass an error. No general repair command is available yet.

Replacement can briefly interrupt readers; it is not a universally atomic
switch. Cress does not promise durability through power loss.

## External schema validation

Use a JSON Schema Draft 2020-12 validator with the schemas in `schemas/`
from the same Cress source tag. Keep all three schema files together:
configuration references the frontmatter schema by relative path.

Parse `cress.toml` into an object before applying `site.schema.json`.
Extract only the TOML between the initial standalone `+++` delimiters of
a Markdown file and parse it before applying `frontmatter.schema.json`.
When frontmatter is absent, validate an empty object. Convert TOML local
dates to `YYYY-MM-DD` strings, and enable the validator's date-format
assertions. Apply `theme.schema.json` directly to parsed `theme.json`.

For example, the parsed site input can look like:

```json
{
  "site": { "title": "Example business", "language": "en" },
  "theme": { "name": "poppy" }
}
```

Schemas allow unknown fields and open custom data. Cress itself emits
unknown-field warnings and checks runtime-only conditions, including
language tags, URL semantics, theme compatibility, source references,
collisions, and output ownership. Always run a real build as well.
No particular editor, validation service, or separate Cress validation
command is required.

For local source setup, see the [developer guide](DEVELOPER.md).
