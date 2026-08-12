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
cmd/cress/            CLI: build, clean, init, serve, theme, version (urfave/cli v3)
internal/config/      load and validate cress.toml
internal/content/     discover Markdown, parse front matter into pages
internal/render/      Markdown to HTML (goldmark, and chroma when highlighting)
internal/theme/       resolve, load and validate themes; the built-in ones are embedded
internal/build/       orchestration: content + theme + config -> public/
internal/clean/       `cress clean`; the only package that deletes
internal/serve/       preview server: build, watch, rebuild, serve
internal/scaffold/    `cress init` and `cress theme init`; both starters are embedded
internal/version/     build metadata, set by the linker at release
schema/               JSON Schema for cress.toml and theme.toml
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
every other element. Baking CSS or a color scheme into the core would put it in
conflict with every custom theme, which is why config values that could inject
markup are escaped rather than rendered.

**Highlighting emits classes, never colors.** `[markdown] highlight` looks like
an exception to the rule above and is not: chroma runs in classes-only mode, so
the core hands a theme more structure to style and still none of the styling. It
is off by default, and a site with no `[markdown]` table produces the HTML it
produced before the table existed. The colors sit in the theme's own
`highlight.css`, self-contained so that a site's `static/highlight.css` replaces
the scheme outright, and linked unconditionally because a template cannot see
`[markdown]` and making it conditional would widen the contract to save a
request.

**URLs are root-relative, not relative.** Every emitted URL is prefixed with the
path component of `base_url`, so one build serves from a subdirectory or a
domain root. Fully relative URLs were the alternative and need no config at all,
but they make every URL depend on the page's own depth, which breaks against a
host that does not redirect `/guide` to `/guide/`. Three places carry the
prefix: the builder, `render` for the links an author wrote, and the theme for
its own asset links, which are the URLs cress does not emit.

**Embedded assets ship verbatim.** The built-in themes
(`internal/theme/builtin/`), the starter site
(`internal/scaffold/builtin/site`) and the starter theme
(`internal/scaffold/builtin/theme`) are embedded with `go:embed` and copied
without transformation, so an edit to any of them changes what every user gets.
Both starters go through one writer, which never overwrites: `--allow-existing`
relaxes the empty-directory precondition and grants no permission to replace a
file.

**The schemas track the structs.** `schema/config.schema.json` mirrors
`config.Config` and `schema/theme.schema.json` mirrors `theme.Meta` (both draft
2020-12). Every TOML file Cress bundles carries a `#:schema` directive pointing
at one of them, so editors validate what a user writes. A new key in either
struct means a schema change in the same commit; `task audit:schema` validates
the bundled files against them.

**The theme checker reports only what it is certain of.** `cress theme validate`
walks the parse trees rather than rendering a sample page, and it tracks what
the dot is through `range`, `with` and variable assignment. Wherever it stops
being able to name a type, it stops checking rather than guessing: a fragment
reached with two different types is skipped, and so is everything under
`.Page.Meta`, which is a map and takes any key by design. That asymmetry is the
whole design. A genuine unknown field already fails the build, so the command's
only contribution is finding one early and finding all of them at once, and a
single false positive would trade that for an author who turns it off. When a
check cannot be made certain it does not ship: that is why nothing verifies
that a referenced `static/` file exists, since a site legitimately supplies its
theme's assets.

**A theme's own metadata can only ever warn.** `theme.toml` is optional
provenance plus the `cress` version a theme was written against, and the loader
treats an unknown key and a version mismatch alike: both produce a build warning
and neither stops a build. The strict-key rule `cress.toml` uses is wrong here,
because an unknown key is also what a published theme looks like to an older
Cress once a later one adds a field, and nothing in the file decides how a page
renders. Only a `theme.toml` that is not TOML at all is an error, since there is
nothing in it to read. The built-in themes deliberately declare no `cress`
version: it ships inside the binary, so the contract it was written against is
always the one rendering it, and a declaration there could never catch a
mismatch and could only ever be wrong.

## The built-in themes

Each directory under `internal/theme/builtin/` is a theme embedded into the
binary with `go:embed` and reachable by that directory's name: `cress`, the
cozy default, and `birch`, a crisp developer-tool look built around code.
`templates/page.html` renders one page, `templates/landing.html` renders a page
that names `layout: landing`, `templates/partials/` holds what those two share,
and `static/` styles the result. Editing these files changes what every user
gets.

Adding a third is adding a directory: the embed pattern takes the whole tree and
`theme.Builtins` reads it back, so nothing lists the names a second time. The
one thing that is not automatic is `config.DefaultTheme`, which names the
fallback for an unset `theme` key and is deliberately not the same question as
which themes exist.

`page.html` is the only template a theme must define. A theme may add
`templates/404.html` to render the 404 page itself; without one, `build`
synthesizes a 404 and renders it through `page.html`. That is why the 404 is
free for every theme instead of being a second required template.

Every other file directly in `templates/` is a layout a page can name.
`theme.Resolve` parses `templates/*.html` into one template set keyed by base
name, so `layout: wide` in front matter reaches `wide.html` and nothing else has
to be registered. Fragments live one level down in `templates/partials/`. They
are parsed into the same set, so any layout can call them, and they are reached
only by their `{{ define }}` name. That tree is walked rather than globbed,
because a partial's path carries no meaning: a theme may group its fragments
into subdirectories without any of it reaching the loader.

The [theme guide](./THEME.md) states those rules for the people who write
themes. What matters here is why the implementation is shaped that way, because
each one has a change that looks like a simplification:

- **A layout name is a map key, not a path.** `HasLayout` looks up a name the
  theme itself declared, so a layout name needs no sanitizing and cannot escape
  the theme. Do not add path cleaning here; there is no path.
- **A miss is not an error.** An unknown layout falls back to `page.html` and
  adds a build warning, which is what lets a site keep building when its theme
  is swapped for one with a different vocabulary. It matches how an unresolvable
  nav entry behaves, and both are deliberate.
- **The layout set comes from file names, not from the parsed set.** Deriving it
  from the parse would put layouts and partials in one namespace, which is the
  thing that split undoes.
- **A `{{ define }}` name ending in `.html` is refused at load.** Such a name
  can take a layout's, and a later parse replaces an earlier one, so the layout
  set would still list the name while the template behind it had become the
  define's. Forbidding the shape makes that unrepresentable rather than
  something to detect. Detecting it instead would mean comparing template sets,
  which cannot see a redefinition: replacing a name adds nothing to compare.

Cress promises no layout vocabulary, so the built-in themes' `landing` is an
offer rather than a standard. Both ship one, because the scaffolded site's home
page names it and a built-in that warned on the scaffold would be a poor
advertisement. Their layouts call the shared partials in `templates/partials/`
for the head, the header, and the footer rather than repeating them, so adding a
third layout cannot leave a theme with two different headers.

`landing` differs from `page` below the header rather than in place of it: it
adds a panel sized to the first screen and keeps the site's usual navigation.
Four details in it are worth preserving in any layout added later:

- The panel's site name is a `<p>`, not a heading. The page's H1 comes from its
  Markdown, and a second one would leave the page with two.
- The header and panel are sized together, as `.landing-top`, so the panel takes
  whatever height the header leaves. Sizing the panel alone needs a magic number
  that goes stale the moment the header's padding changes.
- The arrow at the panel's foot is an `<a href="#content">` with an
  `aria-label`, not an ornament, so it can be tabbed to and followed and does
  not announce as a shape. It is omitted when `.Page.HTML` is empty, where it
  would point at nothing.
- Its animation sits behind `@media (prefers-reduced-motion: no-preference)`.
  Motion that repeats forever is what that preference exists to switch off, and
  the arrow still points down without it.
- The panel's highlight is drawn on a `::before` layer, never on `.hero` itself,
  because a filter or an opacity set on the panel would take the title down with
  the artwork.

The highlight is the one place the template hands the stylesheet a value. Only
the template knows `.Site.Logo`, so `landing.html` emits it as a `--site-logo`
custom property and picks the class that says which highlight to draw:
`hero-logo` for a site with a logo, `hero-wash` for one without. `html/template`
escapes the URL in its CSS context, replacing anything that is not a plain URL
with `#ZgotmplZ`, so a logo path cannot become a style injection.

The logo is flattened with `grayscale(1) brightness(0)`, which discards its
color and keeps only its shape. That is what makes it work for a logo the theme
has never seen: desaturating alone leaves a dark logo invisible on the dark
scheme and a pale one invisible on the light, while a silhouette tinted by
opacity lands at the same weight either way. The `hero-wash` fallback is mixed
from `var(--accent)` rather than stored as a color of its own, for the reason
given in the accent discussion: a second stored color drifts out of step with
the configured one.

The rendered Markdown a theme drops into the page is plain semantic HTML. The
shapes a theme has to style, including the `<figure>` wrapper a standalone image
arrives in, are listed in the [theme guide](./THEME.md#what-cress-emits). Adding
a shape to what `render` emits means adding it there too, because a theme cannot
style what nothing told it about.

## The starter theme

`cress theme init` scaffolds a different theme: the starter at
`internal/scaffold/builtin/theme/`, which is close to unstyled and commented
line by line. It is deliberately not a fork of the built-in theme, so an author
starts from the contract rather than from this project's design decisions.

Four things about it are load-bearing:

- **It ships `landing.html`.** The scaffolded site's home page names
  `layout: landing`, so a starter without it would warn on the first build after
  a user switched themes. A starter that warns out of the box teaches its user
  to ignore warnings.
- **A test builds a site through it.** `html/template` fails on a field a struct
  does not have, so renaming one in `PageData` breaks that test rather than
  quietly emitting an empty string in every theme scaffolded afterwards. Keep it
  that way: the starter is the one template set in this repository whose job is
  to be copied.
- **A test validates it.** Both bundled themes go through
  `cress theme validate` in the test suite, so cress cannot ship an example of
  something its own validator rejects.
- **It ships a `theme.toml` that declares a `cress` version.** Unlike the
  built-in theme, the starter is copied out of the binary and published, so it
  is where an author reads that field for the first time. The test above pins
  the running version to whatever the starter declares before it builds, because
  an untagged build has no version, would skip the check, and would leave a typo
  in that file invisible.

## The template data contract

Every template receives a `theme.PageData`, defined in `internal/theme/data.go`
together with `NavView`, `NavLink`, and `PageView`. `build` fills the value in,
and `theme` owns the definition, because it already owns the templates. That
holds for every template equally: `page.html`, `404.html`, and any layout a page
names all get the same shape, so a layout rearranges what is already there
rather than receiving data of its own.

The field-by-field reference is in the
[theme guide](./THEME.md#the-template-data-contract), where the people writing
against it will look. What a contributor needs is narrower.

**These field names are a public API.** Every theme is written against `.Site`,
`.Nav`, and `.Page`, including themes cress never sees, so renaming or removing
one is a breaking change and takes the same care as changing a CLI flag. The
contract is frozen for the life of a major version.

**The doc comments are part of it.** Each field names its type, when it is
empty, and whether it is already escaped. `.Page.HTML` is a `template.HTML` for
that reason: it holds rendered Markdown, and escaping it a second time would
print the page source to the reader. A new field arrives with a doc comment
answering those three questions, or it arrives incomplete.

**Two things hold it in place.** `Render` and `RenderTemplate` take `PageData`
rather than `any`, so a caller cannot quietly invent a second shape.
`TestPageDataContract` renders a template that reads every field, which turns a
rename into a failing test instead of a broken build on a user's machine. A
field added without a line in that test is a field the suite does not defend.

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
under `internal/scaffold/builtin/site/content/`, which is linted because it
ships verbatim to every user. Rules live in `.gomarklint.json`. One is off
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
