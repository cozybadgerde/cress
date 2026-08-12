# Theme guide

This guide is for writing a Cress theme: the templates and the styling that
turn rendered Markdown into a site. It covers the directory a theme lives in,
the data every template receives, the rules the loader enforces, and the
conventions a theme has to honor to work at any address.

If you only want to use a theme somebody else wrote, the
[user guide](./USER.md#themes) covers picking one and naming it in the config.

## Before you start

A theme is templates and CSS. There is no Go to write, no plugin API, and
nothing to compile. Cress renders your Markdown to HTML and hands it to your
template; everything about how a page looks is yours.

Cress emits plain semantic HTML and no styling at all. That is the deal: the
core will never ship a stylesheet that fights yours, and in exchange every
visual decision, including the ones you would rather not have made, is the
theme's.

## Start from the scaffold

Scaffold a theme rather than starting from an empty directory:

```bash
cress theme init mine
```

That writes `themes/mine/` with two layouts, the partials they share, a
stylesheet, and a `theme.toml` to fill in. Every file is commented, so the
starter is a working tour of what this guide describes. Name it in `cress.toml`
and preview it:

```toml
[site]
theme = "mine"
```

```bash
cress serve
```

The [user guide](./USER.md#themes) documents the command's flags. The starter is
deliberately close to unstyled: it is a correct theme to build on, not a design
to unpick.

## Anatomy of a theme

A theme is a directory under `themes/`, named by whatever `cress.toml` says:

```text
themes/
  mine/
    theme.toml       optional: what the theme says about itself
    templates/
      page.html      required: renders one page
      landing.html   optional: an alternative layout, named by a page
      404.html       optional: renders the 404 page
      partials/      optional: fragments the layouts share
        head.html
        site-header.html
    static/          optional: CSS and other assets, copied to the site root
```

`page.html` is the only file a theme must have. A theme can be one template and
one stylesheet.

Everything in `static/` is copied to the root of the built site, unchanged. A
file at `themes/mine/static/style.css` is served at `/style.css`. Cress converts
nothing and imposes no formats on what you put there.

A directory under `themes/` wins over a built-in theme of the same name, and the
build warns when that happens, since nothing in the finished page says which of
the two rendered it. A name that is neither is an error naming the built-in
ones, so a typo in `theme =` fails the build rather than silently falling back.

## What a theme says about itself

`theme.toml` at the root of the theme is how it introduces itself. The file is
optional and so is every field in it, so a theme without one works exactly the
same. Write it before you give the theme to anybody else:

```toml
name = "Mine"
description = "One line on what the theme is for."
author = "You"
license = "MIT"
homepage = "https://github.com/you/mine"

cress = "1.1"
```

`name` is a display name. Cress finds a theme by its directory, so this is what
people read and never what gets resolved: renaming the directory is what renames
the theme, and the two are allowed to disagree.

`cress` is the field that does something. It names the template contract you
wrote the theme against, which is the version of Cress whose `.Site`, `.Nav`, and
`.Page` fields you built on. Cress compares it against itself and warns in either
direction:

- **An older Cress** may not have every field your templates read. Those render
  as an empty string rather than failing, so without the declaration the site
  looks quietly wrong instead of saying so.
- **A Cress a whole major version newer** no longer promises the contract you
  wrote against, since the field names are only fixed for the life of a major
  version.

Either way it is a warning and the site still builds. Cress will not refuse to
build somebody's site because their theme author was cautious about a version
number.

Two consequences worth knowing:

- **Set it to the version you actually developed against**, not the newest one
  you can think of. Naming a version older than the one running is never
  punished, so `1.1` stays correct for the whole of 1.x.
- **A key Cress does not recognize is a warning, not an error.** That is also
  what a theme looks like to an older Cress once a later one adds a field, and
  nothing in this file decides how a page renders. It should not be able to stop
  a build.

An untagged development build of Cress has no version to compare against, so it
skips the check entirely. A `theme.toml` that is not valid TOML is the one hard
error here, because there is nothing in it to read.

## Layouts

A `.html` file directly in `templates/` is a layout. It renders a whole page. A
page chooses one in its front matter, by the file name without the extension:

```markdown
---
title: Welcome
layout: landing
---
```

That page renders through `templates/landing.html`.

Cress promises no layout vocabulary. Which layouts exist, and what each one
means, is yours to define and to document for the people using your theme.
`landing` means whatever your theme decides it means.

Two consequences follow, and both are deliberate:

- **A page naming a layout you do not have still builds.** It renders through
  `page.html` and the build prints a warning. That is what lets somebody swap
  themes without their site breaking, and it matches how an unresolvable
  navigation entry behaves.
- **Layouts are flat.** A layout is named by its file, so `templates/` is not
  searched recursively. Two files with one base name would be one name.

## Partials

Anything under `templates/partials/` is a fragment. It is reached by its
`{{ define }}` name, from a template that composes it, and never by a page:

```html
{{- define "site-header" -}}
  <header>
    <a href="{{ .Site.BasePath }}/">{{ .Site.Title }}</a>
  </header>
{{- end -}}
```

```html
{{ template "site-header" . }}
```

Every template in a theme is parsed into one set, so any layout can call any
partial. A partial receives whatever data its caller passed, which is why the
call above ends in `.`: it hands the whole `PageData` down.

Because a partial is named by its `{{ define }}` and not by its path, the
directory tree under `partials/` means nothing to Cress. Nest it however suits
the theme. `partials/nav/main.html` is as valid as `partials/head.html`.

Use partials for anything two layouts share. A second layout that repeats the
header instead of calling it is how a theme ends up with two different headers.

### The naming rule

Layouts are named by their file, so a layout name always ends in `.html`. A
`{{ define }}` name must not:

```html
{{ define "head" }}        <!-- correct -->
{{ define "head.html" }}   <!-- refused -->
```

Cress refuses to load a theme that breaks this, naming the file and the define.
The reason is that a define taking a layout's name would replace it: the layout
would still be listed, while the template behind it had become the fragment, and
every page would render through the wrong thing with nothing said about it.

The rule is easy to keep by accident, since a fragment has no reason to be named
like a file. It is worth stating because most other generators address partials
by path, and that habit produces exactly the name this rule forbids.

## The template data contract

Every template receives the same value, whatever renders it. `page.html`,
`404.html`, and every layout a page can name all get one `PageData`, so a layout
rearranges what is already there rather than receiving data of its own.

These field names are Cress's theme API. They are fixed for the life of a major
version: a theme written against them today keeps working through all of 1.x.

### `.Site`

The `[site]` table from `cress.toml`. Every field may be empty unless noted, so
guard each one:

| Field | Holds |
|---|---|
| `.Site.Title` | The site name. |
| `.Site.Description` | The site tagline. |
| `.Site.BaseURL` | The canonical root, empty when unset. |
| `.Site.BasePath` | The path the site is published under. See below. |
| `.Site.Logo` | A logo URL, already rooted. |
| `.Site.LogoDark` | A dark-scheme logo variant. |
| `.Site.Favicon` | A favicon URL, already rooted. |
| `.Site.Accent` | An accent color, as the author wrote it. |
| `.Site.AccentDark` | A dark-scheme accent. |
| `.Site.Footer` | A footer line. |
| `.Site.Copyright` | A copyright line, with `{year}` already expanded. |

### `.Nav`

The resolved navigation, one list per menu: `.Nav.Main` from `[nav.main]` and
`.Nav.Footer` from `[nav.footer]`, each in the order the author wrote it. Each
entry carries `.Title`, `.URL`, and `.Active`, which is true when the entry
points at the page being rendered.

A menu the site did not configure is empty, so `{{ with .Nav.Footer }}` skips
the whole block and leaves no empty landmark behind:

```html
{{- with .Nav.Main }}
<nav>
  {{- range . }}
  <a href="{{ .URL }}"{{ if .Active }} aria-current="page"{{ end }}>{{ .Title }}</a>
  {{- end }}
</nav>
{{- end }}
```

An entry whose target does not exist is dropped with a build warning rather than
rendered as a dead link, so `.URL` is never empty.

### `.Page`

The page being rendered:

| Field | Holds |
|---|---|
| `.Page.Title` | The page title, as plain text. |
| `.Page.HTML` | The rendered Markdown body. Already markup. |
| `.Page.Head` | Metadata elements for the document head. Already markup. |
| `.Page.Description` | The page description, falling back to the site's. |
| `.Page.Language` | The language tag. Never empty. |
| `.Page.Image` | The page's lead image, falling back to the site's. Empty when neither is set. |
| `.Page.ImageAlt` | Describes that image. Empty means the author left it decorative. |
| `.Page.ImageCaption` | The line to render under that image. Empty when there is none. |
| `.Page.URL` | This page's root-relative URL. |
| `.Page.AbsoluteURL` | The full URL, empty when the site sets no `base_url`. |
| `.Page.IsHome` | True only for the page at the site root. |
| `.Page.Meta` | The raw front matter, including keys Cress does not define. |

Five of those are easy to get wrong:

- **`.Page.HTML` is markup, not text.** Emit it with `{{ .Page.HTML }}` and do
  not escape it again. Escaping it a second time shows readers their own page
  source.
- **`.Page.Head` is Cress's half of the document head.** It carries the meta
  description and the canonical link, each omitted when it has no value. Emit it
  and let Cress own what goes in it. The title, charset, viewport, icon, and
  stylesheets stay yours. Rendering it *and* `.Page.Description` emits every tag
  twice, so pick one.
- **`.Page.Language` belongs in an attribute**, not in the head, which is why
  `.Page.Head` cannot carry it: `<html lang="{{ .Page.Language }}">`.
- **`.Page.IsHome`** is for what reads differently on a front page, such as
  dropping the site name from a title that already is the site name.
- **`.Page.Image` is yours to place, and it does not travel alone.** Cress
  resolves which image a page gets and roots the path; where it goes and how big
  it is are the theme's. The alt text and the caption belong to it:

  ```html
  {{ with .Page.Image }}
  <figure>
    <img src="{{ . }}" alt="{{ $.Page.ImageAlt }}" />
    {{- with $.Page.ImageCaption }}
    <figcaption>{{ . }}</figcaption>
    {{- end }}
  </figure>
  {{ end }}
  ```

  Guard the image with `with`: it is empty whenever neither the page nor the
  site names one. Reach the other two through `$`, since `with` has moved the
  dot to the URL. Emit `alt` even when `ImageAlt` is empty, because an empty
  `alt` marks an image as decorative while a missing one leaves a screen reader
  reading out the file name.

  **Render the caption if you render the image.** It is where an attribution, a
  license, or an AI disclosure goes, and those are things a site may be legally
  obliged to show. A theme that displays the picture and drops the caption
  leaves its users no way to comply.

`.Page.Meta` is the escape hatch. Any front-matter key Cress does not define
reaches the template through it, so a theme can invent its own without asking
for a config change.

## URLs, and the one thing you must root yourself

Every URL in `PageData` is already correct. Navigation entries, page URLs, the
logo, and the favicon are all rooted under the path the site is published at, so
a template emits them as they are.

Your own asset links are the exception, because Cress does not know about them:

```html
<link rel="stylesheet" href="{{ .Site.BasePath }}/style.css" />
```

`.Site.BasePath` is empty for a site at a domain root, which is why the form
above is also correct there. A theme that hardcodes `/style.css` works at a
domain root and renders **silently unstyled** under a subdirectory, with nothing
to say which of the two is at fault. Cress cannot detect it, because it cannot
tell which URLs a theme meant as its own.

Use the same form for a link to the site's home page.

## What Cress emits

The HTML a theme wraps is plain and semantic. Two shapes are worth knowing
before you write CSS:

- **A standalone image arrives wrapped in `<figure>`.** An image that is the
  whole of its paragraph gets a `<figure>`, with the image title as a
  `<figcaption>` when the author wrote one. The wrapper is there with or without
  a caption, so you style one shape rather than two. A theme that styles
  `p img` and not `figure img` misses every standalone image on the site.
- **Code fences carry a language class.** A fenced block renders as
  `<pre><code class="language-go">`, with no styling and, by default, no tokens
  inside it. The class is there so a theme can target it. A site can ask for the
  tokens; see [Coloring code](#coloring-code) below.

### Coloring code

A site that sets `highlight = true` in the `[markdown]` table of its
`cress.toml` gets each code block read in the language its fence names. The
wrapper is the markup you already style, with one addition:

```html
<pre class="chroma"><code class="language-go">
```

Inside it, every keyword, string, and comment is wrapped in a span of its own,
and each line in `<span class="line"><span class="cl">`. A line of Go comes out
like this:

```html
<span class="kd">func</span><span class="w"> </span><span class="nf">main</span><span class="p">()</span>
```

The spans carry a class and never a color, so the colors are yours. A fence
naming no language, or one Cress cannot read, is emitted unchanged and carries
none of this.

Style the classes, not the languages. They come from
[chroma](https://github.com/alecthomas/chroma), the tokenizer Cress uses, and
these are the ones worth covering:

| Class            | Token                                   |
|------------------|-----------------------------------------|
| `c` `c1` `cm`    | comments                                |
| `k` `kd` `kn`    | keywords                                |
| `kt` `nc` `nn`   | types, classes, namespaces              |
| `nf` `nb` `nt`   | functions, builtins, tags               |
| `s` `s1` `s2`    | strings                                 |
| `m` `mi` `mf`    | numbers                                 |
| `o` `p`          | operators and punctuation               |
| `gd` `gi`        | removed and added lines in a diff       |
| `err`            | text the tokenizer could not classify   |

Scope the rules under `.chroma` rather than under your own wrapper class. That
is chroma's own convention, so a stylesheet chroma generates applies to a Cress
site unchanged.

Put them in their own `static/highlight.css` rather than in `style.css`, and
link it from your head partial:

```html
<link rel="stylesheet" href="{{ .Site.BasePath }}/style.css" />
<link rel="stylesheet" href="{{ .Site.BasePath }}/highlight.css" />
```

That is what all three built-in themes do. A separate file is a color scheme
somebody can replace with one file and no edits to your theme, which is the
whole reason to keep it out of `style.css`. Make it self-contained for the same
reason: declare the colors it needs in the file itself, because a token
stylesheet that reads variables from `style.css` breaks the moment it is
swapped for one that does not set them.

Link it unconditionally. A template cannot read the `[markdown]` table, so there
is nothing to branch on, and there is nothing to gain either: with the key off
no token spans are emitted, so the file matches nothing.

Two things are worth getting right. Check every token color against your code
background for a contrast ratio of at least 4.5:1, because a token is code
rather than decoration and a comment nobody can read is worse than a comment
with no color at all. Give the dark scheme its own set: a palette legible on
paper rarely survives being put on a dark panel.

Shipping no token styles at all is a valid choice. The spans are inert without
CSS, so the code renders as it does with highlighting off.

## The 404 page

Every build writes a `404.html` at the site root, and a theme does not have to
do anything for that to work: Cress renders it through `page.html`.

Add `templates/404.html` to give it its own shape. It receives the same
`PageData` every other template does. A page the site author wrote at
`content/404.md` takes precedence over both.

## Previewing while you work

```bash
cress serve
```

That builds the site, serves it, and rebuilds when a file changes. Editing a
template or a stylesheet in your theme triggers a rebuild, so the loop is save
and refresh.

Two things to know. A missing address is answered with the site's 404, so you
can see that page without navigating to it directly. And a build that warns
still serves: read what it prints rather than only looking at the page.

## Checking your theme

Previewing shows you the pages your content happens to reach. To check the
whole theme, including the parts no page reaches:

```bash
cress theme validate mine
```

```console
themes/mine: 2 problem(s)
  templates/page.html:14  .Page.Titel is not a field of PageView (did you mean Title?)
  templates/wide.html:3   href="/style.css" is not rooted under {{ .Site.BasePath }}, so it breaks when the site is published to a subdirectory
```

With no name it checks the theme your `cress.toml` already points at. It exits
non-zero when it finds anything, so it belongs in your own pipeline, and
`--quiet` prints nothing and leaves only the exit status.

What it checks:

- **Fields your templates read.** Every one, against the contract above.
- **`templates/page.html` exists, and every template parses.**
- **Your own asset links carry `.Site.BasePath`.**
- **Two files do not define the same `{{ define }}` name.**
- **A file in `templates/` renders something of its own.**
- **`theme.toml` parses**, and its `cress` version matches the Cress running.

Coverage is the reason to run it. A build only executes the templates and
branches some page reaches, so a bad field in a layout no page names, or behind
an `{{ if }}` no page takes, builds perfectly cleanly and breaks the day
somebody writes the page that reaches it. Validating reads every branch of every
template and needs no content at all.

It is deliberately quiet where it cannot be certain. A fragment reached with
something other than the whole `PageData` is checked against whatever its
callers pass it, and one reached with two different things is not checked at
all. Anything it does report is something to fix.

## Gotchas

- **A template that reads a field that does not exist fails the build.** Go
  templates error on an unknown struct field, so `.Page.Titel` is caught, but
  only once a page renders that template and takes that branch.
  `cress theme validate` catches it without either. A misspelled *layout* name
  is not caught at all: it warns and falls back.
- **Drafts are not rendered** unless the build passes `--drafts`, so a theme
  never sees them.
- **`static/` in the site beats `static/` in the theme** on a name collision.
  That is how a site overrides a theme's asset, and it means a theme cannot rely
  on owning a common name like `style.css` if the site ships one too.
- **Two partials with the same `{{ define }}` name** resolve to whichever parsed
  last. No build warns about it, so keep the names distinct;
  `cress theme validate` is what finds one.
- **Indentation inside a partial is emitted verbatim.** Go templates do not
  re-indent a fragment to its call site.

## Before you publish a theme

- `cress theme validate` reports no problems.
- `cress build` runs clean, with no warnings, against a scaffolded site.
- The site renders correctly under a subdirectory: set `base_url` to something
  like `https://example.com/sub`, rebuild, and confirm the stylesheet still
  loads.
- Both color schemes are legible, including with `accent` and `accent_dark`
  set to something other than your defaults.
- The site is usable with no `logo`, no `favicon`, no `[nav.footer]`, no
  `footer`, and no `copyright`. Every one of those is optional.
- A page naming a layout your theme does not define still renders.
- `theme.toml` names you, the license, and the `cress` version you developed
  against.
- Document the layouts your theme offers, since Cress cannot.
