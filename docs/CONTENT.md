# Content guide

Every page on your site is a Markdown file in `content/`. This guide shows
how to write pages that read well, look good in every theme, and work for
every visitor. Open `content/tour.md` next to its preview to see each
example in action.

## Start a page

Create a file ending in `.md` in `content/`. Its name becomes its address:

| File | Address |
|------|---------|
| `content/index.md` | `/` |
| `content/about.md` | `/about/` |
| `content/journal/first-visit.md` | `/journal/first-visit/` |

Begin the file with its settings between two `+++` lines, then write:

```markdown
+++
title = "About us"
description = "Who we are and how we work."
+++

We help small businesses build a useful presence online.
```

The theme shows the title as the page's main heading. A good description is
one plain sentence that tells search engines and visitors what the page is
about. Once `[site] host` is set, the title, description, and title image
also form the preview shown when someone shares a link to the page.

The settings at the top of `content/tour.md` explain every page setting in a
comment. The [frontmatter schema](../schemas/frontmatter.schema.json) lets
your editor check them for you.

## Structure with headings

Headings are the outline of your page. Many visitors scan them first, and
screen reader users jump between them.

- Start with `##`, because the title is already the top-level heading.
- Go down one level at a time: `###` belongs under `##`.
- Keep headings short and specific. "Opening hours" beats "Information".

Every heading gets a link target, so you can point straight at a section:

```markdown
See our [opening hours](#opening-hours), or the
[team's hours](contact.md#opening-hours).
```

Cress builds the target from the heading text: lowercase letters and digits
stay, spaces become hyphens, and everything else is dropped. Letters such as
`Ü` disappear too, so `## Über uns` becomes `ber-uns`. If that looks odd, or
you want links to survive a reworded heading, set the target yourself:

```markdown
## When to visit {#opening-hours}
```

The names `main`, `page-title`, and anything starting with `cress-` belong
to themes, so pick something else.

## Write clear text

Separate paragraphs with a blank line. Keep them short: one idea each.

- Use `*emphasis*` and `**strong text**` sparingly. If everything stands out,
  nothing does.
- Use `` `inline code` `` for commands, file names, and anything a reader
  should type exactly.
- Use `> ` for quotations and `---` on its own line to separate topics.

To show that something changed, strike the old text with `~~old~~`, and say
what changed in words. Screen readers often skip strikethrough:

```markdown
Our workshop moves to Friday. ~~Saturday~~
```

## Link with care

Link text should make sense on its own, because people scan links and
screen readers can list them out of context. "Read our
[pricing](services.md)" is better than "click [here](services.md)".

Link to other pages by their file, relative to the current one:

```markdown
[Our services](services.md)
[First visit](journal/first-visit.md)
```

Cress turns these into the right addresses and fails the build if the file
is missing or still a draft, so broken links never reach your visitors.
Links to other websites are left alone and not checked.

## Use lists and tables

Use bullet lists for things that belong together and numbered lists for
steps. Indent by two spaces to nest a list.

Task lists show progress, such as a project plan:

```markdown
- [x] Agree the project scope.
- [ ] Review the [services](services.md).
```

Visitors see checkboxes but can't tick them. Edit the Markdown to update
them.

Use tables for real comparisons, like prices or opening hours, not for
layout. Keep them narrow, and give every column a header:

```markdown
| Day | Hours |
|-----|-------|
| Monday to Friday | 9:00 to 17:00 |
| Saturday | 10:00 to 14:00 |
```

Wide tables scroll sideways on small screens.

## Add images

Keep a page's images in a folder named after the page: `about.md` uses
`about.assets/`. Reference them relative to the page:

```markdown
![Our team in front of the workshop](about.assets/team.jpg)
```

Put files you use on many pages, like your logo, in `static/` and reference
them from the site root, such as `/images/logo.svg`.

The text in brackets is the image's description for people who can't see
it. Describe what matters in this context: "Our team in front of the
workshop" says more than "photo". Leave it empty only for purely decorative
images.

A caption adds visible context, such as a credit. It can hold plain text and
links:

```markdown
![Our team](about.assets/team.jpg){caption="Photo by [Alex](https://example.org/)"}
```

A caption doesn't replace the description; use both.

The large image at the top of a page is a page setting. The tour shows how
to set it. Use a JPEG or PNG title image for link previews; most services
don't show SVG images.

## Show code

Put code in a fenced block and name its language. Cress highlights it during
the build:

````markdown
```sh
cress site build
```
````

Code blocks can have captions too:

````markdown
```go caption="A greeting in Go"
fmt.Println("Hello")
```
````

## Use HTML when Markdown isn't enough

You can write HTML directly in your pages, for example
`<abbr title="Frequently asked questions">FAQ</abbr>`. Cress publishes it
exactly as written, and visitors' browsers run whatever it contains. Use
HTML you understand, and review anything copied from elsewhere before you
build or preview it.

## Collect pages in a list

A list page shows the pages in its folder as linked entries, which suits a
journal or a portfolio. The sample site's `content/journal/` shows how.

Give each entry a `published_at` date and a short `excerpt`. By default,
lists show the newest page first and leave out pages without a date, so a
missing date is the most common reason a page doesn't show up. Cress warns
you when that happens.

## Work on drafts

Add `draft = true` to a page's settings to keep it out of the build. Drafts and
their images stay private, and links to a draft fail the build until you
publish it. Preview drafts with `cress site preview --drafts`.
