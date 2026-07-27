# Developer guide

This guide covers building Cress from source and finding your way around the
code.

## Before you start

Read the
[contributing guidelines](https://github.com/cozybadgerde/.github/blob/trunk/.github/CONTRIBUTING.md)
and the
[code of conduct](https://github.com/cozybadgerde/.github/blob/trunk/.github/CODE_OF_CONDUCT.md)
first.

## Prerequisites

- Go 1.26 or newer.
- [Task](https://taskfile.dev) to run the project's commands.

The Taskfile is the single source of truth for the build, lint, and test gate.
CI installs Task plus the pinned tools and runs `task check`, so local runs and
CI never drift.

## Getting the code

```bash
git clone https://github.com/cozybadgerde/cress.git
cd cress
task setup
```

`task setup` configures the commit template and downloads dependencies. Install
the pinned lint and security tools once with `task tools`.

## Project layout

```text
cmd/cress/            CLI: build, init, serve, version (urfave/cli v3)
internal/config/      load and validate cress.toml
internal/content/     discover Markdown, parse front matter into pages
internal/render/      Markdown to HTML (goldmark)
internal/theme/       resolve and load themes; the built-in theme is embedded
internal/build/       orchestration: content + theme + config -> public/
internal/serve/       preview server: build, watch, rebuild, serve
internal/scaffold/    `cress init`; the starter site is embedded
internal/version/     build metadata, set by the linker at release
schema/               JSON Schema for cress.toml
docs/                 user, operator, and developer guides
```

The data flow for a build is one direction:

```text
config ─┐
content ─┼─▶ build ─▶ render (Markdown) ─▶ theme (template) ─▶ public/
theme  ─┘
```

Each package has one responsibility. `content` never renders HTML; `render`
never touches templates; `theme` never reads content; `build` is the only
package that writes output.

## The built-in theme

The default theme lives at `internal/theme/builtin/cress/` and is embedded into
the binary with `go:embed`. A `templates/page.html` renders one page, and
`static/style.css` styles it. Editing these files changes the theme every site
gets by default.

## Useful commands

Run the full gate before opening a pull request:

```bash
task check
```

Every step runs even when an earlier one fails, and the failures are listed
together at the end, so one red step does not hide the others.

Other common tasks (see `task --list` for all of them):

```bash
task build            # build the binary into build/
task run -- build     # run cress with arguments
task test             # unit tests
task lint             # go, markdown, and shell linters
task audit            # complexity and schema audits, not part of check
```
