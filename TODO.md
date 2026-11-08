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

## Road to 1.0.0

Follow the [release guide](docs/RELEASE.md) for each release. Tags and
publication need maintainer approval.

- [ ] Add copy-paste commands for the negative signature checks to
  [step 6 of the release guide](docs/RELEASE.md#6-verify-the-published-release):
  a modified checksum file, a modified archive, and a wrong tag identity.
- [ ] Release `v1.0.0-pre.0`.
  Go through the full release guide. Then confirm the draft gate published
  it as a prerelease with all assets, and that the signature passes and all
  three negative checks fail. If a draft fails, test the cleanup and retry
  from step 5.
- [ ] Release `v1.0.0`.
  Rerun the automated checks and repeat only the checklist rows touched since
  the prerelease. Confirm it is published as the latest stable release and
  its signature verifies.

## Optional follow-up: Shell completion

- [ ] Evaluate offline shell-completion generation.
  Check the existing CLI library and agree supported shells, command syntax,
  output behavior, and installation guidance before implementation.
  Keep shell setup explicit.

## Deferred ideas

These are not implementation commitments:

- An update check in a minor release after 1.0.0, tracked in
  [issue #2](https://github.com/cozybadgerde/cress/issues/2).
- An externally maintained glass-like theme, not a fourth bundled theme.
- Automatic table of contents, including per-page control, after a demonstrated
  long-page navigation need. Review dependency suitability, heading depth,
  placement, labels, schemas, and presentation before approval.
- Additional Markdown extensions such as definition lists for services/menu
  content, footnotes, or typography substitutions, after a concrete use-case
  review.
