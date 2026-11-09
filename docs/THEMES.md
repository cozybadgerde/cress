# Theme developer guide

A Cress theme decides how a site looks. Cress does the rest: it reads the
content, checks it, works out the pages, lists, and links, and hands each page
to your templates as ready-made data. Your theme only presents that data. It
never picks pages, sorts lists, or builds URLs.

This guide walks you through creating a theme and explains what you can rely
on. The three built-in themes, Poppy, Clover, and Anemone, are good examples
to learn from. They live in
[`internal/theme/builtin/`](../internal/theme/builtin/).

## Create your first theme

Start with a test site, which includes every page type and the content tour,
and add a theme to it:

```sh
cress site init theme-test
cress theme init my-theme theme-test
```

In `theme-test/cress.toml`, set the theme name:

```toml
[theme]
name = "my-theme"
```

Start the preview:

```sh
cress site preview --port 0 theme-test
```

Now edit the files in `theme-test/themes/my-theme/`. Preview rebuilds on every
save. Refresh the browser to see the result. Look at the landing page, the
journal list, an ordinary page, and `/tour/`, which shows images, captions,
code, tables, and task lists.

The new theme is a deliberately bare skeleton, not a copy of a built-in theme.
It renders everything correctly but plainly, so you can build up from there.

## What a theme contains

```text
themes/my-theme/
  theme.json        name, version, and supported Cress versions
  templates/
    page.gohtml     ordinary pages
    landing.gohtml  landing pages
    list.gohtml     list pages
    404.gohtml      not-found page
    shared.gohtml   anything shared, such as header and footer
  public/
    site.css        files published with the site
  locales/
    en.toml         translated labels, if you use any
```

`page.gohtml`, `landing.gohtml`, `list.gohtml`, and `404.gohtml` are required.
You can add more templates, including in subfolders such as
`partials/header.gohtml`. Only files in `public/` are published, under
`/themes/my-theme/`. Templates, the manifest, and locale files stay private.

A local theme with a built-in name, such as `themes/poppy/`, replaces that
built-in completely. Cress never mixes files from two themes, and it never
falls back to a built-in when your theme has an error. It fails and tells you
why instead. Symlinks and special files anywhere in a theme fail too.

## The manifest

`theme.json` names your theme and says which Cress versions it supports:

```json
{
  "name": "my-theme",
  "version": "1.0.0",
  "cress": {
    "major": 1,
    "minimum": "1.0.0"
  }
}
```

- `name` must match the folder name.
- `version` is your theme's own version, without a `v`.
- Cress loads the theme only if its major version equals `cress.major` and it
  is at least `cress.minimum`. A prerelease such as `1.2.0-rc.1` doesn't meet
  a minimum of `1.2.0`.
- Development builds of Cress report `dev` and check against a
  [baseline version](../internal/theme/manifest.go), with a warning.
- Unknown fields produce a warning. Invalid known fields fail.

The [manifest schema](../schemas/theme.schema.json) lists the allowed values.

## Writing templates

Templates use Go's [`html/template`](https://pkg.go.dev/html/template), which
escapes text automatically. Each template receives one page. A minimal
`page.gohtml` looks like this:

```html
<!doctype html>
<html lang="{{.Site.Language}}">
<head>
  <meta charset="utf-8">
  <title>{{.Title}}</title>
  <link rel="stylesheet" href="{{asset "site.css"}}">
</head>
<body>
  <main id="main">
    <h1>{{.Title}}</h1>
    {{.Body}}
  </main>
</body>
</html>
```

Template errors and missing fields stop the build. In preview, the terminal
shows the error and the last working version stays visible.

### Helpers

Two helpers are available:

- `{{asset "css/site.css"}}` returns the URL of `public/css/site.css`,
  including the site's URL prefix. The file must exist.
- `{{label "read_more"}}` returns a translated text from
  `locales/<language>.toml`, for example `read_more = "Read more"`. The
  language must match exactly: `en-US` doesn't fall back to `en`. A missing
file or key fails the build. An empty optional label stays empty.

Every theme needs a locale file for the site's exact language with non-empty
`not_found_title` and `not_found_message` labels. Other labels are optional
unless your templates use them. A language such as `en-US` does not fall back
to `en`. A locale file that exists must be valid, even if unused.

Templates have no access to files, the network, environment variables, or
commands. This keeps themes predictable and safe to share.

`404.gohtml` receives site data as a `page.Page`, like the other templates.
Cress renders it to `404.html` at the output root. Use the `not_found_title`
and `not_found_message` labels so the page appears in the site's language.
The shared `document-start` and `document-end` templates provide the usual
metadata, navigation, and skip link.

## The page data

Each template receives a `page.Page`. The
[render model](../internal/page/page.go) lists every field. The ones you use
most:

| Field | What it holds |
|-------|---------------|
| `.Title` | Page title, as plain text. |
| `.Type` | `ordinary`, `landing`, or `list`. |
| `.Body` | The page content as finished HTML. |
| `.Description` | Page description, already inherited from the site if needed. |
| `.Image` | Optional title image with `URL`, `Canonical`, `Alt`, `Caption`, and `Focus`. `Canonical` is its full public URL, or empty when no host is configured. |
| `.URL` | The page's address, including the URL prefix. |
| `.Canonical` | Full public URL, or empty when no host is configured. |
| `.StructuredData` | On the home page, the site's `[profile]` as schema.org data, or empty. |
| `.PublishedAt`, `.UpdatedAt` | Optional `YYYY-MM-DD` dates. |
| `.Excerpt` | Optional plain-text summary, used in lists. |
| `.Entries` | On list pages, the entries to show, already selected and sorted. |
| `.Pagination` | On split list pages, `Number`, `Total`, `Previous` and `Next` URLs, and `Pages` with `Number`, `URL`, and `Current`. Empty otherwise. |
| `.Site.Title`, `.Site.Copyright` | Site identity, possibly empty. |
| `.Site.Language` | The site language, such as `en`. |
| `.Site.URL` | The home page address. |
| `.Site.Navigation.Main`, `.Footer` | Ordered links with `Label` and `URL`. |
| `.Site.Brand` | Optional `Color`, `Favicon`, and `Logo`. |
| `.Site.Syntax.Highlight` | Whether syntax coloring is on. |
| `.Site.Custom` | Free-form data from `[theme.custom]`. |

Some things to keep in mind:

- **Everything is ready to use.** URLs already include the prefix, dates are
  validated, and descriptions are inherited. Don't rebuild or reinterpret
  them.
- **Lists are already decided.** Show `.Entries` in the given order. Entries
  have the same fields as pages, but no body.
- **Render pagination.** A long list is split into numbered pages, and
  `.Entries` holds only the current page. Link to the others from
  `.Pagination`, or readers can't reach the older entries.
- **Only `.Body` is HTML.** Everything else is plain text and gets escaped.
  Never turn other values into trusted HTML.
- **Images have no size.** Cress doesn't measure images. To avoid layout
  shifts, give cropped images a fixed shape in CSS. Let uncropped images keep
  their proportions.
- **Focus helps cropping.** `.Image.Focus` names the part of the image to
  keep visible, such as `top`. The built-in themes turn it into a
  `focus-<value>` class. Ignore it if you don't crop.
- **Captions are data.** `.Image.Caption` is a list of parts with `Text` and
  an optional `URL`. Render them inside `figcaption`, and leave the markup out
  when the list is empty.
- **Link previews are yours to render.** The built-in themes add Open Graph
  tags from the title, description, and `.Image.Canonical` when `.Canonical`
  is set. Leave them out when it is empty.
- **Structured data is ready to embed.** Write
  `<script type="application/ld+json">{{.StructuredData}}</script>` when it
  is set. The template engine encodes it as safe JSON.
- **Custom data is yours.** Document any `[theme.custom]` keys your theme
  reads. It is plain data, never HTML.

Drafts reach templates only in draft preview. Templates never see source
files, raw configuration, or pages outside the current build.

## Markup inside the body

`.Body` already contains some markup your CSS can style:

| Markup | Meaning |
|--------|---------|
| `<div class="table-wrapper" tabindex="0">` | Wraps every Markdown table, so wide tables can scroll and receive keyboard focus. Style the overflow and focus. |
| `align-left`, `align-center`, `align-right` | Column alignment from Markdown on table cells. |
| `<figure>` and `<figcaption>` | Captioned images and code blocks. |
| `<span class="captioned-image">` and `.image-caption` | Captioned images inside a paragraph. |
| `<p class="standalone-link">` | A paragraph with exactly one link and nothing else. Built-in themes highlight it as the main action on landing pages. |
| `<del>` | Strikethrough. |
| `<label>` with a disabled checkbox | A task list item. Keep the checkbox visible and disabled. |
| Heading IDs | Link targets for every heading. Don't replace them. |
| `<pre><code class="language-go">` | Code blocks, with token spans when highlighting is on. |

`standalone-link` only describes structure. Your theme decides whether it
means anything, for example combined with the page type.

Avoid IDs that might clash with content headings. `main`, `page-title`, and
IDs starting with `cress-` are reserved for themes and Cress, so content
can't use them. Cress uses `cress-caption-1` and similar IDs for inline
captions, so avoid those too.

## Syntax highlighting

When highlighting is on, code blocks contain `<span>` elements with short
[Chroma](https://github.com/alecthomas/chroma#the-html-formatter) classes,
such as `k` for keywords, `s` for strings, and `c` for comments. Your CSS
gives them color. Style them under `pre code`, because Cress doesn't add
Chroma's usual `.chroma` wrapper. Highlight.js themes use different classes
and won't work as they are.

Respect the site's choice. When `.Site.Syntax.Highlight` is false, don't color
code, even though the `language-*` class is still present. A simple way is to
load the colors conditionally:

```html
{{if .Site.Syntax.Highlight}}
<link rel="stylesheet" href="{{asset "highlight.css"}}">
{{end}}
```

Cress themes need no JavaScript highlighter. If you add one anyway, follow the
same switch and leave already highlighted blocks alone.

## Making your theme accessible

A good theme works for everyone. The
[accessibility guide](ACCESSIBILITY.md) explains why, and what the built-in
themes do. The essentials:

- Set `lang` from `.Site.Language`.
- Use `header`, `nav`, `main`, and `footer`, and label each `nav`.
- Add a skip link to `#main`, and give `main` `tabindex="-1"` so it can
  receive focus.
- Mark the current page in navigation with `aria-current="page"`.
- Keep focus outlines clearly visible.
- Check text contrast, and don't rely on color alone for links.
- Make sure the layout works at 200% zoom and on narrow screens.
- Respect `prefers-reduced-motion` and `forced-colors`.

## Translating labels

Built-in themes provide `en` and `de`. If your theme uses `label`, add a
locale file for each language your users need. A site whose language has no
locale file fails to build with your theme, rather than showing the wrong
language.

## Sharing your theme

Others install your theme by copying it to `themes/<name>/` in their site.
Before sharing:

- Test every page type and the content tour, with and without images.
- Test with highlighting on and off.
- Set `cress.minimum` to the oldest Cress version you have tested.
- Document any `[theme.custom]` keys and the languages you support.

Remember that templates and CSS are trusted code. Site owners should review a
theme before using it, just like raw HTML in content.
