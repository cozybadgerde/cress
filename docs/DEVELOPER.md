# Developer guide

This guide explains how to set up a local development environment, change the
source, and build and test Cress.

## Before you start

These documents explain what Cress should do and why:

- [Product requirements](PRD.md): scope and acceptance criteria for 1.0.
- [Architecture guide](ARCHITECTURE.md): responsibilities, data flow, and
  failure boundaries.
- [Architectural decision records](PRD.md#decision-records): accepted
  decisions and their rationale.
- [Work plan](../TODO.md): implementation phases and their definition of
  done.

Much of the agreed behavior is not implemented yet. The
[contribution notes](../README.md#contributing) explain how to propose a
change.

## Prerequisites

- Linux or macOS.
- [Git](https://git-scm.com/).
- [mise](https://mise.jdx.dev/getting-started.html), which installs every
  other development tool.

mise installs these pinned versions from `mise.toml`:

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.27.2 | Build and test. |
| golangci-lint | 2.15.0 | Lint Go code. |
| gomarklint | 3.3.1 | Lint Markdown. |
| typos | 1.51.1 | Check spelling. |
| Taplo | 0.10.0 | Format and check TOML. |
| GoReleaser | 2.18.3 | Validate release packaging. |

No C toolchain is required; Cress builds with `CGO_ENABLED=0` for releases.

## Getting the code

```sh
git clone https://github.com/cozybadgerde/cress.git
cd cress
```

The default branch is `trunk`.

## Local setup

Install the pinned tools from the repository root:

```sh
mise install
```

If mise asks you to trust `mise.toml`, review it first. Installing tools
needs network access. Go dependencies live in `go.mod` and `go.sum`.

## Project layout

| Path | Contents |
|------|----------|
| `cmd/cress/` | Executable entry point. |
| `internal/` | CLI, generator, content, render model, renderer, and output handling. |
| `schemas/` | Authoring schemas and schema/runtime agreement tests. |
| `docs/` | Product requirements, architecture guide, and this guide. |
| `docs/adr/` | Architectural decision records. |
| `mise.toml` | Pinned tools and developer tasks. |
| `.github/workflows/` | CI workflow. |
| `.goreleaser.yml` | Release packaging scaffold. |

New packages follow the responsibilities in the
[architecture guide](ARCHITECTURE.md#responsibilities) as code arrives.

## Development guidelines

- **Interface first.** Agree on user-facing interfaces before implementing
  them: CLI commands and flags, configuration and frontmatter fields, theme
  contracts, and messages. Record them in the owning document, then build.
- **Standard library first.** Use the Go standard library when it fits the
  need well.
- **Established libraries over custom code.** When the standard library
  does not fit well, use a focused, well-maintained library instead of
  writing our own. For example, use an established CLI library rather than
  parsing commands and flags by hand. Each new dependency still needs a
  current need and a license review.
- **Compact comments.** Keep comments to at most two lines. Write longer
  comments only when the code needs more explanation to be understood
  correctly.

## Development workflow

Run the executable from source:

```sh
go run ./cmd/cress
```

The executable prints root help. Build a prepared site with:

```sh
go run ./cmd/cress site build /path/to/site
```

Replace `/path/to/site` with a site directory containing `content/`.
The [user reference](USER.md) documents the current authoring contract and
new-or-empty output restriction. The
[theme reference](THEMES.md) documents the initial prepared render model.

The CLI uses urfave/cli v3, TOML parsing uses go-toml v2, and Markdown
conversion uses goldmark. Go text provides portable output-name comparison.
Schema tests use a test-only JSON Schema validator. Reviewed licenses and
redistribution notices live in
[third-party notices](THIRD-PARTY-NOTICES.md).

Use `mise tasks ls` to list all tasks. The tasks call ordinary tools, so
native Go commands remain usable directly.

## Building

```sh
mise run build
```

The build writes the executable to `bin/cress`. The `bin/` directory is
ignored by Git.

Validate the release packaging configuration without building or
publishing:

```sh
mise run check:goreleaser
```

The GoReleaser configuration is a packaging scaffold for the three agreed
targets, not a release-ready pipeline. Signing, release-specific component
inventories, review enforcement, and release automation remain in the work
plan. The archive already includes the required notices for current runtime
dependencies.

## Testing and checks

Run all checks and tests without editing files:

```sh
mise run check
```

| Command | Purpose |
|---------|---------|
| `mise run check` | Run every check below and the tests. |
| `mise run check:go-format` | Check Go formatting. |
| `mise run check:go-lint` | Lint Go with golangci-lint's standard set. |
| `mise run check:markdown` | Lint Markdown without external link checks. |
| `mise run check:spelling` | Check spelling with typos. |
| `mise run check:toml-format` | Check TOML formatting. |
| `mise run check:toml` | Validate TOML syntax. |
| `mise run test` | Run `go test ./...`. |
| `mise run fmt` | Format Go and TOML files. |
| `mise run fmt:go` | Format Go files. |
| `mise run fmt:toml` | Format TOML files. |

`check` and `fmt` run their tasks in parallel. Checks never edit files;
run `mise run fmt` to fix formatting.

Unit tests are the default. Integration tests cover real filesystem or
process boundaries. Browser and end-to-end tests are not part of the setup
yet.

Run `go test -short ./...` to skip filesystem integration tests. Schema
tests compare representative parsed TOML values with runtime validation
without downloading schemas or using an external validation service.

CI runs `mise run check` and `mise run build` with the same pinned tools on
pushes to `trunk` and on pull requests.

EditorConfig supplies basic editor defaults. Schema associations for Taplo
will be added when authoring contracts exist.

## Troubleshooting

| Problem | Remedy |
|---------|--------|
| A tool is missing or has the wrong version. | Run `mise install`. `mise run` also installs missing tools. |
| `check:go-format` lists files. | Run `mise run fmt:go`. |
| `check:toml-format` fails. | Run `mise run fmt:toml`. |
| `check:spelling` reports a valid word. | Fix the spelling, or add a `typos` configuration for real project terms. |
