---
title: Writing well
image: /example_minimal.webp
image_alt: A pot of cress on a sunny windowsill
image_caption: AI-generated illustration.
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

![Woodland friends tending trays of cress](/example_busy.webp "AI-generated illustration.")

Then write the alt text. It is not a caption and not a place for keywords. It
is what somebody gets *instead of* the picture, so describe what matters about
this image in this spot. If it is purely decorative, empty alt text
(`![](/example_busy.webp)`) is the honest answer and better than repeating the
sentence above it.

## Caption the things a reader should see

The line under the picture above is a caption, written as a quoted title after
the path:

```markdown
![Woodland friends tending trays of cress](/example_busy.webp "AI-generated illustration.")
```

A caption does a different job from alt text. Alt text stands in for the
picture, so the only person who meets it is somebody who cannot see the image.
A caption is text everybody reads. That makes it the place for whatever the
picture should say out loud: who made it, what license it carries, or that a
machine generated it. Both illustrations in this starter site are AI-generated,
and their captions say so.

Leave the title off and you get no caption, just the picture. Cress wraps a
standalone image either way, so a theme has one shape to style.

## Give the page a lead image

The picture at the top of this page is not in its Markdown at all. It is named
in the front matter:

```markdown
---
title: Writing well
image: /example_minimal.webp
image_alt: A pot of cress on a sunny windowsill
image_caption: AI-generated illustration.
---
```

A lead image is the theme's to place: where it goes and how large it is are the
theme's decisions, not yours. What it says is still yours, which is why the
other two keys are there. `image_alt` is the description somebody gets instead
of the picture, and `image_caption` is the line under it that everybody reads.

Use the caption for anything the image is obliged to say out loud: who took it,
what license it carries, or that a machine made it. The rules about that vary by
where you are and Cress knows none of them, so it gives you the place to put the
statement and leaves the wording to you. Leave the key out and no caption is
rendered.

Set `image` under `[site]` in `cress.toml` to give every page the same one, with
its own `image_alt` and `image_caption`. A page that names its own `image` takes
its own alt and caption too, and never the site's: a credit belongs to the
picture it was written for.

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
