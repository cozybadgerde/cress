# User guide

This guide explains how to author a Cress site: the directory layout, the config
file, content and front matter, navigation, and themes.

## Before you start

Install Cress and create a starter site with the [operator guide](./OPERATOR.md).
The command `cress init my-site` writes everything described below into a new
directory.

## Site layout

A Cress site is a directory with this shape:

```text
my-site/
  cress.toml        config: site metadata and navigation
  content/          Markdown pages
  static/           files copied verbatim to the site root
  themes/           optional: your own themes (one directory each)
  public/           build output (created by `cress build`)
```

Only `cress.toml` and `content/` are required. The built-in theme is used when
`themes/` is absent.

## Configuration

`cress.toml` holds the site metadata and the navigation. A minimal file:

```toml
[site]
title = "My cozy site"
description = "A fresh little site made with Cress."
base_url = "https://example.com"
theme = "cress"

[nav.main]
Home = "index.md"
About = "about.md"
```

The `[site]` fields:

- `title`: the site name, shown in the header and the browser tab.
- `description`: an optional tagline, available to the theme.
- `base_url`: the canonical site root. In-site links stay root-relative, so this
  is used for absolute links only.
- `theme`: the theme name. Empty resolves to the built-in `cress` theme.

Unknown keys are rejected, so a typo fails the build instead of being ignored.

### White-label branding

Five optional fields make the site your own:

- `logo`: a path or URL to a logo image, shown in the navigation. `cress init`
  drops a placeholder at `static/logo.svg`; replace it with your own, or point
  this field at another file in `static/`. The placeholder is a mid-green SVG,
  a tone chosen to hold up against both the light and the dark background.
- `logo_dark`: a second image, used instead of `logo` for readers whose system
  prefers a dark color scheme. Leave it out and `logo` is used for both.
- `favicon`: a path or URL to a favicon. `cress init` drops a placeholder at
  `static/favicon.png`; replace it the same way.
- `accent`: the accent color, a highlight the theme uses on links and other
  touches. A CSS hex value, for example `accent = "#4f7a4a"`.
- `accent_dark`: the accent for readers whose system prefers a dark color
  scheme. Leave it out and `accent` applies to both.

```toml
[site]
logo = "/logo.svg"
logo_dark = "/logo-dark.svg"
favicon = "/favicon.png"
accent = "#4f7a4a"
accent_dark = "#9ccb8f"
```

Leaving `logo` or `favicon` empty simply omits it. An accent that is not a hex
color fails the build with a clear message, and so does a `logo_dark` with no
`logo` beside it: the dark variant replaces the light one and has nothing to
replace on its own.

**Most logos do not need `logo_dark`.** Artwork in a mid-tone reads against
both backgrounds, which is how the scaffolded placeholder gets away with a
single fixed color, and an SVG drawn with `currentColor` adapts on its own.
Reach for the second file when the artwork can do neither: a raster logo, or a
wordmark fixed in near-black or near-white, which disappears into one of the
two backgrounds. Setting both makes the theme render a `<picture>`, so the
right one is chosen before anything is fetched, with no flash of the wrong
image and no JavaScript.

A theme decides whether to honor this. The built-in theme does; a custom theme
that renders a bare `<img>` ignores `logo_dark` without complaint.

**Leaving the accents out is a real choice, not a shortcut.** cress has no
default accent of its own; unset, the theme picks one, and the built-in theme
picks a different value per color scheme. That matters more than it sounds: no
single color reaches the WCAG AA contrast ratio of 4.5:1 against both a light
and a dark background, so a brand color that reads beautifully in one scheme
can be close to invisible in the other. Set `accent` alone only if your color
survives both; otherwise set `accent_dark` as well.

### Footer text and copyright

Two more `[site]` fields fill the footer:

- `footer`: your own message, replacing the "Made with Cress" credit. Leave it
  out and the credit stays.
- `copyright`: a copyright notice. Leave it out and none is rendered.

```toml
[site]
footer = "A small corner of the internet."
copyright = "(c) {year} Cozy Badger"
```

In `copyright`, the token `{year}` becomes the year the site was built, so the
notice does not quietly go stale every January. Write the year yourself when you
want a fixed one or a range: `"(c) 2019-{year} Cozy Badger"` renders the start
year literally and expands only the end. Everything else in either field is
taken as written.

Both fields are **plain text**. Markup in them is escaped rather than rendered,
so `footer = "<b>bold</b>"` shows the angle brackets. That is deliberate: the
theme is the only place styling lives, and a config value that could inject
HTML into every page on the site would be a hole in that rule.

## Content and front matter

Every `.md` file under `content/` becomes a page. The path maps directly to the
output:

- `content/index.md` becomes the home page (`/`).
- `content/about.md` becomes `/about.html`.
- `content/guide/setup.md` becomes `/guide/setup.html`.
- `content/guide/index.md` becomes the directory page `/guide/`.

A file may start with a YAML front-matter block, fenced by `---`:

```markdown
---
title: About
draft: false
---

# About

Your content goes here.
```

Front-matter fields Cress understands:

- `title`: the page title. When absent, Cress uses the first `# ` heading, then
  the file name.
- `draft`: when `true`, the page is skipped unless you build with `--drafts`.

Any other field is passed through to the theme untouched, so a theme can read
its own keys. Unlike `cress.toml`, front matter is deliberately permissive: an
unknown key is never an error, so notes written in Obsidian or another editor
that adds front-matter fields of its own build without being edited first. The
trade-off is that a misspelled `title` is ignored rather than reported.

## Writing good content

Cress does not police what you write. None of the following is enforced, and
none of it is a rule the builder checks. They are the conventions that keep a
site readable as it grows, collected here because each one is easy to get wrong
once and then work around forever.

The starter site puts them into practice in `content/guides/writing.md`, which
you can read on your own site and then delete.

**File and folder names.** The name becomes the URL, so it outlives the file.
Use lowercase and hyphens, avoid spaces, and pick the name before you start
writing. Renaming later breaks every inbound link, and Cress does not warn you.

**Headings.** Use one `# ` per page. Cress falls back to it for the page title,
which makes the first heading structural rather than decorative. Below it, go
down through `##` and `###` in order without skipping a level: headings are the
page's outline, not a way to change type size.

**Links.** Write internal links root-relative, starting with `/`, so they
resolve identically from every page regardless of folder depth. Make the link
text describe the destination; "here" is useless to anyone scanning the page or
listening to it.

**Images.** They live in `static/` and are linked from the site root, as
described under "Site layout" above. Cress has no asset pipeline, so
whatever you commit is what visitors download, at full size: resize before
committing. Choose the format for the artwork, WebP or JPEG for photographs,
SVG for line art and diagrams, PNG when you need transparency or a lossless
original. Always write alt text, and write what somebody would need instead of
the image rather than a repeat of the caption.

**Front matter.** `title` and `draft` are the only keys Cress reads. Everything
else is passed to the theme, so treat front matter as a place to serve the
theme's needs, not a schema to satisfy.

**Drafts.** `draft: true` keeps a page out of the build. It does not retract a
page that has already been published, because a build never deletes: the old
HTML stays in the output until you remove it yourself.

## Navigation

Navigation is explicit. There are two menus, each a flat table under `[nav]`,
and each entry is a label on the left with the content file it links to on the
right:

```toml
[nav.main]
Home = "index.md"
Guide = "guide/index.md"
About = "about.md"

[nav.footer]
Imprint = "imprint.md"
Privacy = "privacy.md"
```

`[nav.main]` is the primary menu, which the theme renders in the header.
`[nav.footer]` is the secondary menu for links that belong out of the way:
legal pages, social links, and the like. Both are optional, and a theme that
only wants one menu can ignore the other.

Cress resolves each path to the page's URL, and each menu follows the order its
entries are written in. An entry pointing at a missing file is reported as a
warning naming its group and then dropped, so the rest of the menu still
renders. A label with spaces needs quoting, as TOML requires:
`"Blog Posts" = "blog.md"`.

Those two group names are the only ones cress knows. A mistyped `[nav.mian]` is
rejected as an unknown key rather than silently rendering nothing.

## Themes

A theme owns the templates and the styling. The built-in `cress` theme needs no
setup. To use your own, create a directory under `themes/` and name it in the
config:

```text
themes/
  mine/
    templates/
      page.html      required: renders one page
      404.html       optional: renders the 404 page
    static/          optional: CSS and other assets, copied to the site root
```

```toml
[site]
theme = "mine"
```

Templates use Go's `html/template`. The `page.html` template receives:

- `.Site`: the `[site]` config (`.Site.Title`, `.Site.Description`,
  `.Site.BaseURL`, `.Site.Logo`, `.Site.Favicon`, `.Site.Accent`,
  `.Site.AccentDark`, `.Site.Footer`, `.Site.Copyright`).
- `.Nav`: the resolved navigation, as `.Nav.Main` and `.Nav.Footer`. Each is a
  list of entries with `.Title`, `.URL`, and `.Active` (true on the current
  page). A group with no entries is empty, so `{{ with .Nav.Footer }}` skips it.
- `.Page`: the current page, with `.Page.Title`, `.Page.URL`, `.Page.HTML` (the
  rendered Markdown body), and `.Page.Meta` (the raw front matter).

## The 404 page

Every build writes a `404.html` at the root of `public/`. GitHub Pages, GitLab
Pages, Netlify, and most other static hosts serve that file for an address that
does not exist, so you get a "page not found" carrying your own navigation and
styling without doing anything.

Three sources can supply it, and the most specific one wins:

1. `content/404.md`, if you wrote one. It renders like any other page, so the
   words and the front matter are yours.
2. `templates/404.html` in the theme, if it defines one. It receives the same
   data `page.html` does, so a theme can give the 404 its own layout.
3. Otherwise cress writes one for you: a short "Page not found" with a link
   home, rendered through the theme's `page.html`.

Write `content/404.md` when you want your own wording. Nothing else needs
changing, and the page you write replaces the generated one:

```markdown
---
title: Lost
---

# Nothing here

That page moved or never existed. Try the [home page](/).
```

`cress serve` answers a missing address with the same file, so the preview shows
what a visitor to the published site sees. Some hosts need the 404 wired up in
their own configuration; check your host's documentation.

## Building and previewing

Preview locally while you write:

```bash
cress serve
```

This builds the site, serves it at `http://localhost:1313`, and rebuilds when a
source file changes. Reload the browser to see updates.

Render the final site into `public/`:

```bash
cress build
```

Add `--drafts` to either command to include pages marked `draft`.

`cress build` only ever writes its own files; it never deletes. So if you rename
or remove a page, its old HTML stays in `public/` and keeps being served. The
build names each such file in a warning, so you can delete the ones you no
longer want.
