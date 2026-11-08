# Keep generation maintainable with a separate page renderer

Status: accepted.

Date: 2026-10-15.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Building a site involves content rules, Markdown processing, page rendering,
and output publication. These responsibilities change for different reasons.
If build coordination, content rules, and template logic are mixed, small
changes can spread across the generator and every theme.

Cress needs generation code that remains understandable, testable, and
maintainable over time without maintaining unnecessary custom machinery.

## Decision drivers

- Keep code understandable as content features and themes evolve.
- Limit coupling so presentation changes do not require content-rule changes.
- Keep build coordination separate from content rules and page rendering.
- Give each responsibility a clear boundary that can be tested independently.
- Prefer standard-library capabilities when they meet the need.
- Preserve consistent content behavior and explicit trust guardrails.

## Considered options

- A generation pipeline with a separate page renderer.
- Let theme rendering depend on generator internals and content rules.

## Decision outcome

Chosen option: "A generation pipeline with a separate page renderer",
because clear responsibilities reduce coupling and make the code easier
to maintain over time.

Prefer standard-library functionality where it meets the requirements.
Use Go's `html/template` for page composition and contextual escaping rather
than introducing another template engine or maintaining our own.

### Pipeline responsibilities

| Component | Responsibility |
|-----------|----------------|
| Generator | Coordinates the build and propagates failures. Delegates content preparation, rendering, and output handling rather than implementing all their rules. |
| Content preparation | Resolves metadata, URLs, navigation, drafts, and lists, and converts Markdown into rendered HTML. |
| Page renderer | Composes prepared page/site data and theme templates into page HTML. Does not discover content or decide content rules. |
| Output handling | Stages and publishes generated pages and assets under its own ownership and recovery rules. |

Cress is the generator/builder; the pipeline above belongs to Cress.
Each theme is a named presentation package consumed by the renderer, not
another application layer. This applies to bundled and third-party themes.
Markdown-to-HTML conversion belongs to content preparation; page rendering
composes that content with the theme's layouts and partials.

Conceptual coordination and data flow, not a prescribed package layout or API:

```text
Generator coordinates:
  |
  +-- Content preparation
  |       |
  |       v
  |   Prepared data
  |       |
  +-- Page renderer <--- Theme templates
  |       |
  |       v
  |   Page HTML
  |       |
  +-- Output handling <--- Other generated files and assets
```

The renderer consumes a dedicated, documented render model: the prepared
page and site data that templates may use. It does not receive raw
configuration or generator internals. Keep core data distinct from open
customizations. Neither the renderer nor templates query the full content
tree or reach back into content preparation to redefine its rules.

These are responsibilities, not a requirement for a full layered framework
or a package and interface for every box. Introduce interfaces where a real
test seam or alternate implementation needs them. Concrete packages, data
fields, and loading procedures belong to implementation.

### Markup and presentation ownership

Cress owns the markup it generates directly from Markdown and preservation
of author-supplied accessibility information. Each theme owns its page
structure, navigation markup, styling, focus presentation, contrast, and
responsive behavior. Bundled and third-party themes have the same
presentation responsibilities.

Publishers own content structure, meaningful image alternatives, and the
effects of custom HTML, styles, and configuration. Authored HTML remains
publisher-controlled under the content trust policy.

First-class inline-image and code-block captions belong to
generator-produced content markup. Cress prepares title-image caption data
for the selected theme to render. Themes style captioned content and
preserve provided caption text and links, whether bundled or third-party.
[The caption requirements](../PRD.md#captions-attribution-and-disclosures)
define the authoring capability. Exact syntax, field names, and render-model
types remain implementation choices.

[The accessibility decision](0016-accessibility-baseline-and-assessment.md)
builds on these boundaries.

### Page-type composition

Require explicit rendering support for ordinary, landing, and list pages
in every theme, including the skeleton. Select templates from the resolved
page type, not folder names. Missing required support is an error, not
permission to silently substitute another type.

Content preparation owns list scope, visibility, ordering, and entry
metadata. The selected theme renders the optional introductory body and
prepared entry sequence, including valid empty collections. It does not
discover, filter, or reorder source pages or construct excerpts.
[The list-page requirements](../PRD.md#list-pages) define that behavior.

Page-type-specific template overrides apply only to that type. Shared
layouts and partials may intentionally affect several types; they must not
silently substitute one page type's layout for another. This is a
template-selection rule, not a sandbox for shared CSS or publisher-authored
HTML.

### Trust guardrails

- Template sources are trusted, not sandboxed; third-party templates require
  review before use.
- Do not expose filesystem, network, environment, or shell capabilities
  through template data or helpers, or execute theme-supplied scripts or
  plugins during generation.
- Keep rendered HTML distinct from ordinary metadata. Metadata,
  customizations, and plain-text labels retain contextual escaping; do not
  bypass escaping for arbitrary values.
- Contextual escaping does not make malicious templates safe or sanitize
  Markdown. [The content trust policy](0015-trusted-markdown-and-raw-html.md)
  permits publisher-controlled raw HTML without sanitization.
- Template failures stop generation with actionable errors.

### Consequences

- Good, because responsibilities can be understood and tested independently.
- Good, because build coordination does not absorb content and template rules.
- Good, because theme changes do not redefine content behavior.
- Good, because standard templates reduce custom code and dependency upkeep.
- Bad, because the documented render model must stay aligned with both
  content preparation and the page renderer.
- Bad, because third-party template sources still require review.

## Confirmation

When implemented, check that content preparation can be tested without
templates and the page renderer consumes only the documented render model
and selected theme templates. Confirm the generator coordinates the work
without moving content rules into orchestration or rendering.
Check that theme changes leave content behavior intact, page-type
composition stays isolated, and the trust and failure guardrails hold.

## More information

See [theme integration](0007-theme-packages-and-integration.md),
[configuration and metadata](0006-configuration-and-metadata-model.md), and
[output handling](0010-staged-output-and-ownership.md).
See [Go's template security model](https://pkg.go.dev/html/template#hdr-Security_Model).
