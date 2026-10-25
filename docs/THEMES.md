# Theme development

Themes are complete embedded or local packages. The embedded skeleton exercises all
page types and supplies offline initialization; it is not a selectable third
built-in theme. Poppy is the default; clover is the other built-in.
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

The required JSON manifest uses full semantic versions without `v`:

```json
{
  "name": "my-theme",
  "version": "1.0.0",
  "cress": { "minimum": "1.0.0", "major": 1 }
}
```

`name` must match the selected directory. `version` identifies the theme.
The runtime must match `cress.major` and meet `cress.minimum`, which must
belong to that major. Prereleases follow semantic-version precedence:
`1.0.0-rc.1` does not satisfy `1.0.0`. Development runtimes reporting `dev`
use the explicit `1.0.0` baseline with a warning, not a compatibility bypass.
Unknown manifest fields warn; invalid known fields fail. See the
[manifest schema](../schemas/theme.schema.json).

Only a missing local package permits built-in lookup for poppy or clover.
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
The initialized theme is a minimal skeleton, not a copy of poppy or clover.

When working in the Cress repository, use `./bin/cress` after
`mise run build` and put the site under ignored `tmp/`, for example
`tmp/theme-test`; create `tmp/` first.
See [developer test sites](DEVELOPER.md#test-sites) for the shared workflow.

## Prepared render model

The renderer accepts a prepared `page.Page` value:

| Field | Go type | Meaning |
|-------|---------|---------|
| `Title` | `string` | Validated page title, escaped by `html/template`. |
| `Description` | `string` | Resolved plain-text description after site inheritance and page overrides. |
| `Type` | `page.Type` | Explicit `ordinary`, `landing`, or `list` behavior. |
| `URL` | `string` | Escaped browsing path, including the site URL prefix and a trailing slash. |
| `Image` | `*page.Image` | Optional resolved title image with `URL`, escaped plain-text `Alt`, and prepared `Caption` parts. |
| `Body` | `template.HTML` | Trusted HTML produced by content preparation from the Markdown body. |
| `Site` | `page.Site` | Prepared shared presentation data. |
| `Canonical` | `string` | Absolute publication URL, or empty without `site.host`. |
| `PublishedAt` | `string` | Optional validated `YYYY-MM-DD` publication date; empty when absent. |
| `UpdatedAt` | `string` | Optional validated `YYYY-MM-DD` update date; empty when absent. |
| `Excerpt` | `string` | Optional escaped plain-text excerpt; never derived from the body or description. |
| `Entries` | `[]page.Entry` | Prepared membership and order for a list page; no entry bodies. |

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

`page.Entry` exposes `Title`, `URL`, `Description`, `Image`, `PublishedAt`,
`UpdatedAt`, and `Excerpt` with the same prepared meanings as page fields.
Entries do not contain page bodies, nested entries, raw source paths, or
selection settings. Themes must preserve membership and order.

The public origin is not needed for browsing and is not exposed as raw
configuration. Templates receive only pages visible in the current build
or preview, including drafts only in explicit draft-inclusive preview.
They receive no raw source metadata or unfiltered content tree.
Custom data remains ordinary data, not trusted
HTML; the developer skeleton does not interpret it.

Content preparation alone marks the rendered body as trusted HTML.
Ordinary strings must not be converted to trusted HTML.

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

`internal/theme/builtin/poppy/` and `clover/` each contain their own complete
templates, public CSS, manifest, and exact `en`/`de` labels. They do not
merge with the skeleton or with each other. Neither consumes custom keys.
Same-name local packages replace them completely.

Their stable manifest baseline is Cress 1.0.0. Embedded packages ship with
the executable and are tested together; local packages undergo runtime
compatibility checks, including during prerelease development.
See the [accessibility assessment](ACCESSIBILITY.md) for automated evidence
and the pending complete-page human review.
