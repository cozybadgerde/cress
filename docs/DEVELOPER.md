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
