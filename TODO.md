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
- [ ] Include generated-markup accessibility checks as content rendering
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

Output replacement, recovery, and writer locking remain intentional
phase 6 work. Until then, builds refuse existing nonempty output; a
filesystem failure can leave incomplete new output requiring inspection.
Do not run concurrent builds for one site. Markdown URL resolution and
asset publication remain phase 3 work, as documented in the user reference.

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

Implementation and local verification are complete. CI remains aligned
through the existing check/build tasks; no CI run is required. No commit
or push was performed for phase 2.

- [ ] Before phase 3, clarify source-relative Markdown and title-image
  references, asset ownership, caption syntax, highlighting options, and
  generated-markup accessibility coverage. Preserve the agreed separation
  between resolved Markdown URLs and unchanged authored HTML.
- [ ] In phase 4, implement published list generation and rendering, plus
  canonical URLs, sitemap, `robots.txt`, and missing-origin warnings.
  Until then, published lists fail explicitly; excluded draft lists remain
  valid. `site.host` is validated but not used for publication metadata yet.

Title-image metadata, navigation, and page URLs now resolve the configured
URL prefix. Markdown body destinations and image asset publication still
wait for phase 3. Branding is recorded under phase 5; output replacement,
recovery, and writer locking remain phase 6 work. No separate cleanup
commit is currently required.

### Phase 3: Content rendering

- [ ] Resolve site-root and source-relative Markdown links and images
  and leave authored HTML and external URLs unchanged [3, 19].
- [ ] Render trusted raw HTML, literal code, and build-time syntax
  highlighting that can be disabled [16, 19].
- [ ] Render captions for inline images and fenced code blocks, and
  prepare title-image caption data [20].
- [ ] Copy page-owned assets and shared static files unchanged, keep
  private sources private, and reject output collisions with both
  sources named [4].

### Phase 4: Lists and publication metadata

- [ ] Build list pages with section-local membership, draft visibility,
  publication and title ordering, reversal, route tie-breaks,
  missing-date warnings, and optional excerpts [21].
- [ ] Generate canonical URLs, the sitemap, and `robots.txt`, and warn
  and omit origin-dependent metadata without a public origin [15].

### Phase 5: Theme integration

- [ ] Define and implement `[brand]` for brand color, logo, and favicon,
  including prepared theme data, asset URLs, schemas, and references.
  Keep it distinct from site metadata and `theme.custom`; clarify field
  formats and logo alternatives before implementation.

- [ ] Define the theme package and JSON manifest, extend the prepared
  render model, and publish the manifest schema [6, 18].
- [ ] Select local themes before built-in themes as complete packages,
  check runtime compatibility, and fail invalid packages without
  fallback [5].
- [ ] Require ordinary, landing, and list support, load exact-language
  labels, and publish only namespaced public theme assets [18].
- [ ] Add `cress theme init` from the embedded skeleton without
  overwriting files [11].

### Phase 6: Output safety

- [ ] Stage complete output, verify Cress ownership, and replace previous
  output as a whole, removing stale files [10].
- [ ] Coordinate writers per site with OS advisory locking and fail fast
  on contention [10].
- [ ] Restore previous output on failed publication, preserve recovery
  evidence, and stop with guidance in ambiguous states [10].

### Phase 7: Preview

- [ ] Serve isolated private preview sessions with an explicit draft
  choice and no effect on deployable output [7].
- [ ] Rebuild on source edits without losing edits made during a rebuild,
  keep the last verified output on failure, and retry blocked rebuilds
  [8].
- [ ] Stop unsuccessfully when no verified output remains, clean up on
  shutdown, and reclaim only verifiably inactive sessions [8, 9].

### Phase 8: Built-in themes and site initialization

- [ ] Build the coach/freelancer and corporate small-business themes
  with all page types, captions, and navigation [1, 2, 20].
- [ ] Add `cress site init` for a ready-to-build site without overwriting
  files [1, 11].
- [ ] Assess default pages against WCAG 2.2 AA with automated checks and
  manual review, and record scope, environments, results, and
  limitations [12].

### Phase 9: Documentation and release

- [ ] Complete author, maintainer, and theme-development references,
  including schema-based validation guidance [6, 11].
- [ ] Review all components and licenses, ship required notices, and
  publish release SBOMs [14].
- [ ] Release from version tags through GitHub Actions and GoReleaser,
  with Cosign keyless signing and documented verification [13].
- [ ] Confirm all acceptance criteria on the supported targets.

Ideas beyond 1.0 are welcome as separate proposals.
