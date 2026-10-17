# Working on Cress

- Read [readme.md](readme.md) for setup, [the PRD](docs/prd.md) for scope,
  [the ADRs](docs/prd.md#decision-records) for decisions, and
  [todo.md](todo.md) for pending work. Agreed behavior is not necessarily
  implemented.
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
- Keep product requirements in the PRD, decision rationale in ADRs, and
  implementation tasks in the work plan. Avoid duplicating those policies.
