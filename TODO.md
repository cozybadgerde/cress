# Work plan

This is Cress's implementation plan. Product scope lives in the
[PRD](docs/PRD.md), decisions in the [ADRs](docs/adr/), and the overall
structure in the [architecture guide](docs/ARCHITECTURE.md).

Phase 0 is complete, and the implementation phases below are approved.
Implementation details are settled within their phase.

## Phase 0: Requirements, architecture, and implementation planning

### Completed foundations

- [x] Record and review the PRD, product principles, and initial ADRs.
- [x] Agree Go, offline generation, and the separation from deployment.
- [x] Agree free-form content, portable routes, TOML authoring, and site
  metadata defaults with explicit page overrides.
- [x] Agree theme rendering boundaries, prepared data, local-first
  selection, and theme-owned translations.
- [x] Separate Cress's generated-markup responsibilities from named-theme
  presentation and publisher-owned content, equally for bundled and
  third-party themes.
- [x] Agree the WCAG 2.2 AA default-page baseline, release expectations,
  combined automated and manual assessment, and publisher-responsibility
  limits.
- [x] Agree linting and CI with unit-focused tests, integration tests where
  they add meaningful confidence, and no automated end-to-end suite initially.
- [x] Agree trusted publisher-authored Markdown and unsanitized raw HTML,
  with escaped ordinary values and no embedded execution during generation.
- [x] Agree pass-through HTML URLs and resolved Markdown links and images,
  without rewrite switches or automatic broken-link checks in 1.0.
- [x] Agree initial Markdown authoring scope, including captions for inline
  images, title images, and fenced code blocks, with author text and links
  for attribution, disclosures, and explanations.
- [x] Agree optional `published_at` for the reader-facing publication date
  and optional `updated_at` for the last-update date, without introducing a
  creation-date field or scheduling behavior.
- [x] Agree section-local list membership, publication/title ordering,
  draft visibility, missing-date warnings and omission, and intro/empty
  behavior.
- [x] Agree optional plain-text `excerpt` frontmatter without automatic
  extraction or description fallback.
- [x] Agree scope-first CLI command organization for site and theme work.
- [x] Agree optional public origin, explicit deployment paths, and
  generator-owned URL resolution for pages, navigation, and assets.
- [x] Agree isolated previews, automatic rebuilds, manual browser refresh,
  and opt-in draft visibility.
- [x] Agree complete output replacement, ownership protection, writer
  coordination, recoverability, and temporary preview retention.
- [x] Agree preview continuation with verified output after recoverable
  failures, unsuccessful shutdown when output cannot be verified, and
  automatic contention retries for established preview rebuilds without
  changing explicit-build or initial-startup fail-fast behavior.
- [x] Agree schemas, stable-major compatibility, and tool-neutral validation.
- [x] Agree minor-only recognition and strict validation of previously
  unknown fields outside protected open customizations, with matching
  release notes, schemas, and references. Patch releases do not introduce
  newly recognized fields.
- [x] Agree project licensing, component review, and signed release direction.
- [x] Separate product requirements, architectural guidance, and
  implementation planning; remove premature implementation prescriptions.
- [x] Give language, generation, contracts, authoring, theme integration,
  rendering, navigation, CLI scopes, and URL policy focused ADR ownership;
  consolidate and renumber the unpublished collection without changing
  accepted decisions.
- [x] Review all project documents for consistent decisions, realistic
  guarantees, and clear language before implementation planning.

These are approved directions; implementation status is tracked below.

### Implementation foundations

- [x] Create the minimal Go scaffold and add the BSD-3-Clause license.
  No third-party components are bundled yet; add notices as needed.
- [x] Add pinned mise tools, basic developer tasks, repository defaults,
  and a brief agent entrypoint.
- [x] Document the agreed architecture without prescribing speculative
  package layouts or interfaces.
- [x] Add basic CI that runs the existing check and build tasks.
- [x] Finish phase 0 by approving the phased implementation plan below.
  Phases may still be reordered before implementation starts.

## Implementation phases

Each phase delivers a working, tested slice and leaves `trunk` releasable as
a development build. Phases name outcomes, not packages. The bracketed
numbers refer to the PRD's [acceptance criteria](docs/PRD.md#acceptance-criteria-for-10).
A phase can partly cover a criterion; the last phase that lists it
completes it.

Every phase is done when:

- `mise run check` and `mise run build` pass locally. Keep CI tasks aligned
  with local checks; CI execution is not a phase-completion gate for now.
- New behavior has unit tests, and integration tests where a real
  filesystem or process boundary matters.
- New or changed fields update the JSON schemas and references in the
  same change.
- Failures are reported clearly, without silent fallbacks.
- Decisions taken during implementation are recorded in the owning
  document, and CI grows with the checks the phase needs.

### Phase workflow

- Before each phase, review open leftovers and agree whether to include
  them in that phase or address them in a separate commit.
- Ask detail questions and resolve material ambiguities before starting
  the phase. Record agreed interfaces in their owning references.
- Implement and verify one phase at a time. Do not start the next phase
  without its clarification and leftover review.
- At phase end, update completed tasks, document remaining loose ends here
  with their intended phase or separate-commit scope, and suggest a commit
  message. Creating a commit requires a separate instruction.
- The maintainer confirms completed commits and pushes. The first planned
  push is after phase 3.

### Planning review

These are clarification gates, not changes to the approved phase order.

- [x] Agree the phase 1 build command, site layout, frontmatter delimiters,
  minimum fields, and routing rules in the [user reference](docs/USER.md).
- [x] Publish the minimal frontmatter schema in phase 1, with executable
  schema/runtime agreement tests. Expand it with site contracts in phase 2.
- [x] Agree phase 1 minimum path and output protection, including
  symlink refusal, source/output separation, duplicate page destinations,
  and file/directory conflicts. Full asset collision handling remains in
  phase 3; ownership, replacement, and recovery remain in phase 6.
  Phase 1 also rejects case-folded and canonically equivalent destinations.
- [x] Exclude drafts in phase 1 and validate their provided known fields.
  Full metadata resolution remains in phase 2.
- [x] Document the initial prepared model and ordinary-page skeleton in the
  [theme reference](docs/THEMES.md). Complete package integration remains in
  phase 5; the skeleton is not a third selectable built-in theme.
- [x] Review phase 1 dependencies and supply required runtime notices.
  Use urfave/cli v3 for the CLI. Continue component review as dependencies
  arrive; phase 9 completes release inventory and enforcement.
- [x] Include generated-markup accessibility checks as content rendering
  arrives in phase 3. Phase 8 adds complete-page assessment for both
  shipped themes.

### Phase 1: Walking skeleton

- [x] Add root help and version output and the `cress site` and
  `cress theme` scopes [11].
- [x] Build ordinary Markdown pages with TOML frontmatter into
  directory-index routes through the embedded theme skeleton [3].
- [x] Write output only to a new or empty directory. Refuse anything else
  until phase 6 adds ownership checks and replacement.

#### Phase 1 handoff and leftovers

Phase 1 is complete. Implementation and local verification passed, and
the maintainer confirmed commit `12e8ad2`. CI execution is not required
for phase closure under the current workflow.

- [x] Before phase 2, review the current authoring and render interfaces,
  including published-page title requirements, the English-only skeleton,
  and reserved site configuration that is not yet loaded.

At the phase 1 handoff, output replacement, recovery, and writer locking
were deferred to phase 6. Markdown URL resolution and asset publication
were deferred to phase 3. Those follow-ups are recorded below.

### Phase 2: Site contracts

Agreed interface: optional `cress.toml` with `[site]`, `[site.image]`,
`[navigation]`, and `[theme.custom]`. Site fields are `title`, `description`,
`copyright`, `language`, `host`, and `path`. Page metadata keeps its required
title, inherits only description and image, and adds explicit page types.
See the [user reference](docs/USER.md) for validation and override rules.

- [x] Validate site configuration and frontmatter: errors for malformed
  input and invalid known fields, warnings for unknown fields, and an
  open theme-customization area [6].
- [x] Resolve site metadata defaults and explicit page overrides, page
  types, and draft exclusion for builds [7, 17].
- [x] Configure navigation, public origin, and deployment path, and
  resolve generator-owned URLs beneath the deployment path [3].
- [x] Publish the site-configuration schema and expand the initial
  frontmatter schema with the new contracts [6].

#### Phase 2 handoff and leftovers

Implementation and local verification are complete. The maintainer
confirmed commit `9af8877`. CI remains aligned through the existing
check/build tasks; no CI run is required.

- [x] Before phase 3, clarify source-relative Markdown and title-image
  references, asset ownership, caption syntax, highlighting options, and
  generated-markup accessibility coverage. Preserve the agreed separation
  between resolved Markdown URLs and unchanged authored HTML.
- [x] In phase 4, implement published list generation and rendering, plus
  canonical URLs, sitemap, `robots.txt`, and missing-origin warnings.
  Until then, published lists fail explicitly; excluded draft lists remain
  valid. `site.host` is validated but not used for publication metadata yet.

Title-image metadata, navigation, and page URLs now resolve the configured
URL prefix. Markdown body destinations and image asset publication still
wait for phase 3. Branding is recorded under phase 5; output replacement,
recovery, and writer locking remain phase 6 work. No separate cleanup
commit is currently required.

### Phase 3: Content rendering

Agreed interface: source-stem `.assets/` directories, shared `static/`
files, strict source-to-output reference resolution, text-and-link caption
attributes and image metadata, and default-on Chroma highlighting through
`[markdown].highlight`. The [user reference](docs/USER.md) owns syntax and
validation. Phase 2 leftovers require no separate cleanup commit.

- [x] Resolve site-root and source-relative Markdown links and images
  and leave authored HTML and external URLs unchanged [3, 19].
- [x] Render trusted raw HTML, literal code, and build-time syntax
  highlighting that can be disabled [16, 19].
- [x] Render captions for inline images and fenced code blocks, and
  prepare title-image caption data [20].
- [x] Copy page-owned assets and shared static files unchanged, keep
  private sources private, and reject output collisions with both
  sources named [4].

#### Phase 3 handoff and leftovers

Implementation and local checks/build are complete. Unit and filesystem
tests cover source resolution, draft isolation, unchanged asset bytes,
collisions, invalid sources, code escaping, and caption markup/association.
The new Chroma and regexp2 dependencies have reviewed MIT notices; the SVG
formatter and its embedded font are not linked. CI remains aligned through
the existing tasks; no CI execution was required.

Phase 3 and the local-playground follow-up are committed as `23e2e09` and
`05d0fe3`. Push confirmation remains with the maintainer. No separate
cleanup commit is currently required.

- [x] Create a temporary gitignored playground for manual validation
  before site initialization was available; retired after phase 8.
- [x] When `site init` arrives in phase 8, replace the temporary playground
  setup guidance with the supported initialization workflow.
- [x] Before phase 4, review the remaining list and publication-metadata
  work and clarify date formats, list configuration and render data,
  excerpts, and origin-dependent generated files and warnings.
- [ ] In phase 8, assess complete pages for both shipped themes against
  WCAG 2.2 AA. Current tests cover generated markup and caption association,
  not complete-page accessibility or publisher-authored HTML.

At the phase 3 handoff, output ownership, replacement, recovery, and writer
locking remained phase 6 work. That follow-up is recorded below.
List and publication-metadata implementation is tracked in phase 4 below.

### Phase 4: Lists and publication metadata

Agreed interface: optional date-only `published_at` and `updated_at`, plain
`excerpt`, and `[list]` with `order = "published_at" | "title"` and
`reverse = false`. Title ordering uses case-folded Unicode lexical keys.
Prepare entries before rendering. Generate canonicals, sitemap entries
with explicit update dates, and permissive robots output; without an
origin, warn once and omit origin-dependent metadata. The
[user reference](docs/USER.md) owns the complete authoring contract.
Phase 3 leftovers stay in phases 6 and 8.

- [x] Build list pages with section-local membership, draft visibility,
  publication and title ordering, reversal, route tie-breaks,
  missing-date warnings, and optional excerpts [21].
- [x] Generate canonical URLs, the sitemap, and `robots.txt`, and warn
  and omit origin-dependent metadata without a public origin [15].

#### Phase 4 handoff and leftovers

Implementation is complete. Coverage includes date validation on drafts,
logical list membership across equivalent source layouts, chronological
and Unicode title ordering, reversal and route ties, missing-date warnings,
prepared metadata-only entries, empty-list introductions, and escaping.
Publication coverage includes root and prefixed canonicals, XML sitemap
URLs and explicit update dates, missing-origin behavior, and generated-file
collisions that preserve previous output. No dependencies were added.

`mise run check` and `mise run build` pass locally. CI remains aligned
through the existing tasks; no CI run was required. Phase 4 is committed
as `77399e7`; push confirmation remains with the maintainer.

- [x] Before phase 5, agree brand fields and formats, theme manifest and
  runtime compatibility, local/built-in selection, required templates and
  prepared data, exact-language labels, public asset namespacing, and
  `theme init` behavior.

No separate cleanup commit was required before phase 5. Output replacement,
recovery, and locking were assigned to phase 6; the temporary site scaffold
and complete-theme accessibility assessment remain in phase 8. For subpath deployments,
host-root robots handling stays an external deployment responsibility,
as documented in the user reference.

### Phase 5: Theme integration

Agreed package and branding contracts live in the
[theme reference](docs/THEMES.md) and [user reference](docs/USER.md).
Phases 5 and 6 may follow consecutively without a commit pause; retain the
clarification and leftover review before phase 6.

- [x] Define and implement `[brand]` for brand color, logo, and favicon,
  including prepared theme data, asset URLs, schemas, and references.
  Keep it distinct from site metadata and `theme.custom`; clarify field
  formats and logo alternatives before implementation.

- [x] Define the theme package and JSON manifest, extend the prepared
  render model, and publish the manifest schema [6, 18].
- [x] Select local themes before built-in themes as complete packages,
  check runtime compatibility, and fail invalid packages without
  fallback [5].
- [x] Require ordinary, landing, and list support, load exact-language
  labels, and publish only namespaced public theme assets [18].
- [x] Add `cress theme init` from the embedded skeleton without
  overwriting files [11].

#### Phase 5 handoff and leftovers

Local checks and build pass. Brand fields, manifests, runtime compatibility,
exact-language labels, asset namespaces, and non-overwriting initialization
are implemented with schema/runtime and filesystem coverage. Go modules'
semantic-version helper has reviewed BSD-3-Clause notices.

Phases 5 and 6 are committed together as `5dd386c`.
Push confirmation remains with the maintainer.

At the phase 5 handoff, the two named built-in themes remained phase 8 work;
selected names
resolve locally or fail explicitly. Unconfigured builds retain the embedded
development skeleton. No separate cleanup commit was needed before phase 6.
Site initialization, temporary-playground retirement, and complete-theme
accessibility assessment were assigned to phase 8.

### Phase 6: Output safety

Clarified before implementation: private gitignored `.cress/` state, paired
public ownership marker, no adoption of old nonempty output, verified
automatic recovery with an explicit stop on ambiguity, and a fail-fast OS
lock covering the entire build. No phase 5 cleanup is required first.

- [x] Stage complete output, verify Cress ownership, and replace previous
  output as a whole, removing stale files [10].
- [x] Coordinate writers per site with OS advisory locking and fail fast
  on contention [10].
- [x] Restore previous output on failed publication, preserve recovery
  evidence, and stop with guidance in ambiguous states [10].

#### Phase 6 handoff and leftovers

Complete staged replacement, paired ownership, whole-build OS locking, and
verified recovery are implemented. Coverage exercises stale/manual file
removal, source and staging errors, interrupted transactions, injected
promotion/restoration/cleanup/state-record failures, unsafe targets, path
aliases, cross-process contention, and lock release on process exit.
Unit-only mode, race checks, and both Linux cross-builds pass alongside
the repository checks/build. No dependency was needed for output safety.
CI continues to use the same local tasks; no CI run was required.

Phases 5 and 6 are committed together as `5dd386c`
(`feat: add local themes and safe output publication`).
Push confirmation remains with the maintainer.

- [x] Retire the old playground entirely after site initialization arrives;
  no pre-ownership output migration remains necessary for that fixture.
- [x] Before phase 7, clarify preview CLI/draft choices, session ownership,
  watch/retry scheduling, and cleanup. Extend the writer with explicit
  session targets and verified-usable-output reporting for preview errors.
- [x] In phase 8, include private `.cress/` state in initialized sites'
  Git ignore rules.

No separate cleanup commit is required. Preview/session behavior remains
phase 7 work. Built-in themes, site initialization, playground retirement,
and complete-theme accessibility assessment remain in phase 8.
Ambiguous recovery still requires manual inspection; no repair/adoption
command or power-loss durability guarantee is introduced.

### Phase 7: Preview

Clarified before implementation: loopback-only
`cress site preview [directory] --drafts --port N`, default port 8080,
explicit free-port selection with `--port 0`, unchanged public origin and
deployment prefix, native fsnotify events with a short debounce, and
private temporary sessions with lock-verified reclamation. Included drafts
require publishable metadata. No cleanup commit is required first.
The unowned playground output was preserved during phase 7 and later retired.

- [x] Serve isolated private preview sessions with an explicit draft
  choice and no effect on deployable output [7].
- [x] Rebuild on source edits without losing edits made during a rebuild,
  keep the last verified output on failure, and retry blocked rebuilds
  [8].
- [x] Stop unsuccessfully when no verified output remains, clean up on
  shutdown, and reclaim only verifiably inactive sessions [8, 9].

#### Phase 7 handoff and leftovers

Preview implementation is complete. The writer now separates the shared
site lock from target recovery, so previews never require or mutate
deployable output. Included drafts use normal metadata, rendering, source
resolution, assets, lists, and publication preparation.

Coverage includes concurrent draft-free and draft-inclusive sessions,
prefix serving and updates, invalid-edit preservation, edits during
rebuilding, contention retries without another edit, nested source edits,
atomic configuration saves, generated-output event exclusion, publication
and cleanup failure usability, fatal unverified output, inactive-session
reclamation, changed identities, and signal shutdown. Local checks/build,
unit-only mode, race checks, and supported Linux cross-builds pass.
CI remains aligned through the existing tasks; no CI execution was required.

New runtime dependencies are fsnotify and its Go system bindings, with
reviewed BSD-3-Clause notices and the Go patent grant.
Phase 7 is committed as `86aaf05`; push confirmation remains with the
maintainer.

- [x] Before phase 8, clarify the two built-in themes' names, presentation
  contracts and defaults, site-init CLI and scaffold contents, and the
  complete-page accessibility review scope. Revisit existing playground
  and site-ignore leftovers first; no separate cleanup commit is required.

At the phase 7 handoff, built-in themes, site initialization, playground
retirement, generated site ignore rules, and complete-theme accessibility
assessment remained phase 8 work. The temporary playground has since been
retired in favor of initialized test sites. Native-watch filesystem limitations
and manual ambiguous-recovery handling are documented in the user reference.

### Phase 8: Built-in themes and site initialization

Clarified: `poppy` (freelancer and default) and `clover` (small business),
system fonts and local CSS only, exact English/German labels, no initial
custom keys, non-overwriting site init into a new/empty directory, and an
English sample scaffold. Human accessibility assessment remains an explicit
release blocker; no automated-browser scope or AA claim is added.

- [x] Build the coach/freelancer and corporate small-business themes
  with all page types, captions, and navigation [1, 2, 20].
- [x] Add `cress site init` for a ready-to-build site without overwriting
  files [1, 11].
- [ ] Assess default pages against WCAG 2.2 AA with automated checks and
  manual review, and record scope, environments, results, and
  limitations [12].

#### Phase 8 handoff and leftovers

Embedded themes and site initialization are implemented. Both themes build
the seven-page scaffold offline; local override and invalid-local failure,
unsafe-target refusal, output preservation, complete-page structure, and
default palette contrast have automated coverage. No dependencies or
third-party assets were added.

The generated `cress.toml` documents all supported settings in place,
including defaults and commented optional examples. Both theme selections
retain ready-to-build defaults; optional examples are parsing-tested.
The footer-linked content tour owns the main frontmatter hints; other pages
document only distinct features, such as list settings. The tour demonstrates
CommonMark, tables, captions, title images, and authored HTML.
Both themes have build coverage for the tour. No new leftovers were added.

The remaining accessibility task is intentionally unchecked:
[the assessment record](docs/ACCESSIBILITY.md) records automated evidence
and a complete A/AA human-review matrix without claiming hands-on results.
Complete that review before release. The old playground is removed entirely;
developers use initialized sites under ignored `/tmp/`. Developer, theme-author,
and publisher guides document fresh test-site initialization.

Suggested commit message: `feat: add poppy and clover themes and site init`.
No commit or push was performed. Phase 9 follows after its clarification
gate, including review of every repository Markdown file and honest
documentation of external release blockers.

### Phase 9: Documentation and release

Clarified: `v*` tag-triggered GitHub Actions with strict full SemVer,
checks before publication, the three agreed targets, explicit version
injection, and keyless Cosign checksum signing. Pin external Actions by
reviewed commit. The maintainer clarified that the requested inventory is
machine-readable output only: generate CSV reports with go-licenses.
No added license checks, CycloneDX tooling, custom release executable,
or checked-in generated inventory. Generate notice text with the same tool
and ship a short install/verify reference. Audit every
repository Markdown document, including ADRs and scaffold content.
No tag, publication, commit, or push is authorized by this phase.

- [x] Complete author, maintainer, and theme-development references,
  including schema-based validation guidance [6, 11].
- [x] Generate release-specific CSV dependency-license reports and archive
  notice text with go-licenses [14].
- [x] Prepare version-tag GitHub Actions and GoReleaser automation,
  with Cosign keyless signing and documented verification [13].
- [ ] Publish an explicitly approved release and verify its real signature
  bundle, expected identity/issuer, archive and report digests, and rejection
  of tampered files and an unexpected identity [13, 14].
- [ ] Confirm all acceptance criteria on the supported targets.

#### Phase 9.5: Markdown features and a third built-in theme

Agreed scope: strikethrough, linkable Markdown headings, static task lists,
and one paper/Scandi theme alongside poppy and clover. Three bundled themes
are enough for this scope. Heading targets are approved under
[ADR 0017](docs/adr/0017-markdown-heading-targets.md). Other interfaces and
the visual design remain to be agreed before implementation.

- [ ] Before implementation, update the owning product requirements and
  decision records for the agreed scope, including the third bundled
  theme and its acceptance coverage.
- [x] Agree heading-ID behavior: automatic IDs, explicit author overrides,
  non-ASCII text, duplicate headings, and collisions with theme and
  generated IDs. Decide separately whether visible permalink controls
  are needed; linkable headings do not require them.
- [x] Record heading-link requirements, authoring rules, and theme ID
  responsibilities in their owning documents before implementation.
- [x] Implement heading targets with explicit-ID priority and reservations;
  cover both heading syntaxes, source-aware errors, links, captions, and
  unchanged raw HTML. Demonstrate and build them with both bundled themes [22].
- [ ] Agree static task-list markup and accessibility behavior. Checkboxes
  show authored completion state; they are not interactive tasks.
- [ ] Add strikethrough and static task lists using Goldmark's individual
  built-in capabilities, not the complete GFM bundle. Cover the agreed
  behavior and existing content interactions.
- [ ] Agree the paper/Scandi theme's presentation contract before building.
  Aim for a playful, editorial alternative for makers, independent shops,
  cafes, and creative businesses, distinct from poppy rather than another
  off-white palette. Explore framed images, restrained rotation of image
  frames, and subtle shadows without rotating text. Keep pages useful
  without photographs.
- [ ] Build only the selected third theme with all required page types,
  content features, navigation, and labels. Update initialization and
  references, and assess complete default pages against the existing
  accessibility release baseline.

Deferred, not implementation commitments:

- The glass-like theme may become an externally maintained theme later.
  It is not a fourth bundled theme or a promised external deliverable.
- Automatic table-of-contents generation, including a per-page frontmatter
  option, needs a demonstrated long-page navigation use case before approval.
  Goldmark supports this through an external extension, not a built-in one;
  review dependency suitability, heading depth, placement, labels, schemas,
  and theme presentation if revisited.
- Other Markdown extensions remain suggestions. Definition lists merit a
  use-case review for services and menu content; do not add them, footnotes,
  automatic typography substitutions, or other extensions by default.

#### Phase 9 handoff and leftovers

Release preparation and the documentation audit are implemented, not a
claim that 1.0 is ready or published. The
[release reference](docs/RELEASE.md) owns the concrete maintainer procedure;
the [install reference](docs/INSTALL.md) ships in every archive.

The Markdown audit covers root guidance/README/work plan, all guides,
all 16 ADRs, notices, the accessibility record, and all seven scaffold pages.
Accepted ADR rationale remains unchanged. Current references now describe
site initialization, poppy/clover and local-first selection, exact labels,
explicit draft-inclusive preview, current schemas, and release tooling
instead of the old walking-skeleton status.

GoReleaser invokes go-licenses report for CSV output and generated notice
text. The custom helper, generated approval inventory, license policy
checks, scanner exceptions, and CycloneDX tooling are removed.
The third-party reference points to generated artifacts and go.mod rather
than duplicating a dependency/license catalogue. ADRs remain untouched.

The workflow's strict SemVer guard accepts valid versions and rejects
shortened, malformed, and leading-zero versions.
Tags use `v1.0.0` or `v1.0.0-pre.0`, never `v-1.0.0-pre.0`.
GoReleaser automatically marks suffix-bearing versions as GitHub prereleases.

Local Markdown links and internal anchors are checked; scaffold Markdown
is reviewed and built as content, not linted as a guide.
Setup documentation points to mise.toml and go.mod rather than maintaining
duplicate tool versions or dependency inventories. Compatibility baselines
and release-tag examples remain where they explain required behavior.

- [x] Review the software design audit (local `audits/software-design-2026-10-25.md`)
  and fix #2 before #1: preserve preview root ownership on replacement failure,
  then separate the site lock from target-bound writers without changing
  recovery behavior. Committed together as `35a9e6c`
  (`fix: clarify preview and writer resource ownership`).
- [ ] Revisit audit #3 (private theme-file reads) and #4 (unchecked scaffold
  substitutions) separately; neither is an automatic release blocker.
- [ ] Review the code quality audit (local `audits/code-quality-2026-10-26.md`)
  before approving output-file organization (#5, optional, not approved).
  Findings #1 to #4 (preview, Markdown, recovery, content loading) are done.
  Findings are proposals, not new release blockers.
- [x] Fix the two P2 findings from the local theme design audit
  (`audits/theme-design-2026-10-26.md`): lazy-load list-entry images and
  style focusable Markdown table wrappers for horizontal scrolling.
  Keep page title images eager and leave authored HTML unchanged.
- [x] Apply the theme P3 refinements: a thinner inset main focus ring,
  shared outer measures for header/content/footer, syntax color tokens,
  and inline-code styling kept separate from fenced code.
  Image dimensions are not modeled; layout-shift prevention remains
  outside this change. Firefox screenshots cover 320px and 1280px layouts;
  cross-browser keyboard/touch and screen-reader review remain pending
  in the human accessibility matrix.

- [x] Follow up on the local landing-page design critique
  (`audits/landing-design-critique-2026-10-27.md`): demonstrate a
  Services-first, offer-first scaffold journey and explore distinct
  Poppy/Clover landing compositions using existing prepared data.
  Done: Services-first scaffold journey, distinct Poppy/Clover landing
  compositions, consistent site frame, brand accent, and title-image band.
- [x] Prepare persona content fixtures under the gitignored
  `tmp/theme-exploration/`: coach with Poppy and cleaning with Clover.
  Both demonstrate an offer-first, Services-first journey and secondary
  contact route. These are fictional design examples, not shipped scaffold
  replacements or working booking, purchase, or newsletter services.
- [x] Choose distinct landing compositions and check consultant and
  restaurant fixtures. Poppy: calm single column, sans headings throughout,
  outlined primary link. Clover: solid title block, heavy section rules,
  filled primary link. The primary link is a CSS convention (a lone-link
  paragraph after the first paragraph); no data fields or link classes.
  The scaffold now leads with a Services link. Firefox covers 320/1280px.
- [x] Keep header, navigation, and footer at one frame width on every page.
  The frame is sized to the reading measure (Poppy 52rem, Clover 56rem
  with 1.0625rem body text), so content fills it and shares both edges.
  Clover's header and footer drop their borders; the surface tint separates.
- [x] Use `[brand] color` as the built-in themes' accent (links, buttons,
  Clover's title block). Cress warns below 4.5:1 against white and keeps
  building; Clover's header/footer links use the text color. ADR 0007,
  USER.md, and ACCESSIBILITY.md record the change. The scaffold seedling is
  replaced by a neutral, labelled photo placeholder.
- [x] Crop landing title images to a 5:2 band so the introduction and
  primary link stay near the first screen. Clover renders the image below
  its title block instead of inside it.
- [x] Replace Poppy's cream background with a chalk off-white
  (`#fbfbf9`, surface `#f0f0ec`).
- [ ] Loose threads: Poppy's off-white background lowers brand contrast slightly
  (a 4.5:1 color on white is about 4.34:1 on it);
  the 5:2 landing crop is checked only with the placeholder, not real
  photos; a `[site.image]` default also fills every landing hero; CSS ignores
  text nodes, so a paragraph with text around a single link right after the
  first paragraph is also styled as the primary link. Keyboard and
  screen-reader review belong to the accessibility matrix.

External release blockers remain:

- [ ] Complete the human accessibility matrix for both themes and all
  default page types; fix and retest every applicable A/AA failure.
- [ ] Record native Linux amd64 and Linux arm64 acceptance results.
  Cross-builds and local macOS execution do not prove native Linux behavior.
- [ ] Complete the supported-target/product acceptance review and record
  results before approving a release tag.
- [ ] After an explicitly approved first release, record real keyless
  signing and negative verification evidence; snapshots cannot provide it.
- [x] Maintainer confirms phases 8/9 and scaffold updates committed as
  `567f6b6` (`feat: add themes, init and docs`).
- [x] Audit resource-ownership fixes are committed as `35a9e6c`.
- [ ] Maintainer confirms any push. No repository write to GitHub was
  performed during the audit follow-up.

Suggested phase 9 commit: `build: prepare signed releases and refresh docs`.
For a combined phase 8/9 commit:
`feat: add bundled themes, site init, and release preparation`.

Ideas beyond 1.0 are welcome as separate proposals.
