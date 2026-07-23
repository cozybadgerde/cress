# CLAUDE.md — cress

Project notes for AI-assisted work on cress. This file is **cress-specific
only**. Role, communication, the change workflow, testing categories
(`unit`/`integration`/`e2e`), the commit-message format, and the architecture
pushback rules all live in the global `~/.claude/CLAUDE.md` and are not repeated
here. The repo's `.gitmessage` is the commit template.

## Issue tracking

The [GitHub issue tracker](https://github.com/cozybadgerde/cress/issues) is the
source of truth for planned work, bugs, and feature requests. Track new issues
there, not in a file; cress keeps no `TODO.md` backlog.

- Use the `gh` CLI (GitHub) for issue operations; it authenticates against the
  signed-in GitHub account and infers the repo from the git remote. Check
  `gh issue <verb> --help` when unsure; `gh --version` prints the build.

### Reading issues

```sh
gh issue list                   # open issues (the default state)
gh issue list -s all            # -s open|closed|all
gh issue list -l enhancement    # filter by label
gh issue view <n>               # title and body
gh issue view <n> --comments    # its comments
```

### Closing issues

GitHub auto-closes an issue when a commit or merged PR that lands on the default
branch carries a closing reference. Use the commit footer for this:

```text
- resolves: #7
```

Once that lands on the default branch, the issue closes on its own. Do not run
`gh issue close` for work resolved this way. For a non-code closure (wontfix,
duplicate), close manually with a reason and comment:

```sh
gh issue close <n> -r "not planned" -c "<reason>"
```

### Labels

On GitHub, labels are **repo-scoped** (unlike Codeberg's org-level set), so each
repo owns its own. List the current labels before tagging:

```sh
gh label list
```

Reuse the standard set where it exists: `bug`, `enhancement`, `question`,
`duplicate`, `invalid`, `good first issue`, `help wanted`, and the
project-specific `contribution welcome` and `upstream`. Create a missing one
with `gh label create "<name>"` rather than inventing near-duplicates.

### Creating an issue with `gh`

`gh issue create` takes `--label` directly; pass the body from a file or inline,
and `--template <name>` fills from a `.github/ISSUE_TEMPLATE` form:

```sh
gh issue create --title "<title>" --body-file <file> --label enhancement
gh issue comment <n> --body-file <file>       # add a comment
```

Issue templates are inherited from the org-wide defaults in
[`cozybadgerde/.github`](https://github.com/cozybadgerde/.github/tree/trunk/.github/ISSUE_TEMPLATE),
not kept in this repo. For an enhancement, mirror that template: a
`### Problem or motivation`, `### Proposed solution`, and `### Alternatives
considered` section. Footer the body with where the idea came from when it helps
traceability.

## What cress is

A cozy static site generator: Markdown plus a small TOML config plus a theme,
rendered into a directory of static HTML. `cress build` renders a site,
`cress init` scaffolds a new one, and `cress serve` previews it with rebuild on
change. It is deliberately less opinionated than Hugo: the content folder tree
maps straight to the output tree, navigation is an explicit flat list, and the
theme is the only place styling lives.

## Architecture / data flow

```text
config ─┐
content ─┼─▶ build ─▶ render (Markdown) ─▶ theme (template) ─▶ public/
theme  ─┘
```

- **`config`** loads `cress.toml`: the `[site]` metadata (title, description,
  base_url, theme, plus the white-label `logo`/`favicon`/`accent`) and a flat
  `[nav]` table (label = content path). Unknown keys are rejected as typos. An
  empty theme resolves to the built-in default; an empty `accent` resolves to
  `#9cb43b` and is validated as a hex color (it is interpolated into CSS). The
  `[nav]` order is recovered from the TOML parse metadata, since decoded tables
  are otherwise unordered.
- **`content`** walks `content/`, splits YAML front matter from the Markdown
  body, and produces a `Page` for each file. It maps the source path to an
  output path and URL (`about.md` -> `about.html`; an `index` file collapses to
  its directory URL). It does no HTML conversion.
- **`render`** owns the goldmark configuration (CommonMark + GFM) and converts a
  Markdown body to an HTML fragment. Nothing else.
- **`theme`** resolves a theme (a directory under `themes/`, else the embedded
  default) and loads its `html/template` set and static assets. `page.html` is
  the required entry template.
- **`build`** is the orchestrator and the only package that writes output. It
  resolves nav links against the collected pages, renders each non-draft page
  through the theme, and copies theme and site static assets into `public/`.
- **`serve`** builds once, serves the output over HTTP, and rebuilds on
  debounced filesystem events. It never watches the output directory.
- **`scaffold`** writes the embedded starter site for `cress init`.

## Package layout

| Path                | Responsibility                                          |
|---------------------|---------------------------------------------------------|
| `cmd/cress`         | CLI (urfave/cli v3). Parses args + formats output only. |
| `internal/config`   | Load and validate `cress.toml`.                         |
| `internal/content`  | Discover Markdown, parse front matter, build `Page`s.   |
| `internal/render`   | Markdown to HTML (goldmark). Pure.                      |
| `internal/theme`    | Resolve and load themes; embedded default theme.        |
| `internal/build`    | Orchestrate content + theme + config into `public/`.    |
| `internal/serve`    | Preview server: build, watch, rebuild, serve.           |
| `internal/scaffold` | `cress init` starter site (embedded).                   |
| `internal/version`  | Build metadata (set via goreleaser ldflags).            |

## cress-specific conventions

- **Layering:** command bodies in `cmd/cress` parse args and format output; all
  real logic lives in `internal/`. `build` is the single writer of output; keep
  `content`, `render`, and `theme` free of cross-cutting I/O.
- **The theme is the only styling surface.** Core emits plain semantic HTML;
  code fences carry `class="language-..."` but no styling. Do not bake CSS or
  syntax highlighting into the core - it would conflict with custom themes.
- **URLs are flat and root-relative.** `content/x.md` -> `public/x.html`, linked
  as `/x.html`; an `index` file collapses to its directory (`/`, `/guide/`). The
  builder writes root-relative links, so a site is served from its domain root.
- **Navigation is explicit.** There are no automatic list or section pages. A
  `[nav]` entry (label = content path) points at a content file; a missing
  target is a warning, not a hard error, so the rest of the build still succeeds.
- **Embedded assets:** the default theme (`internal/theme/builtin/cress`) and the
  starter site (`internal/scaffold/builtin`) are embedded with `go:embed`. Both
  are copied verbatim, so edits to those files change what ships.
- **Schema stays in sync with config:** `schema/config.schema.json` (draft
  2020-12) mirrors the `Config` struct. Validate the bundled starter config with
  `task audit:schema` (taplo); the starter carries the `#:schema` directive.

## Roadmap notes

Planned but out of scope for v0.1: internationalization (i18n). Deliberately not
planned: plugins or a JS ecosystem, data files, shortcodes, taxonomies, and
pagination.

## Build / test / lint

The **Taskfile is the single source of truth** - CI installs Task plus the
pinned tools and runs `task check`, so local and CI never drift. Don't run raw
`go test ./...` / `golangci-lint` in CI; call the task.

| Command             | What it does                                          |
|---------------------|-------------------------------------------------------|
| `task check`        | Full gate: fmt:check, vet, lint, sec, test, test:race |
| `task test`         | Unit tests                                            |
| `task lint`         | golangci-lint + gomarklint + shellcheck               |
| `task tools`        | Install all pinned dev tools (matches CI)             |
| `task audit:schema` | Validate the bundled starter config against its schema |
| `task run -- <args>`| Run cress with args (e.g. `task run -- build`)        |

CI lives in `.github/workflows/` (`ci`, `release`, `security`). Default branch
is `trunk`; releases fire on `v*` tags via goreleaser.
