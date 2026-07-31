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
cmd/cress/            CLI: build, clean, init, serve, version (urfave/cli v3)
internal/config/      load and validate cress.toml
internal/content/     discover Markdown, parse front matter into pages
internal/render/      Markdown to HTML (goldmark)
internal/theme/       resolve and load themes; the built-in theme is embedded
internal/build/       orchestration: content + theme + config -> public/
internal/clean/       `cress clean`; the only package that deletes
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
package that writes output, and `clean` the only one that removes it. Those
last two are deliberately separate: a build runs unattended and must never
delete, so the code that can delete is reachable only from a command a user
names.

Command bodies in `cmd/cress` parse arguments and format output. All real logic
lives under `internal/`, which is what keeps it testable without a terminal.

## Design decisions

These are the choices a change is most likely to walk into without knowing they
were choices. Each one has an alternative that looks obviously better until you
hit the case it fails on.

**The build never deletes.** It creates and overwrites its own files and leaves
everything else alone, reporting what it did not write as a warning. The
tempting alternative is pruning whatever a previous run did not produce. No
guard can decide which paths are safe to remove across every machine, CI runner,
and platform a build might run on, and the cost of being wrong is somebody's
data. Removing output is `cress clean`'s job instead.

**Deleting takes more consent than naming the command.** `cress clean` was
originally specified so that typing it was consent enough. It is not, because
`--output` makes the target whatever was last typed, so the resolved absolute
path and the file count are confirmed before anything goes. `--force` skips the
question and nothing else; the guards stay hard errors with or without it.

**The theme is the only styling surface.** The core emits plain semantic HTML.
Code fences carry `class="language-..."` and no styling, and the same goes for
every other element. Baking CSS or syntax highlighting into the core would put
it in conflict with every custom theme, which is why config values that could
inject markup are escaped rather than rendered.

**URLs are root-relative, not relative.** Every emitted URL is prefixed with the
path component of `base_url`, so one build serves from a subdirectory or a
domain root. Fully relative URLs were the alternative and need no config at all,
but they make every URL depend on the page's own depth, which breaks against a
host that does not redirect `/guide` to `/guide/`. Three places carry the
prefix: the builder, `render` for the links an author wrote, and the theme for
its own asset links, which are the URLs cress does not emit.

**Embedded assets ship verbatim.** The default theme
(`internal/theme/builtin/cress`) and the starter site
(`internal/scaffold/builtin`) are embedded with `go:embed` and copied without
transformation, so an edit to either changes what every user gets.

**The schema tracks the config struct.** `schema/config.schema.json` (draft
2020-12) mirrors `config.Config`, and the scaffolded `cress.toml` carries a
`#:schema` directive pointing at it, so editors validate a user's config. A new
config key means a schema change in the same commit; `task audit:schema`
validates the bundled starter against it.

## The built-in theme

The default theme lives at `internal/theme/builtin/cress/` and is embedded into
the binary with `go:embed`. A `templates/page.html` renders one page, and
`static/style.css` styles it. Editing these files changes the theme every site
gets by default.

`page.html` is the only template a theme must define. A theme may add
`templates/404.html` to render the 404 page itself; without one, `build`
synthesizes a 404 and renders it through `page.html`. That is why the 404 is
free for every theme instead of being a second required template.

The rendered Markdown a theme drops into the page is plain semantic HTML, with
one shape worth knowing about: an image that is the whole of its paragraph
arrives wrapped in a `<figure>`, with the image title as a `<figcaption>` when
the author wrote one. The wrapper is there with or without a caption, so a
theme styles one shape rather than two. A theme that styles `p img` and not
`figure img` misses every standalone image on the site.

## The template data contract

Every template receives a `theme.PageData`, defined in `internal/theme/data.go`
together with `NavView`, `NavLink`, and `PageView`. `build` fills the value in,
and `theme` owns the definition, because it already owns the templates.

Templates use Go's `html/template`, and every one of them receives:

- `.Site`: the `[site]` config, as `.Site.Title`, `.Site.Description`,
  `.Site.BaseURL`, `.Site.Logo`, `.Site.LogoDark`, `.Site.Favicon`,
  `.Site.Accent`, `.Site.AccentDark`, `.Site.Footer`, and `.Site.Copyright`,
  plus the derived `.Site.BasePath` described below.
- `.Nav`: the resolved navigation, as `.Nav.Main` and `.Nav.Footer`. Each is a
  list of entries with `.Title`, `.URL`, and `.Active`, true on the current
  page. A group with no entries is empty, so `{{ with .Nav.Footer }}` skips it.
- `.Page`: the current page, with `.Page.Title`, `.Page.URL`, `.Page.HTML` (the
  rendered Markdown body), `.Page.Description`, and `.Page.Meta` (the raw front
  matter), plus four that are easy to miss:
  - `.Page.Language`, resolved from the page's front matter then the site
    config. Never empty, and it belongs in the document's `lang` attribute:
    `<html lang="{{ .Page.Language }}">`.
  - `.Page.AbsoluteURL`, the full URL, `Site.BaseURL` joined with `URL`. Empty
    when the site sets no `base_url`, since there is no host to build one from,
    so a template that renders it has to guard it.
  - `.Page.IsHome`, true only for the page served at the site root. Themes use
    it for what reads differently on a front page, such as dropping the site
    name from a title that already is the site name.
  - `.Page.Head`, ready-made metadata elements for the document head, already
    escaped. It carries only what describes the page; the title, charset,
    viewport, icon, and stylesheets stay the theme's own. A theme that would
    rather arrange the metadata itself reads `.Page.Description` and
    `.Page.AbsoluteURL` and omits this, because rendering both emits every tag
    twice.

Those field names are a public API. Every theme is written against `.Site`,
`.Nav`, and `.Page`, including themes cress never sees, so renaming or removing
one is a breaking change and takes the same care as changing a CLI flag. Each
field carries a doc comment naming its type, when it is empty, and whether it
is already escaped. `.Page.HTML` is a `template.HTML` for that reason: it holds
rendered Markdown, and escaping it a second time would print the page source to
the reader.

Two things hold the contract in place. `Render` and `RenderTemplate` take
`PageData` rather than `any`, so a caller cannot quietly invent a second shape.
`TestPageDataContract` renders a template that reads every field, which turns a
rename into a failing test instead of a broken build on a user's machine.

One field carries an obligation for the theme. `.Site.BasePath` is the path the
site is published under, from the path component of `base_url`, and it is empty
for a site at a domain root. Every URL in `PageData` is already rooted under it,
so a theme needs it only for its own asset links:

```html
<link rel="stylesheet" href="{{ .Site.BasePath }}/style.css" />
```

A theme that hardcodes `/style.css` instead works at a domain root and renders
unstyled under a subdirectory, and nothing reports it: the build cannot tell
which URLs a theme meant to be its own. The built-in theme uses the form above
for its stylesheet and its home link, which is the pattern to copy.

## Image assets

These rules govern what cress itself ships: the embedded theme, the starter
site, and the images under `docs/`. They are not rules for a user's own site.
`static/` is copied verbatim, and cress neither converts nor refuses whatever
format an author puts there.

One format per role:

| Role                 | Format | Why                                                             |
| -------------------- | ------ | --------------------------------------------------------------- |
| Logo, line art       | SVG    | Scales from one file, and stays a few hundred bytes.             |
| Logo, shaded artwork | WebP   | Tracing a gradient illustration yields a large, poor SVG.        |
| Favicon              | PNG    | WebP favicon support is uneven, and there is nothing to save.    |
| Content images       | WebP   | Illustration and photography, where the compression pays.        |

The logo row splits on the artwork, not on the role. The starter site's
placeholder is flat line art, so it is an SVG of a few hundred bytes. Cress's
own mark is a shaded illustration, so it is WebP.

There is no asset pipeline and there is not going to be one. Committed files are
already in their final format and size, so converting is something you do before
`git add`, never something `cress build` does.

- **Size it for where it is shown, then check it there.** The header logo
  renders at `1.7rem`, and the content column is `44rem`, so a content image
  needs 1408px at most to stay sharp on a 2x display. Verify at that size rather
  than at full resolution: a resample that looks perfect at 1:1 can still
  destroy small text.
- **Lossless for artwork carrying text**, lossy (q80 to q90) for illustration
  and photography. Sharp type is where lossy artifacts appear first.
- **Everything under `internal/` is embedded** with `go:embed` and ships in
  every binary forever, so it earns a stricter bar than anything in `docs/`.
- **Masters stay out of the repository.** What is committed is derived and sized
  for its use; the original belongs in whatever tool drew it.
- **Name a file for its role**, not for what it depicts.

The starter site's placeholders are deliberately generic rather than cress's own
mark, because that slot stands in for the user's brand: a site that never
touches it should not end up advertising cress. The placeholder logo uses a
single mid-tone green that reads against a light and a dark background alike, so
it needs no dark variant.

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

`task lint:md` covers the root Markdown files, `docs/`, and the starter site
under `internal/scaffold/builtin/content/`, which is linted because it ships
verbatim to every user. Rules live in `.gomarklint.json`. One is off
deliberately: `no-bare-urls`, because the scaffolded style guide demonstrates
that a bare URL becomes a link on its own, and that line is the point rather
than an oversight. gomarklint does support an inline
`<!-- gomarklint-disable -->` comment, but the scaffold is copied into user
sites, so a linter directive there would leak this repo's tooling into theirs.

## Continuous integration

Workflows live in `.github/workflows/`: `ci`, `release`, and `security`. CI
installs Task plus the pinned tools and runs `task check`, which is why the
Taskfile is the single source of truth rather than a convenience wrapper. Do not
call `go test ./...` or `golangci-lint` directly from a workflow; call the task,
so a change to the gate reaches CI and your machine at the same time.

The default branch is `trunk`, and a release fires from a `v*` tag via
goreleaser.

Dependabot (`.github/dependabot.yml`) watches the workflow actions and the
`go.mod` requirements. Each ecosystem groups its patch and minor bumps into one
weekly pull request, and majors arrive on their own. The Go module list includes
indirect dependencies, because Dependabot skips those by default and that is
where drift hides. Dependabot alerts and security updates are repository
settings rather than keys in that file: they cover what GitHub's advisory
database knows about, and the config covers what has no advisory yet.
