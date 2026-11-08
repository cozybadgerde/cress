# Add first-class captions to Markdown

Status: accepted.

Date: 2026-10-20.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Business sites need captions for photo credits, licenses, and disclosures.
CommonMark has no caption syntax, so publishers would otherwise write raw
HTML. That HTML is trusted and opaque, bypasses theme styling, and makes
accessible associations the author's task.

## Decision drivers

- Avoid raw HTML for a common publishing need.
- Associate captions accessibly with what they describe.
- Allow links in captions for credits and licenses.
- Keep values escaped and the syntax small.

## Considered options

- A `caption` attribute on images, code fences, and title images.
- Use the image title as the caption.
- Leave captions to raw HTML.

## Decision outcome

Chosen option: "A `caption` attribute on images, code fences, and title
images", because it covers the needs with one small, explicit syntax.
Image titles keep their tooltip meaning, and raw HTML stays opaque.

Captions use one-line Markdown limited to text and links. Cress rejects
other syntax with a source-aware error instead of dropping it.

| Target | Syntax |
|--------|--------|
| Image | `![Alt text](photo.jpg){caption="Photo: [Jane](https://example.com/)"}` |
| Code fence | `` ```go caption="Handler setup" `` |
| Title image | `caption` in the page's `[image]` frontmatter |

A standalone image or code block renders as `figure` with `figcaption`.
An inline image gets a group with `aria-describedby` that points to its
caption, using an ID in the reserved `cress-*` namespace.

### Consequences

- Good, because publishers caption content without raw HTML.
- Good, because themes style one consistent caption structure.
- Good, because captions are associated accessibly by default.
- Bad, because the attribute syntax is specific to Cress.
- Bad, because captions do not support Markdown beyond text and links.

## Confirmation

Unit tests cover the three targets, standalone and inline images, links,
escaping, and errors for unsupported syntax. Each shipped theme styles
captions in its default pages.

## More information

The [content guide](../CONTENT.md) explains the syntax. See the
[content trust policy](0015-trusted-markdown-and-raw-html.md) and
[heading targets](0017-markdown-heading-targets.md) for the reserved IDs.
