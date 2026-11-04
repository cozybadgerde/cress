# Work plan

This plan tracks open work only. Product scope lives in the
[PRD](docs/PRD.md), rationale in the [ADRs](docs/adr/), and structure in the
[architecture guide](docs/ARCHITECTURE.md). Completed work is retained in
Git history, not repeated here.

## Phase workflow

Agree material interfaces and visual choices before implementation. Work
through one task at a time, reviewing relevant leftovers before starting.
New proposals below are not permission to expand scope or publish a release.

For code changes, run `mise run check` and `mise run build`; release
configuration changes also require `mise run check:goreleaser`. Use focused
unit tests and useful integration coverage, without adding a browser suite.
Update affected schemas and owning references alongside behavior changes.
Remove completed tasks and keep remaining work here. Commits, pushes, tags,
and publication require maintainer approval.

## Phase 1: Theme composition and identity

Evaluate consistency within each theme against its
[presentation contract](docs/USER.md#built-in-themes), not by making the themes
uniform. Preserve Poppy's freelancer focus, Clover's established small-business
presentation, and Anemone's maker/shop/cafe editorial character.

- [ ] Tighten mobile header/navigation composition within each theme.
  Review wrapping, identity width, and space before headings with short/long
  site titles and English/German labels at 320px and intermediate widths.
  Preserve clear current-page state, keyboard focus, usable touch targets,
  and each theme's character. Do not hide navigation or shrink targets solely
  to keep links on one row.
- [ ] Refine caption hierarchy within each theme.
  Compare title-, body-, and list-image captions with that theme's metadata,
  prose, and links. Tune secondary tone and attachment spacing while retaining
  readable contrast, recognizable links, and upright captions. Sizes, colors,
  and spacing may differ between themes.
- [ ] Discuss intrinsic image dimensions before changing the render model.
  Determine whether reserving local-image space warrants prepared width/height
  data and how SVG, remote, and unknown-size images behave. Keep this separate
  from crop choices; do not fetch remote images or expand the model without
  an agreed current need.

## Phase 2: Project branding and documentation

Schedule this as a dedicated session after theme polish, identity review,
and real-photo checks.

- [ ] Establish Cress's project brand using maintainer-prepared images.
  Review asset rights and variants; agree placement in documentation and other
  project-facing surfaces before implementation. Keep project branding separate
  from publisher identity and generic scaffold content.
- [ ] Review project documentation together with the branding work.
  Check clarity, navigation, examples, and agreement with current behavior.
  Keep requirements, rationale, and implementation tasks in their owning
  documents rather than duplicating policies.

## Phase 3: Release tooling and acceptance

The [release reference](docs/RELEASE.md) owns the maintainer procedure.
These tasks are not authorization to tag or publish.

- [ ] Prepare fresh pre-publication evidence.
  Retain existing checks/build and run the functional suite uncached against
  the intended release revision.
- [ ] Add an archive smoke check after packaging and before publication.
  Confirm exactly macOS arm64, Linux amd64, and Linux arm64; required notices
  and install guidance; nonempty per-target license CSVs; and checksum coverage
  for every archive/report. Run the host-compatible archived binary to verify
  its version and build initialized sites with all three embedded themes.
  Use disposable directories and preserve existing output on failures.
- [ ] Plan post-publication Cosign verification.
  Require the exact tag identity/issuer, verify archive/report digests, and
  reject modified checksums, modified artifacts, and wrong identities.
  Keep this distinct from pre-publication gating.
- [ ] Complete the human accessibility matrix for all three themes and
  default page types. Record environments, scope, results, and limitations in
  [the assessment](docs/ACCESSIBILITY.md); fix and retest applicable A/AA
  failures. Include keyboard, screen-reader, touch, zoom/reflow, and
  table-wrapper behavior with and without overflow. Exploratory checks do not
  complete this gate.
- [ ] Complete supported-target/product acceptance against the
  [PRD criteria](docs/PRD.md#acceptance-criteria-for-10).
  Record criterion-specific environments and results for the release revision.
  Existing automated and native-platform checks do not replace the remaining
  manual/product assessment.
- [ ] Review accessibility outcomes, platform evidence, and distribution
  rights before approving a release tag.

## Phase 4: Approved publication and verification

Depends on release readiness and explicit maintainer approval.

- [ ] Publish the approved release through the documented tag workflow.
- [ ] Record real keyless-signature and negative-verification evidence for
  the published artifacts. Snapshots cannot provide this evidence.

## Optional follow-up: Shell completion

- [ ] Evaluate offline shell-completion generation.
  Check the existing CLI library and agree supported shells, command syntax,
  output behavior, and installation guidance before implementation.
  Keep shell setup explicit.

## Deferred ideas

These are not implementation commitments:

- An externally maintained glass-like theme, not a fourth bundled theme.
- Automatic table of contents, including per-page control, after a demonstrated
  long-page navigation need. Review dependency suitability, heading depth,
  placement, labels, schemas, and presentation before approval.
- Additional Markdown extensions such as definition lists for services/menu
  content, footnotes, or typography substitutions, after a concrete use-case
  review.
