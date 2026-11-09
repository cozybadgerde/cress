# Work plan

This plan tracks open work only. Product scope lives in the
[PRD](docs/PRD.md), rationale in the [ADRs](docs/adr/), and structure in the
[architecture document](docs/ARCHITECTURE.md). Completed work is retained in
Git history, not repeated here.

## Phase workflow

Agree material interfaces and visual choices before implementation. Work
through one task at a time, reviewing relevant leftovers before starting.

For code changes, run `mise run check` and `mise run build`; release
configuration changes also require `mise run check:goreleaser`. Use focused
unit tests and useful integration coverage, without adding a browser suite.
Update affected schemas and owning references alongside behavior changes.
Remove completed tasks and keep remaining work here. Commits, pushes, tags,
and publication require maintainer approval.

## Road to 1.0.0

Follow the [release guide](docs/RELEASE.md) for each release. Tags and
publication need maintainer approval. Ideas for later releases live in
[GitHub issues](https://github.com/cozybadgerde/cress/issues).

- [ ] Add feeds for list pages.
  Publish a web feed for each list page, so readers can follow a journal.
- [ ] Release `v1.0.0-pre.0`.
  Go through the full release guide. Then confirm the draft gate published
  it as a prerelease with all assets, and that the signature passes and all
  three negative checks fail. If a draft fails, test the cleanup and retry
  from step 5.
- [ ] Release `v1.0.0`.
  Rerun the automated checks and repeat only the checklist rows touched since
  the prerelease. Confirm it is published as the latest stable release and
  its signature verifies.
