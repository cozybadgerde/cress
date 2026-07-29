---
title: Guides
---

# Guides

This directory is itself a lesson. Any folder under `content/` may hold an
`index.md`, and that file serves the folder's own URL: `content/guides/index.md`
is the page you are reading, at `/guides/`. The root `content/index.md` is not a
special case, it is the same rule applied to the top of the tree.

Everything else in a folder sits beneath it. `content/guides/writing.md` becomes
`/guides/writing.html`, one step down from here.

That is the whole of it. Group pages into a folder when they belong together,
give the folder an `index.md` so its URL leads somewhere, and add it to
`[nav.main]` in `cress.toml` by its content path:

```toml
[nav.main]
Guides = "guides/index.md"
```

## What is in here

- [Style guide](/guides/styleguide.html) shows every Markdown element Cress
  supports, rendered by the current theme. Reach for it when you want to know
  what you can write.
- [Writing well](/guides/writing.html) is about the judgement calls Markdown
  cannot make for you: how to name a file, where a heading belongs, what to do
  about an image.

Both are yours to delete once you have found your feet.
