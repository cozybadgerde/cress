# Trust publisher-authored Markdown and raw HTML

Status: accepted.

Date: 2026-10-15.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Page authors sometimes need presentation that Markdown alone cannot
express, such as styled images or small product cards. Requiring theme
changes or another authoring language for each case would limit that work.

Cress generates publisher-controlled sites, not a service accepting public
content submissions. However, a file being local does not establish its
trustworthiness: copied snippets, imports, and contributions can contain
unwanted browser behavior.

The authoring model needs a clear boundary between publisher-controlled
HTML, ordinary text values, and code the generator may execute.

## Decision drivers

- Give authors an HTML escape hatch without another component language.
- Match the publisher-controlled source model.
- Avoid promising protection that contextual escaping does not provide.
- Keep ordinary values escaped and generation free of embedded execution.
- Make publisher responsibility and preview risks explicit.

## Considered options

- Trust publisher-controlled Markdown bodies and allow unsanitized HTML.
- Sanitize HTML to an allowed formatting subset.
- Disable raw HTML and require other presentation mechanisms.

## Decision outcome

Chosen option: "Trust publisher-controlled Markdown bodies and allow
unsanitized HTML", because publishers control the site's source and need
an escape hatch beyond Markdown's presentation capabilities.

Allow inline and block HTML in Markdown bodies by default, without a
sanitization filter. Publishers must review content obtained from other
sources before previewing or publishing it.

Authored HTML is an opaque escape hatch. Do not repair or rewrite its
URLs; authors own their correctness. Markdown links and images follow
[the URL-resolution policy](0014-public-origin-and-deployment-path.md).

This accepts active browser content, not merely formatting. Scripts,
event handlers, embedded resources, and styles can affect visitors and
local previews. Cress does not promise cross-site scripting (XSS)
protection or a sandbox for publisher-authored HTML.

### Generation and rendering boundaries

- Content preparation converts the trusted Markdown body into rendered
  HTML. The page renderer composes that prepared HTML with the theme.
- Keep metadata, plain-text labels, and ordinary customization values
  contextually escaped. Trusting the body does not make arbitrary values
  trusted HTML.
- HTML inside code spans and fenced code blocks remains literal example
  text, not active markup.
- Do not execute embedded scripts or other authored code during generation,
  or fetch remote resources referenced by authored HTML.
- Browser-side execution and resource loading are separate from generator
  capabilities. Local preview follows the same content trust policy.

Supporting browser-side scripts does not add generator plugins, build
scripts, or access to filesystem, environment, shell, or network helpers.
It also does not promise offline browsing of externally referenced
resources.

### Scope

Custom HTML and styles can affect accessibility and theme portability.
Their use does not guarantee consistent presentation or a conforming
finished site.

This record settles trust and authored-HTML pass-through, not all Markdown
capabilities.
[First-class captions](../PRD.md#captions-attribution-and-disclosures)
provide author-supplied text and links without requiring raw HTML or
changing the escaping of ordinary values.
[The accessibility baseline](0016-accessibility-baseline-and-assessment.md)
defines assessment obligations without changing the content trust policy.
The Markdown library and concrete render-model types remain implementation
choices.

### Consequences

- Good, because authors can use HTML without changing a theme.
- Good, because the policy matches publisher control of site sources.
- Good, because no sanitization guarantee needs to be maintained.
- Bad, because unsafe source HTML can harm visitors or preview users.
- Bad, because publishers must review contributed and copied content.
- Bad, because custom markup can reduce accessibility and portability.

## Confirmation

During implementation, confirm inline and block HTML render in builds and
previews without sanitization. Check that HTML examples inside code spans
and fenced blocks remain literal and ordinary text values remain escaped.

Confirm generation does not execute embedded code or fetch resources
referenced by authored HTML. Confirm HTML URLs are not rewritten or checked
for broken targets. Check that author documentation explains browser-side
risks and author-owned URLs, and does not describe local preview as a
sandbox.

## More information

See [rendering responsibilities and trust guardrails](0008-rendering-and-template-trust-boundary.md),
[public URL policy](0014-public-origin-and-deployment-path.md),
and [the product requirements](../PRD.md).
