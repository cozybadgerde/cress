# Working on Cress

- Read [the developer guide](docs/DEVELOPER.md) for setup,
  [the PRD](docs/PRD.md) for scope,
  [the ADRs](docs/PRD.md#decision-records) for decisions,
  [the architecture document](docs/ARCHITECTURE.md) for structure, and
  [GitHub issues](https://github.com/cozybadgerde/cress/issues) for pending
  work. Agreed behavior is not necessarily implemented.
- Track work in GitHub issues using `gh`, not a local work plan. Work through
  one task at a time. Agree on material interfaces and visual choices before
  implementation. Commits, pushes, tags, and publication require maintainer
  approval.
- Follow the [development guidelines](docs/DEVELOPER.md#development-guidelines):
  interface first, standard library first, established libraries over custom
  code, and comments of at most two lines.
- Use these skills when available: `context-cozy-badger` for company context,
  `convention-md` for Markdown, `convention-go` for Go, and `impeccable` for
  user interface design.
- Keep tooling and code small. Add packages, interfaces, dependencies, and
  configuration only for a current need.
- Cress generates offline; deployment stays external. Themes present
  prepared data, not content-selection or routing rules.
- Preserve output ownership, recovery, draft isolation, and the documented
  HTML trust boundary. Do not hide failures or introduce silent fallbacks.
- Use `mise run check` and `mise run build` for changes. Use
  `mise run check:goreleaser` for release configuration. Keep checks read-only;
  formatting is `mise run fmt`.
- Prefer unit tests; add integration coverage where useful. Automated
  browser/end-to-end tests are outside the initial scope.
- Accessibility means following best practices, not a legal audit or WCAG
  conformance claim. Don't add audit matrices, per-criterion evidence, or
  approval gates beyond the release checklist.
- Write `docs/` for humans: short guidance without disclaimers or meta
  statements about the document itself. Title each file "X guide" (walks
  through or explains) or "X document" (details to look up), never mixed.
- Keep product requirements in the PRD, decision rationale in ADRs, and
  implementation tasks in GitHub issues. Avoid duplicating those policies.
