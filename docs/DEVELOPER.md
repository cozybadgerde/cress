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

Generator capabilities are implemented; external release evidence and
human accessibility review remain outstanding in the work plan. The
[contribution notes](../README.md#contributing) explain how to propose a
change.

## Prerequisites

- Linux or macOS.
- [Git](https://git-scm.com/).
- [mise](https://mise.jdx.dev/getting-started.html), which installs every
  other development tool.

[`mise.toml`](../mise.toml) defines the development tools, pinned versions,
and tasks. Use `mise install` as described below rather than installing
each tool separately.

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
needs network access. [`go.mod`](../go.mod) declares Go dependencies;
[`go.sum`](../go.sum) records their checksums. Go commands download required
modules automatically.

## Project layout

| Path | Contents |
|------|----------|
| `cmd/cress/` | Executable entry point. |
| `internal/` | CLI, generator, content, render model, renderer, and output handling. |
| `schemas/` | Authoring schemas and schema/runtime agreement tests. |
| `docs/` | Product requirements, architecture guide, and this guide. |
| `docs/adr/` | Architectural decision records. |
| `mise.toml` | Pinned tools and developer tasks. |
| `.github/workflows/` | CI and version-tag release workflows. |
| `.goreleaser.yml` | Archive, dependency-report, checksum, and signing configuration. |

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
output ownership and recovery workflow. The
[theme reference](THEMES.md) documents the prepared render model.

See the [architecture guide](ARCHITECTURE.md#responsibilities) for component
responsibilities and [third-party notices](THIRD-PARTY-NOTICES.md) for
generated dependency reports and redistribution notices.

Use `mise tasks ls` to list all tasks. The tasks call ordinary tools, so
native Go commands remain usable directly.

### Test sites

Create fresh test sites with `site init`. The repository's `/tmp/` tree is
gitignored, including sources and generated output:

```sh
mise run build
mkdir -p tmp
./bin/cress site init tmp/poppy
./bin/cress site build tmp/poppy
./bin/cress site preview --port 0 tmp/poppy
```

Open the printed URL and stop preview with Ctrl-C. To compare clover:

```sh
./bin/cress site init --theme clover tmp/clover
./bin/cress site preview --port 0 tmp/clover
```

For the paper/Scandi theme, use `--theme anemone` and a new site directory,
such as `tmp/anemone`. Review it with and without images as well as with
the content tour.

Initialization requires a new or completely empty target with an existing
parent directory. Reuse an initialized site by editing its sources and
rebuilding; choose another new directory for a fresh scaffold.
The sample covers every page type and includes a content tour and commented
configuration. Add test-specific content, drafts, or shared `static/` assets
as needed. Use preview's `--drafts` flag to inspect unfinished pages.
Preview does not change deployable `public/` output.

Theme authors can use the same scaffold with a
[local development theme](THEMES.md#testing-a-theme).
Publishers can follow [testing-a-site guidance](USER.md#testing-a-site)
with an installed executable. Manual test sites do not replace automated tests.

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

The release configuration packages the three agreed targets with notices,
installation guidance, per-target CSV license reports, and signed checksums.
See [release maintenance](RELEASE.md) for review, snapshot validation, and
the external evidence still required before publishing.

## Testing and checks

Run all checks and tests without editing files:

```sh
mise run check
```

| Command | Purpose |
|---------|---------|
| `mise run check` | Run every check below and the tests. |
| `mise run check:go-format` | Check Go formatting. |
| `mise run check:go-lint` | Lint Go using the repository configuration. |
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

CI runs `mise run check`, `mise run build`, and `mise run check:goreleaser`
with the same pinned tools on
pushes to `trunk` and on pull requests. These tasks automatically include
the configuration/frontmatter schema tests and filesystem integration
tests. Current phase-completion and commit/push coordination live in the
[work plan](../TODO.md#phase-workflow).

EditorConfig supplies basic editor defaults. External schema validation is
documented in the [user reference](USER.md#external-schema-validation);
no editor-specific integration is required.

## Troubleshooting

| Problem | Remedy |
|---------|--------|
| A tool is missing or has the wrong version. | Run `mise install`. `mise run` also installs missing tools. |
| `check:go-format` lists files. | Run `mise run fmt:go`. |
| `check:toml-format` fails. | Run `mise run fmt:toml`. |
| `check:spelling` reports a valid word. | Fix the spelling, or add a `typos` configuration for real project terms. |
