# Theme development

Theme package integration is not implemented yet. The current embedded
skeleton exercises ordinary and landing rendering during development; it
is not a selectable third built-in theme.

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

`Site` contains plain-text `Title`, `Copyright`, and `Language`; resolved
home `URL`; ordered `Navigation.Main` and `Navigation.Footer` slices of
`page.Link` values with `Label` and `URL`; and open `Custom` TOML data.
The supplied language spelling is preserved. Missing site title or
copyright is represented by an empty string, and missing navigation groups
have no entries.

`Image.Caption` is an ordered slice of `page.CaptionPart` values. Each part
has escaped plain-text `Text` and an optional resolved `URL` for a link.
Render the caption beside its image with `figure` and `figcaption`; omit
caption markup when the slice is empty. Caption parts are ordinary data,
not trusted HTML.

The public origin is not needed for browsing and is not exposed as raw
configuration. No template receives draft pages, raw source metadata, or
the unfiltered content tree. Custom data remains ordinary data, not trusted
HTML; the developer skeleton does not interpret it.

Content preparation alone marks the rendered body as trusted HTML.
Ordinary strings must not be converted to trusted HTML.

The renderer receives no raw configuration, source tree, or generator
internals. It neither discovers pages nor resolves routes. The skeleton
uses `html/template` with no filesystem, network, environment, or shell
helpers, and template errors stop generation.

## Embedded skeleton

The executable embeds the templates in `internal/render/skeleton/`.
Ordinary and landing pages select distinct templates, with shared document
composition. Missing or unknown type support fails, without substitution.
A list template is scaffolded but rendering and generation fail explicitly
until prepared list entries arrive in phase 4.

The skeleton renders the configured document language, page title,
description, title image and caption, site identity, main/footer links,
copyright, main landmark, and prepared body. Empty optional groups are omitted. It
adds no hardcoded translated labels or translation fallback. The footer
links use a list; main links use a navigation landmark.
Minimal inline styles keep code readable without browser JavaScript.

These templates form a development interface, not the complete
theme contract. Phase 5 adds complete packages, manifests, compatibility,
translations, initialization, branding, and complete required page-type
support. Phase 8 supplies and assesses the two selectable built-in themes.
