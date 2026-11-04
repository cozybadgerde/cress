# Theme development

Themes are complete embedded or local packages. The embedded skeleton exercises all
page types and supplies offline initialization; it is not a selectable
built-in theme. Poppy is the default; Clover and Anemone are alternatives.
See [built-in presentation and labels](USER.md#built-in-themes).

## Package contract

```text
themes/my-theme/
  theme.json
  templates/
    page.gohtml
    landing.gohtml
    list.gohtml
    shared.gohtml
  public/
    site.css
  locales/
    en.toml
```

Run `cress theme init my-theme [site-directory]` to create this structure
without overwriting an existing target. Initialization supplies minimal CSS
and no translated labels, since its templates request none. Select the
package with `[theme] name = "my-theme"` in `cress.toml`.

Open the generated `theme.json` for a manifest initialized by your
executable. The [manifest schema](../schemas/theme.schema.json) lists its
required fields and allowed values. Version fields use full semantic
versions without `v`.

`name` must match the selected directory. `version` identifies the theme.
The runtime must match `cress.major` and meet `cress.minimum`, which must
belong to that major. A prerelease does not satisfy a minimum set to its
later stable version. Development runtimes reporting `dev` use the
baseline in [compatibility validation](../internal/theme/manifest.go)
with a warning, not a compatibility bypass.
Unknown manifest fields warn. Invalid known fields fail.

Only a missing local package permits built-in lookup for poppy, clover, or anemone.
Unreadable, invalid, incompatible, or incomplete local packages
never fall back. No package files merge. Symlinks and special files anywhere
inside a selected package are rejected.

Templates are UTF-8 Go `html/template` files below `templates/`; nested
partials retain their relative names, such as `partials/header.gohtml`.
Every package requires nonempty `page.gohtml`, `landing.gohtml`, and
`list.gohtml`. Template errors and missing map keys fail generation.
Templates remain trusted sources requiring publisher review.

Two presentation-only helpers are available:

- `{{asset "css/site.css"}}` requires an existing file at
  `public/css/site.css` and returns its prepared, URL-escaped
  `<site.path>themes/<name>/css/site.css` browsing URL.
- `{{label "read_more"}}` reads a string from `locales/<site.language>.toml`,
  whose top level is a plain-text label map, such as
  `read_more = "Read more"`. Matching is exact: `en-US` never falls back to
  `en`. A missing requested language or key fails, without empty/key-name
  substitution. An explicitly authored empty label remains empty.

Packages that request no labels need no locale file. A supplied matching
locale must parse as a string map even when unused. Private files, manifests,
templates, and locale sources never publish. Public files copy unchanged
into the theme namespace, including nested paths, with normal collision
checks. CSS URLs are not rewritten.

## Testing a theme

Start with a disposable site containing every page type and the content
tour, then add a local theme:

```sh
cress site init theme-test
cress theme init my-theme theme-test
```

In `theme-test/cress.toml`, change the existing `[theme]` table's name to
`"my-theme"`. Build and preview:

```sh
cress site build theme-test
cress site preview --port 0 theme-test
```

Edit files under `theme-test/themes/my-theme/`; preview rebuilds when they
change. Refresh the browser manually. Review the home landing page, journal
list, ordinary pages, and `/tour/` for typography, images, captions, and code.
The initialized theme is a minimal skeleton, not a copy of a bundled theme.

For source builds and ignored fixture directories, see
[developer test sites](DEVELOPER.md#test-sites).

## Prepared render model

The renderer accepts a prepared `page.Page` value.
The [render-model types](../internal/page/page.go) are the field and type
reference for pages, sites, images, navigation, and entries. The notes below
explain how to use those values without repeating the struct inventory.

`Title` is a validated page title. `Description` is resolved after inheritance
and page overrides. `Image` is optional. Only landing pages inherit the site
image default. `URL` is an escaped browsing path with the site prefix and a
trailing slash. `Canonical` is an absolute publication URL, or empty when no
public origin is configured.

`PublishedAt` and `UpdatedAt` are validated date-only strings, or empty when
absent. `Excerpt` is explicit plain text, never extracted from the body or
description. `Body` is trusted HTML from content preparation.

`Site` contains plain-text `Title`, `Copyright`, and `Language`; resolved
home `URL`; ordered `Navigation.Main` and `Navigation.Footer` slices of
`page.Link` values with `Label` and `URL`; and open `Custom` TOML data.
The supplied language spelling is preserved. Missing site title or
copyright is represented by an empty string, and missing navigation groups
have no entries.

`Site.Brand` holds optional `Color`, `Favicon`, and `Logo` values. Colors
are validated `#RRGGBB` strings; asset URLs are already prefixed or external.
`Logo` uses the image URL/alternative fields without captions. Do not
reinterpret brand settings as routing or content-selection rules.

`Image.Caption` is an ordered slice of `page.CaptionPart` values. Each part
has escaped plain-text `Text` and an optional resolved `URL` for a link.
Render the caption beside its image with `figure` and `figcaption`; omit
caption markup when the slice is empty. Caption parts are ordinary data,
not trusted HTML.

`Image.Focus` is empty or a keyword from the
[image schema](../schemas/frontmatter.schema.json). It names the part to
keep in view when a theme crops the image. Built-in themes
add a `focus-<value>` class and keep that third in view; themes that do not
crop may ignore it.

The render model does not expose intrinsic image width or height. Built-in
themes reserve space for fixed crops and logos through CSS; uncropped images
may shift surrounding content while loading. This applies to local raster
images, SVGs, and remote or unknown-size images alike. Generation does not
fetch remote resources to determine dimensions. Keep uncropped presentations
proportional rather than inventing dimensions or imposing a crop solely to
reserve space.

Entries use the same prepared meanings as page fields. They do not contain
page bodies, nested entries, raw source paths, or selection settings.
Themes must preserve prepared membership and order.

The public origin is not needed for browsing and is not exposed as raw
configuration. Templates receive only pages visible in the current build
or preview, including drafts only in explicit draft-inclusive preview.
They receive no raw source metadata or unfiltered content tree.
Custom data remains ordinary data, not trusted
HTML; the developer skeleton does not interpret it.

Content preparation alone marks the rendered body as trusted HTML.
Ordinary strings must not be converted to trusted HTML.

### Syntax highlighting

`.Site.Syntax.Highlight` is the resolved site-wide syntax-highlighting choice
from `[syntax] highlight`, enabled by default. It controls visible syntax
coloring, not just token generation. Themes must respect `false`: do not
apply token coloring or activate an alternative highlighter. Ordinary code
fonts, backgrounds, padding, and scrolling remain theme-owned.

Fenced code blocks render inside `<pre><code>`. A specified language adds
`language-<name>` to `<code>`, preserving the authored language name. When
highlighting is enabled, Chroma adds token `<span>` elements with CSS classes.
Common token classes include `k` for keywords, `s` for strings, and `c` for
comments. Chroma's class-based output leaves token colors to theme CSS; the
bundled themes style these classes in their public stylesheets. Untagged or
unrecognized languages remain escaped plain code, and highlighting can be
disabled with `[syntax] highlight = false`. The language class remains when
disabled; its presence is not permission to highlight. The markup has no
Chroma-specific IDs or `data-*` attributes. See Chroma's
[HTML formatter documentation](https://github.com/alecthomas/chroma#the-html-formatter)
for formatter details.

The initialized theme demonstrates conditional token-color CSS using
`{{if .Site.Syntax.Highlight}}`. A separate stylesheet can be linked likewise:

```html
{{if .Site.Syntax.Highlight}}
<link rel="stylesheet" href="{{asset "highlight.css"}}">
{{end}}
```

Place that stylesheet at `public/highlight.css`. Target token spans beneath
`pre code`; Cress does not emit Chroma's usual `.chroma` wrapper. Common
subtypes include `kd` (keyword declaration), `s2` (double-quoted string), and
`c1` (single-line comment). Not every token has a span or a class.
Highlight.js styles use different classes and are not drop-in Chroma styles.
If a custom theme adds an alternative highlighter, honor the same switch and
do not reprocess already-tokenized Chroma blocks. Cress and its supplied
themes need no browser highlighter.

### Other prepared markup

Markdown paragraphs containing only one direct link receive
`class="standalone-link"`, independently of page type and position.
CommonMark autolinks and reference-style links qualify, as does formatting
inside a link. Surrounding text, additional links, or formatting wrapped
around the link do not. Tight-list text blocks are not paragraphs and
receive no class. Authored HTML remains unchanged.

This hook describes structure, not a primary action or button. Themes may
combine it with page type, position, or other presentation rules, or ignore
it. Cress does not choose its appearance or change the link's semantics.

Strikethrough in `Body` uses `<del>`. Task lists contain native disabled
checkboxes, each wrapped with its first paragraph's inline content in a
`<label>`. Preserve labels, completion state, list structure, and item links.
Do not enable the checkboxes or add task-state persistence. Native styling
is the default; themes may style these elements without hiding their state.

Markdown headings in `Body` already have prepared link targets. Preserve
them without generating replacement IDs. `main`, `page-title`, and the
`cress-*` namespace are reserved for theme and generated markup; other
theme IDs must not collide with content headings. Avoid generated caption
IDs such as `cress-caption-1` even inside the reserved namespace. Raw HTML
IDs remain publisher-controlled. See [heading links](USER.md#heading-links).

Markdown tables retain their table markup inside a
`<div class="table-wrapper" tabindex="0">`. This wrapper provides a styling
hook and keyboard focus target. Themes own overflow handling, spacing, and
focus styling; Cress supplies no table layout styles. Authored HTML tables
remain unchanged.

Explicit Markdown column alignment becomes `align-left`, `align-center`,
or `align-right` classes on header and body cells, not inline styles.
Unspecified alignment adds no class. Themes decide how to display this
author-provided intent; the bundled themes and skeleton apply matching
`text-align` rules.

The renderer receives no raw configuration, source tree, or generator
internals. It neither discovers pages nor resolves routes. The skeleton
uses `html/template` with no filesystem, network, environment, or shell
helpers, and template errors stop generation.

## Embedded skeleton

The executable embeds the package in `internal/theme/skeleton/`.
Ordinary, landing, and list pages select distinct templates, with shared
document composition. Missing or unknown type support fails, without substitution.
Lists render their introduction before linked entry titles, explicit dates,
and excerpts. Empty collections add no list markup or prescribed wording.

The skeleton renders the configured document language, page title,
description, title image and caption, site identity, main/footer links,
copyright, main landmark, and prepared body. Empty optional groups are omitted.
It adds no hardcoded translated labels or translation fallback. The footer
links use a list; main links use a navigation landmark.
Minimal inline styles keep code readable without browser JavaScript.
Canonical links are included only when the prepared URL is nonempty.

## Built-in packages

`internal/theme/builtin/poppy/`, `clover/`, and `anemone/` each contain complete
templates, public CSS, manifest, and exact `en`/`de` labels. They do not
merge with the skeleton or with each other. None consumes custom keys.
Same-name local packages replace them completely.

Their landing CSS emphasizes the body's first paragraph and a directly
following paragraph that contains only one link. This is a styling
convention over the neutral `standalone-link` hook, not primary-action metadata.
Anemone uses an underlined action rather than a button. It also supplies
the side-masthead layout, image framing, rotation, and shadows entirely in
its templates and CSS, without generator-owned layout rules.

Each package's manifest owns its version and compatibility declarations.
Embedded packages ship with the executable and are tested together.
Local packages undergo runtime compatibility checks, including during
prerelease development.
See the [accessibility assessment](ACCESSIBILITY.md) for automated evidence
and the pending complete-page human review.
