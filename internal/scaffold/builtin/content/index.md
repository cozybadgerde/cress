---
title: Welcome
---

# Welcome

This is your new **Cress** site. Fresh little sites, fast.

Edit `content/index.md` to change this page, add more Markdown files under
`content/`, and list them in `cress.toml` to grow the navigation.

## Getting around

- Every `.md` file under `content/` becomes a page: `content/about.md` renders
  to `about.html`.
- `content/index.md` is the home page.
- Files in `static/` are copied to the root of the site, so `static/logo.svg`
  is linked as `/logo.svg`.
- Styling lives entirely in the theme, so your Markdown stays clean.

Run `cress serve` to preview locally, and `cress build` when you are ready to
publish the `public/` directory.
