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

That writes `themes/mine/` with two layouts, the partials they share, and a
stylesheet. Every file is commented, so the starter is a working tour of what
this guide describes. Name it in `cress.toml` and preview it:

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

A directory under `themes/` wins over the built-in theme of the same name. Any
other name with no matching directory is an error, so a typo in `theme =` fails
the build rather than silently falling back.

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
| `.Page.URL` | This page's root-relative URL. |
| `.Page.AbsoluteURL` | The full URL, empty when the site sets no `base_url`. |
| `.Page.IsHome` | True only for the page at the site root. |
| `.Page.Meta` | The raw front matter, including keys Cress does not define. |

Four of those are easy to get wrong:

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
- **Code fences carry a language class and nothing else.** A fenced block
  renders as `<pre><code class="language-go">`, with no styling and no tokens
  inside it. The class is there so a theme can target it.

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

## Gotchas

- **A template that reads a field that does not exist fails the build.** Go
  templates error on an unknown struct field, so `.Page.Titel` is caught. A
  misspelled *layout* name is not: it warns and falls back.
- **Drafts are not rendered** unless the build passes `--drafts`, so a theme
  never sees them.
- **`static/` in the site beats `static/` in the theme** on a name collision.
  That is how a site overrides a theme's asset, and it means a theme cannot rely
  on owning a common name like `style.css` if the site ships one too.
- **Two partials with the same `{{ define }}` name** resolve to whichever parsed
  last. Nothing warns, so keep the names distinct.
- **Indentation inside a partial is emitted verbatim.** Go templates do not
  re-indent a fragment to its call site.

## Before you publish a theme

- `cress build` runs clean, with no warnings, against a scaffolded site.
- The site renders correctly under a subdirectory: set `base_url` to something
  like `https://example.com/sub`, rebuild, and confirm the stylesheet still
  loads.
- Both color schemes are legible, including with `accent` and `accent_dark`
  set to something other than your defaults.
- The site is usable with no `logo`, no `favicon`, no `[nav.footer]`, no
  `footer`, and no `copyright`. Every one of those is optional.
- A page naming a layout your theme does not define still renders.
- Document the layouts your theme offers, since Cress cannot.
