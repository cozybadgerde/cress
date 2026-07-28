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

Three optional fields make the site your own:

- `logo`: a path or URL to a logo image, shown in the navigation. `cress init`
  drops a placeholder at `static/logo.png`; replace it with your own, or point
  this field at another file in `static/`.
- `favicon`: a path or URL to a favicon. `cress init` drops a placeholder at
  `static/favicon.png`; replace it the same way.
- `accent`: the accent color, a soft highlight the theme uses on links and other
  touches. A CSS hex value, for example `accent = "#9cb43b"` (the default).

```toml
[site]
logo = "/logo.png"
favicon = "/favicon.png"
accent = "#9cb43b"
```

Leaving `logo` or `favicon` empty simply omits it. An invalid `accent` (anything
that is not a hex color like `#9cb43b`) fails the build with a clear message.

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
    static/          optional: CSS and other assets, copied to the site root
```

```toml
[site]
theme = "mine"
```

Templates use Go's `html/template`. The `page.html` template receives:

- `.Site`: the `[site]` config (`.Site.Title`, `.Site.Description`,
  `.Site.BaseURL`, `.Site.Logo`, `.Site.Favicon`, `.Site.Accent`).
- `.Nav`: the resolved navigation, as `.Nav.Main` and `.Nav.Footer`. Each is a
  list of entries with `.Title`, `.URL`, and `.Active` (true on the current
  page). A group with no entries is empty, so `{{ with .Nav.Footer }}` skips it.
- `.Page`: the current page, with `.Page.Title`, `.Page.URL`, `.Page.HTML` (the
  rendered Markdown body), and `.Page.Meta` (the raw front matter).

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
