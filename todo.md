# Work plan

This is Cress's implementation plan and backlog, not another source of
product requirements or architectural decisions.
Use the [PRD](docs/prd.md) for product scope and the [ADRs](docs/adr/) for
architectural guidance and guardrails.

Phase 0 ends with an approved implementation plan that creates phases 1
through n here. Resolve implementation details when their work is planned
and implemented, not through speculative prerequisites or proof exercises.
Backlog items are real work, not approved designs or a fixed sequence.

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
- [ ] Document the agreed architecture without prescribing speculative
  package layouts or interfaces.
- [ ] Finish phase 0 by approving a phased implementation plan against the
  PRD's acceptance criteria and create phases 1 through n in this file.

## Backlog

These workstreams inform the eventual implementation plan. They do not
require complete specifications or feasibility demonstrations in advance.

- [ ] Add CI using the existing developer tasks. Grow linting, unit tests,
  useful integration coverage, and accessibility checks as features land.
- [ ] Implement authoring, configuration, metadata resolution, routing,
  Markdown rendering, captions, lists, and asset handling.
- [ ] Implement the theme contract, both built-in themes, and theme
  initialization.
- [ ] Implement site initialization, generation, preview, output safety,
  and recovery.
- [ ] Publish schemas and author, maintainer, and theme-development
  documentation, including schema-based validation guidance.
- [ ] Implement release automation, component/license review, and
  machine-readable SBOM publication.
- [ ] Review default pages for accessibility and record scope, results, and
  limitations. Complete compatibility and target-platform release checks
  against the PRD.

Features outside 1.0 are not scheduled here. Consider them through separate
proposals when there is a concrete need.
