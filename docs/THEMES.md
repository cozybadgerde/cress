# Theme development

Theme package integration is not implemented yet. The current embedded
skeleton exercises ordinary-page rendering during development; it is not
a selectable third built-in theme.

## Prepared render model

The renderer accepts a prepared `page.Page` value. Its initial fields are:

| Field | Go type | Meaning |
|-------|---------|---------|
| `Title` | `string` | Validated page title, escaped by `html/template`. |
| `URL` | `string` | Escaped public browsing path, including a trailing slash. |
| `Body` | `template.HTML` | Trusted HTML produced by content preparation from the Markdown body. |

Content preparation alone marks the rendered body as trusted HTML.
Ordinary strings must not be converted to trusted HTML.

The renderer receives no raw configuration, source tree, or generator
internals. It neither discovers pages nor resolves routes. The skeleton
uses `html/template` with no filesystem, network, environment, or shell
helpers, and template errors stop generation.

## Embedded skeleton

The executable embeds `internal/render/skeleton/page.gohtml`. The template
produces a document title, viewport metadata, a main landmark, an escaped
page heading, and the prepared body. Its current language is English;
configured language support arrives with site contracts.

This ordinary-page template is a development interface, not the complete
theme contract. Phase 5 adds complete packages, manifests, compatibility,
translations, initialization, and explicit ordinary, landing, and list
support. Phase 8 supplies and assesses the two selectable built-in themes.
