+++
# Frontmatter is TOML between +++ lines; it configures this page, not the site.
# A title is required for published pages; it does not inherit from cress.toml.
title = "Content tour"
# Omit to inherit site.description, or use "" to clear it for this page.
description = "A small style guide for writing and previewing Cress content."
# Page types do not inherit; "ordinary" is the default.
type = "ordinary" # Also "landing" or "list"; see the home and journal pages.
draft = false # True hides this page from builds; preview it with --drafts.
# Optional calendar dates describe publication and updates, not scheduling.
published_at = 2026-10-21
updated_at = 2026-10-23
# Optional plain text for lists; not extracted from the body or description.
excerpt = "Explore headings, links, images, tables, code, and captions."

# An image replaces the whole site default; image = { src = "" } clears it.
# Put that clearing example above [image] if you use it instead of this table.
[image]
src = "index.assets/placeholder.svg"
alt = "Placeholder for a photo"
caption = "A placeholder image from the [home page](index.md)."
+++

## A guide you can edit

Open `content/tour.md` beside its preview to compare source and result.
This page demonstrates CommonMark, tables, and Cress captions. Keep it as
a reference, or remove it and its footer link before publishing.

The frontmatter above sets this page's title, description, dates, and title
image. The theme supplies the page's main heading, so start body headings
at level two. Ordinary page titles are required; descriptions and images
inherit from `cress.toml` when omitted.

## Text and emphasis

Write paragraphs separated by a blank line. Use *emphasis*, **strong text**,
or ***both*** to draw attention to a few words. Inline code such as
`cress site build` stays literal, including `<example>` tags.

A single source newline
continues the same paragraph. A backslash ends this line explicitly:\
This line starts after a hard break.

Escape punctuation to show \*literal asterisks\*. Character references
such as &amp; and &copy; render as their characters.

## Heading levels

### A section within a section

Use a clear hierarchy rather than choosing headings for their size.

#### A smaller topic

Keep related information together.

##### A detail

Most business pages need only two or three heading levels.

###### The smallest heading

All six HTML heading levels are available, including the page title.

## Links

Link to [our services](services.md), [the home page](index.md), or
[a public website](https://example.org/). Source-relative Markdown links
are resolved to the published page URL.

[Reference-style links][contact] keep repeated destinations in one place.
Angle brackets make explicit automatic links: <https://example.org/> or
<hello@example.org>.

[contact]: contact.md "Contact our team"

## Lists

- A short service description.
- A second item with **important details**.
  - A nested point.
  - Another supporting detail.

1. Tell us what you need.
2. Agree on the next step.
3. Review the result together.

## Quotations and separators

> Good work starts with a useful conversation.
>
> A quotation can contain several paragraphs.
>
> - It can also contain a list.

A horizontal rule separates related sections:

---

## Tables

Tables extend CommonMark. Colons choose left, center, or right alignment.

| Service | Availability | Example price |
|:--------|:------------:|--------------:|
| First conversation | Weekdays | EUR 40 |
| Planning session | By arrangement | EUR 90 |

## Images and captions

Use meaningful alternative text; captions are separate visible context.
This ordinary image has a hover title but no caption:

![Placeholder for a photo](index.assets/placeholder.svg "A placeholder")

Add an immediate caption attribute to create a figure:

![Placeholder for a photo](index.assets/placeholder.svg){caption="A local image with a [link to its page](index.md)."}

Inline images work too:
![Placeholder](index.assets/placeholder.svg){caption="A small inline image."}
stays within this paragraph.

The title image uses the frontmatter `[image]` table. Captions support
plain text and links, not emphasis or raw HTML.

## Code

Fenced code stays literal. A recognized language enables build-time
highlighting, and an optional caption explains the example:

```go caption="A greeting, highlighted without browser scripts."
package main

import "fmt"

func main() {
    fmt.Println("Hello, Cress")
}
```

An untagged fence stays plain:

```
<p>This is code, not an HTML paragraph.</p>
```

Indenting by four spaces also creates a plain code block:

    cress site build
    cress site preview --port 0

## Authored HTML

Raw HTML passes through unchanged. For example, an inline
<abbr title="Frequently asked questions">FAQ</abbr> can add meaning.

<div>
<p>This paragraph is an authored HTML block.</p>
</div>

Use HTML only from sources you trust. Cress does not sanitize it or rewrite
its URLs; browsers may load resources or execute scripts supplied in HTML.

## What is not enabled

This tour uses supported features, not every Markdown extension.
Task-list checkboxes, strikethrough, footnotes, and automatic links from
bare URLs are not enabled. Use ordinary lists, emphasis, and explicit links.
