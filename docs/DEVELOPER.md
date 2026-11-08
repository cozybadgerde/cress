# Developer guide

This guide explains how to set up a local development environment, change the
source, and build and test Cress.

## Before you start

These documents explain what Cress should do and why:

- [Product requirements](PRD.md): scope and acceptance criteria for 1.0.
- [Architecture document](ARCHITECTURE.md): responsibilities, data flow, and
  failure boundaries.
- [Architectural decision records](PRD.md#decision-records): accepted
  decisions and their rationale.

The [contribution notes](../README.md#contributing) explain how to propose a
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
| `docs/` | Guides and documents; see the [README](../README.md#documentation). |
| `docs/adr/` | Architectural decision records. |
| `mise.toml` | Pinned tools and developer tasks. |
| `.github/workflows/` | CI and version-tag release workflows. |
| `.goreleaser.yml` | Archive, dependency-report, checksum, and signing configuration. |

New packages follow the responsibilities in the
[architecture document](ARCHITECTURE.md#responsibilities) as code arrives.

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
The generated `cress.toml`, the [content guide](CONTENT.md), and the
[build and preview document](OUTPUT.md) describe authoring, output
ownership, and recovery. The
[theme developer guide](THEMES.md) documents the prepared render model.

See the [architecture document](ARCHITECTURE.md#responsibilities) for component
responsibilities and the [release guide](RELEASE.md#license-reports) for
dependency license reports and notices.

Use `mise tasks ls` to list the tasks defined in
[`mise.toml`](../mise.toml). The tasks call ordinary tools, so native Go
commands remain usable directly.

## Building

```sh
mise run build
```

The build writes the executable to `bin/cress`. The `bin/` directory is
ignored by Git. Use this executable for the test sites below.

### Test sites

Create fresh test sites with `site init`. The repository's `/tmp/` tree is
gitignored, including sources and generated output:

```sh
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

For Anemone, use `--theme anemone` and a new directory such as `tmp/anemone`.
Review the content tour and pages with and without images.

Edit the fixture's sources to test a change. Do not initialize over an
existing project. Add test-specific content, drafts, or shared assets as
needed. See the [user guide](USER.md) for the authoring workflow
and [output protection](OUTPUT.md#output-protection) for safety rules.

Theme authors can use the same scaffold with a
[local development theme](THEMES.md#create-your-first-theme).
Manual test sites do not replace automated tests.

For packaging changes, also run `mise run check:goreleaser`.
The [release guide](RELEASE.md) cover the pre-tag checklist, local
rehearsal, and publication.

## Testing and checks

Run all checks and tests without editing files:

```sh
mise run check
```

The task definitions in [`mise.toml`](../mise.toml) list the checks and tools.
Use `mise tasks ls` for their current names and descriptions.
Checks never edit files. Run `mise run fmt` to apply formatting.

Unit tests are the default. Integration tests cover real filesystem or
process boundaries. A persistent browser or end-to-end suite is outside
the initial scope.

Run `go test -short ./...` to skip filesystem integration tests. Schema
tests compare representative parsed TOML values with runtime validation
without downloading schemas or using an external validation service.

CI runs `mise run check`, `mise run build`, and `mise run check:goreleaser`
with the same pinned tools on
pushes to `trunk` and on pull requests. These tasks automatically include
the configuration/frontmatter schema tests and filesystem integration
tests.

EditorConfig supplies basic editor defaults. External schema validation is
documented in the [user guide](USER.md#editor-support);
no editor-specific integration is required.

## Troubleshooting

| Problem | Remedy |
|---------|--------|
| A tool is missing or has the wrong version. | Run `mise install`. `mise run` also installs missing tools. |
| `check:go-format` lists files. | Run `mise run fmt:go`. |
| `check:toml-format` fails. | Run `mise run fmt:toml`. |
| `check:spelling` reports a valid word. | Fix the spelling, or add a `typos` configuration for real project terms. |
