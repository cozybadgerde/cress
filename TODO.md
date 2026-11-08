# Work plan

This plan tracks open work only. Product scope lives in the
[PRD](docs/PRD.md), rationale in the [ADRs](docs/adr/), and structure in the
[architecture document](docs/ARCHITECTURE.md). Completed work is retained in
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

## Phase 3: Release tooling and acceptance

The [release guide](docs/RELEASE.md) owns the maintainer procedure.
These tasks are not authorization to tag or publish.

- [ ] Agree and implement a pre-1.0 update check.
  Help installed users discover newer releases. Decide the CLI interface
  and whether checking is explicit or part of version output, including
  network behavior, version selection, and failure handling before
  implementation. Keep site generation offline.
- [ ] Prepare fresh pre-publication evidence.
  Retain existing checks/build and run the functional suite uncached against
  the intended release revision.
- [ ] Verify the draft publication gate on GitHub.
  Local snapshots do not exercise uploads or draft publication. Confirm
  successful publication preserves assets and the stable/prerelease flag.
  Confirm failed checks leave drafts unpublished, and establish a safe
  retry procedure for failed drafts before relying on workflow reruns.
  A 2026-11-06 failure rehearsal and retry created two unpublished drafts
  for the same tag with different archives. The updated gate published a
  prerelease with matching artifact digests and refused retries while a
  release or failed draft existed. Verify a stable publication and a fresh
  retry after deliberate draft cleanup before closing this task.
- [ ] Plan post-publication Cosign verification.
  Require the exact tag identity/issuer, verify archive/report digests, and
  reject modified checksums, modified artifacts, and wrong identities.
  Keep this distinct from pre-publication gating.
- [ ] Review release readiness with the
  [pre-tag checklist](docs/RELEASE.md#2-go-through-the-pre-tag-checklist).
  Cover accessibility across the three themes, supported targets, product
  behavior, and distribution rights. Keep brief notes and retest fixes.

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
