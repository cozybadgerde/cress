---
title: Style guide
---

# Style guide

This page shows every Markdown element Cress supports, rendered by the current
theme. It has two jobs. If you are writing content, it is a menu of what you can
reach for. If you are building a theme, it is a checklist of what to style.

Keep it as a reference, or delete it once you have found your feet.

## Headings

The page title above is the only level-one heading. Use `##` down to `######`
for the structure inside a page.

## Heading level two

### Heading level three

#### Heading level four

##### Heading level five

###### Heading level six

## Text

Regular paragraphs need no ceremony. You can make text **bold**, _italic_,
**_both at once_**, or ~~struck through~~. Inline `code` sits between backticks,
and [links](https://github.com/cozybadgerde/cress) point wherever you like.
Bare URLs become links on their own, like https://cozybadger.de.

> Blockquotes are good for asides and pull quotes.
>
> They can run to several paragraphs.

## Lists

Unordered lists use dashes:

- Cress sprouts
- A curious mouse
- A cup of tea

Ordered lists count:

1. Write some Markdown
2. Run `cress build`
3. Ship a fresh little site

Lists nest, and task lists show progress:

- Content
  - Pages
  - Navigation
- Theme
  - [x] Templates
  - [x] Styles
  - [ ] A dark-mode tweak

## Code

Inline `code` is handy mid-sentence. For longer snippets, fence a block and tag
it with a language so themes can style it:

```go
package main

import "fmt"

func main() {
	fmt.Println("Fresh little sites, fast.")
}
```

```css
.site-title {
  color: var(--accent);
}
```

## Tables

Columns can be left, center, or right aligned:

| Feature       | Markdown     | Aligned |
| :------------ | :----------: | ------: |
| Bold          | `**text**`   |     yes |
| Inline code   | `` `code` `` |     yes |
| Strikethrough | `~~text~~`   |     yes |

## Images

Drop a file in `static/` and link it with a root-relative path. Everything in
`static/` is copied straight into the root of the built site, so
`static/example_minimal.webp` is served from `/example_minimal.webp`. It is the
same directory the `logo` and `favicon` keys in `cress.toml` point into.

Start the path with `/`. This page lives at `/guides/styleguide.html`, and the
image below still resolves, because the link does not depend on where the page
sits. A relative `./example_minimal.webp` would have looked inside `/guides/`
and found nothing.

![A pot of cress on a sunny windowsill](/example_minimal.webp "AI-generated illustration.")

Write alt text that describes the picture for anyone who cannot see it. Add a
quoted title after the path and Cress turns it into a visible caption, the way
the line under the picture above is. [Writing well](/guides/writing.html) has
more on both, and on picking a format.

## Dividers

Three dashes on their own line make a horizontal rule:

---

## The theme decides how this looks

Everything above is plain, semantic HTML. What it _looks_ like is the theme's
call, and Cress ships three:

| Theme   | Feels like                                              |
|---------|---------------------------------------------------------|
| `cress` | the default: warm paper, a serif body, quiet and roomy   |
| `birch` | a developer tool: crisp, sans-serif, built around code   |
| `poppy` | a hand-made book: photos taped in, quotes on torn scraps |

Change one line in `cress.toml` and reload this page:

```toml
[site]
theme = "poppy"
```

Every heading, quote, table and code block on this page is drawn differently by
each of them, and none of your Markdown changes. That is the whole idea: you
write the words, the theme decides the look. Write your own with
`cress theme init`.

That is the whole toolbox. Happy writing.
