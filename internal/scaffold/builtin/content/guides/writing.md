---
title: Writing well
---

# Writing well

Markdown lets you write anything, and Cress builds whatever you write. That
leaves a handful of small decisions to you. They are the ones that decide
whether a site is still pleasant to read a year from now.

Here is one opinionated pass through them. Disagree freely, then delete this
page.

## Name the file before you write it

The filename becomes the URL. `content/guides/writing.md` is served at
`/guides/writing.html`, and that address is what ends up in somebody's
bookmarks, in a chat message, in a search result.

So: lowercase, hyphens between words, no spaces. `first-week.md` will still
make sense in March. `Notes (final) v2.md` will not, and it makes an awkward
URL besides.

Renaming a page later breaks every link that pointed at the old name. Cress
will not warn you, and neither will the people who quietly find a dead end.
That is a reason to choose the name carefully on the first day, not a reason
to never rename.

## Let headings carry the outline

One `#` per page, at the top. Cress reads that first heading as the page title
when the front matter does not give one, which makes it structural rather than
decorative.

Everything below it is `##` and `###`, in order, without skipping a level. A
heading is a claim about structure: it says a new part starts here. If you
reach for one because the text needs a visual break, what you actually want is
a paragraph, a list, or a horizontal rule.

## Link inward with a leading slash

Internal links start at the root: `/about.html`, `/guides/styleguide.html`.
Written that way they work from every page, whatever folder it sits in.

Write link text that says where it goes. "See [the style
guide](/guides/styleguide.html)" tells someone scanning the page what they will
get. "See [here](/guides/styleguide.html)" tells them nothing, and tells a
screen reader even less.

## Choose an image, then shrink it

Cress has no asset pipeline. Whatever you put in `static/` is what your readers
download, at exactly the size you committed it. A 4 MB photograph stays a 4 MB
photograph.

Pick the format to suit the artwork:

- **WebP or JPEG** for photographs and anything shaded.
- **SVG** for line art, diagrams, and logos, because it stays sharp at any size.
- **PNG** when you need real transparency or a lossless original.

![Woodland friends tending trays of cress](/example_busy.webp)

Then write the alt text. It is not a caption and not a place for keywords. It
is what somebody gets *instead of* the picture, so describe what matters about
this image in this spot. If it is purely decorative, empty alt text
(`![](/example_busy.webp)`) is the honest answer and better than repeating the
sentence above it.

## Publish when it is ready

`draft: true` in the front matter keeps a page out of the build. No file is
written for it, so nothing goes out early. Use `cress build --drafts` when you
want to see one in context.

What it does not do is retract anything. A page that has already been published
leaves its HTML in `public/` even after you delete the Markdown, because a Cress
build never removes files. Drafts are for pages that have not gone out yet, not
a way to pull back one that has.

---

None of this is enforced. Cress builds a site with fifteen H1s and a 12 MB PNG
just as cheerfully as it builds a careful one. The care is the part only you
can add.
