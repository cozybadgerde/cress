# Cress product requirements

Status: approved product direction; scaffold exists, generator not implemented.

Target release: 1.0.0.

Product owner: Daniel Schier (Cozy Badger).

Date: 2026-10-16.

## Product purpose

Cress is an open-source static site generator for small-business websites.
It turns local content, configuration, and a theme into deployable static
files. It does not deploy those files.

The primary maintainer is a technical person at Cozy Badger or an
open-source enthusiast. The website serves a solopreneur or small business,
such as a restaurant, coach, photographer, or freelance consultant.

Business owners can prepare Markdown in their preferred editor and send it
to the maintainer. The maintainer builds locally or through continuous
integration and continuous delivery (CI/CD).

This document describes product needs, scope, and observable acceptance
criteria. The [ADRs](adr/) provide architectural guidance and guardrails.
The [work plan](../todo.md) tracks remaining decisions and implementation.
Agreed field names and observable behavior are part of the product
contract. Remaining authoring syntax, package layouts, algorithms, and
setup instructions belong to implementation and its documentation.

## Product principles

- Prefer predictable, documented behavior over implicit conventions.
- Keep the initial site structure and central configuration easy to grasp.
- Provide curated themes suited to small businesses.
- Keep content separate from presentation and deployment.
- Report mistakes rather than silently changing the author's intent.
- Preserve room for later features without implementing speculative systems.

Cress does not need to match the breadth of Hugo or Jekyll.

## Users and use cases

| User | Need |
|------|------|
| Technical maintainer | Initialize, customize, preview, and generate a site reliably |
| Business owner | Supply Markdown content without learning a template language |
| Theme author | Build and test a theme against a documented contract |
| Site visitor | Access usable, responsive, accessible business information |

The first release supports company pages, services or menus, opening hours,
contact information, and occasional articles. Images support lightweight
photographic content, but a specialized gallery engine is not required.

## Scope

### Required in 1.0

- Ordinary content pages, landing pages, and explicit list pages.
- Main and footer navigation.
- Opening hours, contact information, menus, and services as page content.
- Site identity, brand settings, and page metadata.
- Basic search engine optimization (SEO), a sitemap, and `robots.txt`.
- Title images and inline images, without image transformation.
- Optional captions for inline images, title images, and fenced code blocks.
- One configured language per site.
- Two built-in themes and locally supplied third-party themes.
- Site initialization, static generation, and local preview.
- Theme initialization for theme development.
- Version information and usage help.
- Machine-readable schemas and human-readable authoring references.

### Excluded from 1.0

- Deployment, hosting management, and automatic theme downloads.
- Booking, reservations, payments, shopping, forms, and site search.
- A content management system or browser-based content editor.
- Multilingual content and browser-selected language switching.
- Image resizing, optimization, or automatic responsive image generation.
- Feeds, `llms.txt`, and additional wide/narrow layout variants.
- Dedicated site, theme, and configuration validation subcommands.
- Authored-HTML URL rewriting and optional rewrite switches.
- Dedicated or automatic broken-link checks.
- A generator accessibility-validation command.
- A configurable build output directory.
- Automatic browser reload and injected preview error overlays.
- Future-date scheduling.

External links to services remain ordinary content. They do not imply
integrations with those services. Other unlisted features are not implied
release commitments.

## Authoring and content

Authors use Markdown with TOML frontmatter and one central TOML site
configuration. Site identity, brand values, theme customizations, navigation,
and Markdown rendering options remain distinct concerns.

Content organization is free-form rather than divided into mandatory page
and article trees. Page behavior is explicit, not inferred from business
directory names. Landing and list types are not restricted to section
index files.

Support page titles, descriptions, title images, draft status, and page
types. Site-wide defaults apply only to documented inheritable metadata;
explicit page values override them. There is no directory-level metadata
inheritance. Page type and draft status remain page-specific.

`published_at` records an optional reader-facing publication date;
`updated_at` records an optional last-update date. A creation-date field is
not introduced now. Dates do not schedule publication; draft visibility
remains separate. Date formats belong to the authoring contract defined
during implementation.

Support an optional, page-specific `excerpt` frontmatter field as plain
text, escaped like ordinary metadata. It is separate from the page
description. If absent, provide no excerpt: do not extract body text or
fall back to the description.

Main and footer navigation are explicitly configured, flat, and ordered.
Each theme renders populated groups and omits empty groups.
Support internal page links and external HTTP/HTTPS links. Invalid known
values are errors; checking whether internal navigation targets exist is
not a 1.0 requirement.

Markdown tables are required. Menus and opening hours can be ordinary
tables or prose; Cress does not impose a separate business-data model.
Footnotes are not a 1.0 requirement.

Treat publisher-authored Markdown bodies as trusted content. Allow inline
and block raw HTML without sanitization in builds and local previews.
Publishers must review imported, contributed, or copied HTML; local files
and previews are not a browser-content security boundary.

Pass authored HTML URLs through unchanged. Their correctness and
suitability for the deployment path are the author's responsibility.
Do not repair them or issue automatic broken-link warnings during builds
or previews. Required input validation remains in effect.

Metadata, plain-text labels, and ordinary customization values remain
escaped. HTML in code spans and fenced blocks remains literal example
text. Generation does not execute embedded code or fetch resources
referenced by authored HTML; browsers may execute or load that content.

Syntax highlighting is enabled by default and can be disabled. Recognized
code languages receive build-time highlighting without browser JavaScript.
Untagged and unrecognized languages remain plain code.

### Captions, attribution, and disclosures

Provide optional first-class captions for inline images, title images,
and fenced code blocks without requiring authors to write raw HTML.
Display provided captions with their associated image or code block.
Captions contain author-provided text and links for explanations, creator
credits, disclosures, or a combination of these uses.
Disclosures can include AI-generation notices where applicable.

Keep captions distinct from image alternatives and hover titles.
Caption text and link labels remain escaped; caption links follow the
existing URL policy. Captions do not change code escaping or syntax
highlighting.

Publishers decide when captions or disclosures are needed and supply their
wording. Do not infer creators, licenses, or AI generation, require captions
on every item, or introduce a separate attribution or disclosure data model.
A caption is not a guarantee that a license or legal obligation is met.

### List pages

An explicit list page can be a section index or another page; its filename
does not restrict its type. Render its Markdown body as optional
introductory content before the generated entries. An empty collection is
valid: preserve the introduction and let the theme present the empty list
without an error or prescribed empty-state wording.

Select ordinary pages that are logical direct children of the list's
containing section. Exclude that section's own root page, the list itself,
and all landing and list pages. Do not recurse into deeper descendants.
Selection follows the page hierarchy, not physical Markdown-file depth.

For example, a list in `blog/list.md` can include `/blog/article/` from
either `blog/article.md` or `blog/article/index.md` as alternative source
layouts. `/blog/archive/article/` is outside that list's scope.

Apply the current build or preview's draft visibility before constructing
the collection. Themes receive the resulting entries, not hidden drafts.

| Ordering mode | Default direction |
|---------------|-------------------|
| Publication date, the default mode | Newest `published_at` first. |
| Resolved page title | A-Z. |

When dates or titles are equal, sort by published route in ascending order.
Reversing either mode reverses the complete included sequence.
`updated_at` does not affect ordering.

In chronological mode, warn about each otherwise eligible visible page
without `published_at` and omit it from the list. The page itself remains
generated normally. Do not invent a date; invalid known values still
follow the normal error policy. Alphabetical lists include undated pages
without missing-date warnings.

Entries contain prepared page metadata, including `excerpt` when provided,
not full article bodies or automatically generated excerpts. Themes choose
how to present the available metadata without changing membership or order.
Manual collections can be ordinary Markdown links; no manual weights or
general collection/query system is required.

Frontmatter field names for selecting and reversing an order remain
implementation choices.

## Routes and assets

Generate portable directory-index routes that do not depend on
hosting-specific extensionless-file rewrites.

Support deployment at both a domain root and beneath a URL prefix.
Cress-resolved navigation, page and asset URLs, canonical URLs, and sitemap
entries must resolve correctly in both cases. Authored HTML URLs remain
author-controlled.

The deployment path defaults to `/`; non-root publication requires an
explicitly configured path. A public origin is recommended for publication
metadata, but not required for ordinary browsing. Generator-owned internal
browsing links remain independent of that hostname.

Resolve site-root Markdown links and images beneath the configured
deployment path, and source-relative Markdown links and images to their
published destinations. External URLs remain unchanged.

Support page-owned assets and shared static files without transforming
their contents. Preserve asset ownership and relative organization without
confusing assets with content sections.

Reject conflicting output destinations rather than silently choosing a
source by precedence or build order. Report the conflicting sources.
Do not publish private configuration or theme sources, or allow paths to
escape their intended boundaries.

## Themes

Provide two built-in themes:

- A coach/freelancer theme.
- A more corporate small-business theme.

Both support the required content capabilities, responsive layouts, and
the accessibility target. Concrete names and visual designs are not
selected here.

A theme is a named presentation package. Bundled and third-party themes
have the same role and follow the same integration contract. Bundling
changes delivery, not architectural ownership.

Built-in themes work without separate runtime files or downloads.
Provide an offline, minimal renderable theme skeleton for theme
initialization. It is developer scaffolding, not a third built-in theme.

Support trusted local third-party themes. A selected local theme takes
priority over a same-name built-in theme as a complete package, without
implicit merging or fallback from an invalid local package.

Themes present prepared content and metadata; changing a theme must not
redefine content selection, URLs, inheritance, or draft visibility.
Support ordinary, landing, and list rendering without silently substituting
another page type when required support is missing.

Cress prepares list membership, visibility, order, and entry data.
Each theme renders the introduction and prepared entry sequence without
discovering, filtering, or reordering source pages.

Cress renders inline-image and code-block captions as content and prepares
title-image caption data. Each theme renders provided title-image captions
and styles captioned content without dropping its text or links.
These responsibilities apply equally to bundled and third-party themes.

Keep ordinary brand settings distinct from expert theme customizations.
Themes document the custom settings they consume; schemas for arbitrary
customizations are not required.

Theme labels follow the configured site language. Missing requested labels
or unavailable translations fail clearly rather than produce silent or
mixed-language substitutions.

Reject incompatible themes before rendering, with actionable diagnostics.
Only designated public theme assets appear in generated output.

## Maintainer workflow

The CLI groups operations by scope, with site work under `cress site`
and theme development under `cress theme`. Version information and common
help remain accessible at the root. Site initialization produces a
ready-to-build site; theme initialization produces a renderable local
package. Neither silently overwrites user files.

Generation works offline with local inputs and available themes, produces
deployable static files, and does not deploy them.

Normal builds exclude drafts. Preview includes drafts only when explicitly
requested for that session, not inherited from site configuration.
Rebuilding preserves that choice. Communicate draft visibility clearly.

Each preview session generates and serves its own output without requiring
or modifying deployable output. A session must not overwrite another
session's output or change its draft choice.

Source edits trigger automatic rebuilding. Invalid source edits report
actionable errors while the last successful preview remains available and
the server keeps running. The latest source changes must not be lost,
including edits that arrive during a rebuild. Generated output must not
trigger rebuilds.

If replacing preview output fails but previous output is successfully
restored, keep serving that verified output and watching for edits.
If replacement succeeds but cleanup fails, retain and serve the verified
new output.
Report either failure without claiming a successful rebuild.

If replacement or recovery leaves no verified usable output, stop serving
and exit unsuccessfully. Preserve recoverable content and evidence, and
provide actionable recovery guidance rather than speculative repairs.

An established preview keeps an automatic rebuild pending if another
writer blocks it, and retries when writer ownership becomes available.
Each running session must catch up to the latest source when generation
can succeed, without another edit or restart. Report the blocked rebuild
and pending retry without claiming success. This retry behavior applies
only to a busy writer, not other failures. Explicit builds and initial
preview startup fail immediately if another writer holds ownership.

Report rebuild outcomes in the terminal. Browser refresh is manual;
generated pages contain no injected reload client, error overlay, or
reload-related HTML modifications.

Store preview output in private temporary session directories. Clean up
finished or verifiably inactive sessions without deleting active sessions,
unrelated files, or ambiguous recovery evidence. Cleanup failures must be
visible.

Build and preview-startup errors have an unsuccessful exit status suitable
for CI. Established previews report rebuild failures without presenting
incomplete output as successful. Unknown-key warnings alone do not fail an
otherwise valid build.

## Output safety

Cress replaces its generated output as a complete result, removing stale
files. Manual additions to recognized generated output are not preserved.

Validation, rendering, or copying failures must leave previous successful
output untouched. Do not overwrite unrelated directories or follow unsafe
output paths merely because they use an expected name.
Verify Cress ownership before replacing nonempty output. Refuse
unrecognized nonempty directories, symlinks, and non-directory targets
without changing them.

Output replacement and interruption handling must preserve recoverable output,
report failures, and avoid guessed destructive recovery. Ambiguous states
require actionable guidance rather than automatic deletion.

Coordinate competing writers so they cannot corrupt a site's output or
recovery state. Read-only preview serving must not prevent generation.
Replacement may interrupt readers; it is not guaranteed to look like one
instantaneous switch. Cress does not guarantee that output or recovery
state survives power loss.

## Contracts and validation

Publish JSON schemas and human-readable references for site configuration,
frontmatter, and theme manifests. Schemas describe the same contract as the
corresponding Cress release.

Malformed configuration, frontmatter, or theme manifests are errors.
Known fields used incorrectly are errors. Unknown fields produce warnings.
Keys inside the documented open theme-customization area are custom data,
not unknown core fields.

Make published schemas usable by compatible external validation tools,
without prescribing an editor or workflow. Schemas describe parsed data;
tools must parse TOML or extract frontmatter from Markdown as needed.
Specific editor integrations and custom language servers are not required.
Schema checks cover parsed field structure and values. They do not replace
generation-time checks such as theme compatibility and output safety.

For supported site and theme contracts, preserve documented fields,
interfaces, behavior, and defaults across Cress 1.x minor and patch
releases. Breaking these contracts requires a new major release.
Older executables need not support new features.

Outside the documented open theme-customization area, unknown fields are
tolerated, not supported extensions. They have no stable-major
compatibility guarantee. Minor releases may recognize them and apply new
validation and behavior, even if previously warning-only inputs then fail
or change their output. Patch releases do not introduce newly recognized
fields.

New fields must preserve documented behavior and defaults when absent.
Within a stable major, Cress must not claim keys inside the documented
open customization area as core fields. Announce newly recognized fields
and their validation rules in release notes, with matching release schemas
and authoring references.

## Quality and distribution

### Accessibility

Cress owns the markup it generates directly from Markdown and preservation
of author-supplied accessibility information. Each theme owns its page
structure, navigation markup, styling, focus presentation, contrast, and
responsive behavior, whether bundled or third-party. Publishers own content
structure, image alternatives, and the effects of custom HTML and styles.

Target Web Content Accessibility Guidelines (WCAG) 2.2 Level AA for complete
default pages generated with each theme shipped with Cress, covering
supported page types and default interactions. Known applicable Level A or
AA failures in assessed default pages block release until fixed.

Use linters and continuous integration (CI), with unit tests as the default
and integration tests where boundary coverage adds meaningful confidence.
Automated end-to-end tests and browser automation are outside the initial
scope. Complement automation with documented human accessibility review
before 1.0 and after relevant generated-markup or theme-behavior changes.
Record scope, environments, results, and limitations.

Document author responsibilities for image alternatives and content
structure, and limitations of custom colors, raw HTML, and third-party
themes. Assessment covers the reviewed default pages. It does not certify
a publisher's finished site or prove legal compliance.

Do not treat a passing build or automated check as certification.
Additional AAA goals, broader compatibility coverage, and usability studies
can follow, but do not justify deferring known baseline failures.
Concrete tools and assessment coverage belong to implementation planning.
Generator accessibility-validation commands are outside 1.0 scope.

### Search engine optimization

Support page titles, descriptions, site defaults, canonical URLs, title
images, a sitemap, and `robots.txt`. Additional structured data and social
metadata fields are not implied.

Without a public origin, generation succeeds with an actionable warning
and omits metadata that requires that origin. Do not invent a hostname or
advertise an omitted sitemap. Ordinary content and browsing remain usable.

### Platforms and releases

Support Linux for CI/build environments and Linux/macOS for development.
Initial release targets are macOS Apple Silicon, Linux ARM64, and Linux
Intel/AMD 64-bit. Windows, Intel macOS, and 32-bit ARM are not 1.0 targets.
Generated files are independent of the generator's operating system.

Distribute versioned releases under BSD-3-Clause through GitHub Releases.
Provide downloadable archives with the executable, required licenses and
notices, and installation and verification instructions.

Users must be able to verify release authenticity and artifact integrity,
including the identity of the expected release publisher.
Publish a release-specific machine-readable software bill of materials
(SBOM) covering bundled code and assets, protected by release verification.

Review direct and transitive dependencies and non-code assets under their
own licenses. Release checks reject unreviewed components and unapproved
licenses. The project license does not grant rights to third-party assets
or Cozy Badger trademarks.

Apple Developer ID signing and notarization are deferred. Document the
macOS platform trust limitation rather than promise warning-free
installation. Containers and Homebrew are future distribution options.

## Acceptance criteria for 1.0

1. An initialized site builds offline using either built-in theme.
2. Both themes render ordinary, landing, and list pages, contact information,
   opening hours, menu tables, images, and ordered main/footer navigation.
3. Free-form content organization produces portable directory-index routes.
   Cress-resolved links and assets work at a domain root and beneath a URL
   prefix; authored HTML URLs are not rewritten.
4. Page-owned and shared assets are copied without transformation. Private
   sources stay private, and output collisions identify the sources.
5. A local third-party theme builds offline. Invalid or incompatible themes
   fail clearly without silent fallback.
6. Invalid known settings fail generation; unknown keys warn. Schemas agree
   with runtime validation and support external validation of parsed site
   configuration, frontmatter, and theme manifests without a required editor.
   Within a stable major, only minor releases may introduce newly recognized
   fields, with documented validation and behavior; formerly warning-only
   inputs may then fail.
   Preserve documented contracts and open theme customizations.
7. Normal builds exclude drafts. Draft-inclusive and draft-free previews
   are isolated from one another and from deployable output.
8. Preview starts without existing deployable output and rebuilds after
   source edits. Invalid edits retain the last successful preview and report
   errors. Recoverable publication or cleanup failures keep verified output
   available; unverifiable output stops serving and ends the session
   unsuccessfully with recovery guidance. Automatic rebuilds retry writer
   contention so competing sessions catch up without another edit or restart;
   explicit builds and initial preview startup remain fail-fast.
   Manual refresh shows successful changes without injected reload behavior.
9. Preview cleanup protects active sessions, unrelated files, and ambiguous
   recovery evidence, and reports failures.
10. Successful builds remove stale output. Generation failures preserve
    previous output; publication and interruption failures preserve
    recoverable state and refuse unsafe or ambiguous destructive actions.
    Competing writers cannot corrupt output.
11. Site and theme initialization work offline without silent overwrites.
    Version information, usage help, and maintainer workflows are documented.
12. Complete default pages produced with shipped themes meet the documented
    accessibility release baseline. Record automated and manual assessment,
    scope, environments, results, limitations, and author responsibilities;
    known applicable A or AA failures in assessed default pages block release.
13. Releases cover the supported targets and work without separate built-in
    theme files. Authenticity and integrity verification accepts genuine
    artifacts and rejects tampering or unexpected publisher identities.
14. Release SBOMs cover bundled components and their license reviews.
    Required licenses and notices are supplied, and unreviewed or unapproved
    components fail release checks.
15. Titles, descriptions, title images, canonical URLs, sitemap entries,
    and `robots.txt` reflect the documented metadata and deployment location.
    Without a public origin, generation warns and omits origin-dependent
    metadata while preserving ordinary browsing.
16. Syntax highlighting works without browser JavaScript, can be disabled,
    and leaves untagged or unrecognized code languages plain.
17. Metadata uses documented site defaults and explicit page overrides,
    without directory-level inheritance of metadata or page behavior.
18. Themes preserve required page-type behavior and the configured language.
    Missing rendering support or requested translations fail clearly.
19. Publisher-authored raw HTML renders without sanitization in builds and
    previews. HTML examples remain literal, ordinary values stay escaped,
    and generation neither executes embedded code nor fetches resources
    referenced by authored HTML. Authored HTML URLs pass through unchanged,
    without broken-link checks.
20. Optional captions for inline images, title images, and fenced code blocks
    display author-provided text and links with the associated content in
    both shipped themes. Keep image alternatives separate and preserve code
    escaping and syntax highlighting.
21. List pages select eligible logical direct children, apply current draft
    visibility, and provide reversible publication/title ordering with
    deterministic route ties. Chronological lists omit undated entries with
    warnings; alphabetical lists retain them. Introductions and empty lists
    render normally. Explicit plain-text excerpts are optional, with no
    automatic extraction or description fallback.

These criteria define readiness, not an implementation sequence.
No performance threshold or adoption target has been agreed.

## Decision records

Accepted ADRs provide architectural guidance, rationale, and guardrails.
They do not imply implementation. Details explicitly left open belong to
implementation planning. Examples are illustrative unless stated as
requirements.

- [0001: Implementation language](adr/0001-go-implementation-language.md).
- [0002: Offline generation and deployment boundary](adr/0002-offline-generation-and-deployment-boundary.md).
- [0003: Free-form content and directory-index routes](adr/0003-content-tree-and-routes.md).
- [0004: Shared contract policy](adr/0004-shared-contract-policy.md).
- [0005: Authoring formats](adr/0005-toml-authoring-formats.md).
- [0006: Configuration and metadata model](adr/0006-configuration-and-metadata-model.md).
- [0007: Theme packages and integration](adr/0007-theme-packages-and-integration.md).
- [0008: Generation pipeline and page renderer](adr/0008-rendering-and-template-trust-boundary.md).
- [0009: Scope-first CLI commands](adr/0009-workflow-cli.md).
- [0010: Complete output, ownership, and recovery](adr/0010-staged-output-and-ownership.md).
- [0011: BSD-3-Clause and license review](adr/0011-bsd-3-clause-license.md).
- [0012: Signed releases with GoReleaser](adr/0012-release-artifacts.md).
- [0013: Explicit site navigation](adr/0013-explicit-site-navigation.md).
- [0014: Public origin and deployment path](adr/0014-public-origin-and-deployment-path.md).
- [0015: Trusted Markdown and raw HTML](adr/0015-trusted-markdown-and-raw-html.md).
- [0016: Accessibility baseline and assessment](adr/0016-accessibility-baseline-and-assessment.md).
