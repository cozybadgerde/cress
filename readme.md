# Cress

Cress is a Go static site generator for small-business websites by
Cozy Badger. It focuses on predictable Markdown authoring, TOML
configuration, and curated themes.

Cress generates static files locally. Deployment is outside its scope.

## Project status

The project is in phase 0: requirements, architectural guidance,
foundations, and implementation planning. A minimal Go executable and
developer tooling are in place. The executable reports that Cress is not
implemented and exits unsuccessfully; no generator or release pipeline
exists yet.

Product owner: Daniel Schier (Cozy Badger).

## Development

Install [mise](https://mise.jdx.dev/getting-started.html), then run from the
repository root:

```sh
mise install
mise run check
mise run build
```

Review `mise.toml` before trusting it if mise requests approval.
Tool versions are pinned there; Go dependencies belong in `go.mod` and
`go.sum` when dependencies are added. Tool installation needs network
access; it is separate from Cress's offline-generation requirement.

| Command | Purpose |
|---------|---------|
| `mise run check` | Check Go formatting/lint, Markdown, spelling, TOML, and tests without editing files. |
| `mise run fmt` | Format Go and TOML files. |
| `mise run test` | Run `go test ./...`. No tests exist in the initial scaffold. |
| `mise run build` | Build `bin/cress`. |
| `mise run check:goreleaser` | Validate GoReleaser configuration without building or publishing. |

`check` and `fmt` group independent tasks that may run in parallel.
Run checks individually with `mise run check:go-format`, `check:go-lint`,
`check:markdown`, `check:spelling`, `check:toml-format`, or `check:toml`.
Formatting tasks are `mise run fmt:go` and `mise run fmt:toml`.
Use `mise tasks ls` to list all tasks and their descriptions.

The tasks call ordinary tools; native Go commands remain usable directly.
Go linting uses golangci-lint's standard set. Markdown checks use
gomarklint without external HTTP checks. Its list-spacing rule is disabled
because it misreads wrapped list items; typos checks spelling.
Taplo formats and checks TOML. Schema associations will be added when
authoring contracts exist; frontmatter extraction remains separate work.
EditorConfig supplies basic editing defaults.

The GoReleaser configuration is a packaging scaffold for the three agreed
targets, not a release-ready pipeline. Signing, component inventories,
third-party notices, and release automation remain in the work plan.

## Project documents

- [Product requirements](docs/prd.md): scope, principles, and acceptance
  criteria for 1.0.
- [Architectural decision records](docs/adr/): accepted decisions and their
  rationale, listed in the [decision index](docs/prd.md#decision-records).
- [Work plan](todo.md): phase-0 work, backlog, and the eventual implementation
  phases.

The PRD defines product scope; accepted ADRs provide architectural guidance
and guardrails, not a detailed implementation plan.
The work plan tracks delivery without adding product commitments.

## License

Cress is licensed under [BSD-3-Clause](LICENSE).
No third-party code or assets are bundled yet. Add the required notices
when introducing bundled components.
See [the licensing decision](docs/adr/0011-bsd-3-clause-license.md).
