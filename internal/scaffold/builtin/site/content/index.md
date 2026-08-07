---
title: Welcome
layout: landing
---

# Welcome

This is your new **Cress** site. Fresh little sites, fast.

Edit `content/index.md` to change this page, add more Markdown files under
`content/`, and list them in `cress.toml` to grow the navigation.

## Getting around

- Every `.md` file under `content/` becomes a page: `content/about.md` renders
  to `about.html`.
- Folders nest as you would expect: `content/guides/writing.md` becomes
  `/guides/writing.html`.
- An `index.md` serves the URL of the folder it sits in. That makes
  `content/index.md` this page, and `content/guides/index.md` the page at
  `/guides/`.
- Files in `static/` are copied to the root of the site, so `static/logo.svg`
  is linked as `/logo.svg`.
- Styling lives entirely in the theme, so your Markdown stays clean.

## Why this page looks different

Its front matter says `layout: landing`, so the built-in theme renders it
through `templates/landing.html` instead of the usual `templates/page.html`:
the page opens with a panel filling the first screen, and this text follows
below it. Every other page here leaves `layout` unset and gets the standard
layout.

Layout names belong to the theme rather than to Cress. The built-in theme
offers `landing`; another theme offers whatever it documents, and a name no
theme defines falls back to the standard layout with a warning at build time.

Run `cress serve` to preview locally, and `cress build` when you are ready to
publish the `public/` directory.
