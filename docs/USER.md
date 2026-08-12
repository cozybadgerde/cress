# User guide

This guide explains how to use Cress: creating a site, the config file, content
and front matter, navigation, themes, and the commands that build and preview
what you wrote.

## Before you start

Install Cress with the [operator guide](./OPERATOR.md). When the site is ready
to go out, the [publishing guide](./PUBLISH.md) covers getting it onto a host.

## Create a site

Scaffold a new site into a directory:

```bash
cress init my-site
```

This writes `cress.toml`, a `content/` directory with a handful of starter
pages, and a `static/` directory holding the default logo and favicon. The
starter pages are a working example of everything below, and they are yours to
delete once you have found your feet.

The target directory has to be empty, though dotfiles do not count, so
`git init` first and then `cress init` works. To scaffold into a directory that
already holds other files, add `--allow-existing`.

That flag only relaxes the empty-directory check. It grants no permission to
overwrite, because `cress init` never overwrites either way: a starter file that
is already there is left alone and counted in the summary, so re-running the
command only fills in what is missing.

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

Only `cress.toml` and `content/` are required. A built-in theme is used when
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
- `base_url`: the full address the site is published at, including any path.
  Set it to where the site really lives; see [Publishing to a
  subdirectory](#publishing-to-a-subdirectory) below.
- `theme`: the theme name. Empty resolves to `cress`, the default. `birch` is
  built in too; any other name must match a directory under `themes/`.
- `language`: the site's language tag, such as `en`, `de`, or `en-GB`. It
  becomes the document's `lang` attribute, which is what tells a screen reader
  how to pronounce the page and a browser whether to offer a translation.
  Leaving it out resolves to `en`, because a document with no language declared
  is worse than one declaring a plausible default. A page can override it in its
  front matter. The value is passed through as written, so a typo reaches the
  output unreported.

Unknown keys are rejected, so a typo fails the build instead of being ignored.

### Publishing to a subdirectory

Most small sites sit at the root of a domain, and a `base_url` with no path
(`https://example.com`) or no `base_url` at all both say so. Links come out
rooted at the domain: `/about.html`, `/style.css`. Leaving the key empty is the
same as leaving it out, and neither is a problem: because those links are
root-relative, one build serves correctly from every domain that points at it.
Only a subdirectory makes `base_url` necessary.

A site published to a subdirectory is different. GitHub Pages and GitLab Pages
both do this by default for a project: the site lands at
`https://user.github.io/my-site/` rather than at the domain root. Put the whole
address in `base_url`, path included:

```toml
[site]
base_url = "https://user.github.io/my-site"
```

Cress then roots every link it writes under that path: `/my-site/about.html`,
`/my-site/style.css`. The files in `public/` do not move; only the links inside
the pages change. `cress serve` previews the site at the same path, so what you
see locally is what gets published.

If you leave `base_url` at the domain root and publish to a subdirectory
anyway, the pages load but every link in them points one level too high: the
site renders unstyled, with a dead navigation and no images. Nothing in the
build can detect this, because where a site is published is not something the
files know. `base_url` is how cress finds out, so it is worth getting right
before the first publish.

`base_url` must be a full URL with a scheme and host. `example.com/my-site` is
rejected, because a host mistaken for a path would root the whole site under a
directory that does not exist.

### White-label branding

Five optional fields make the site your own:

- `logo`: a path or URL to a logo image, shown in the navigation. `cress init`
  drops a placeholder at `static/logo.svg`; replace it with your own, or point
  this field at another file in `static/`. The placeholder is a mid-green SVG,
  a tone chosen to hold up against both the light and the dark background.
- `logo_dark`: a second image, used instead of `logo` for readers whose system
  prefers a dark color scheme. Leave it out and `logo` is used for both.
- `favicon`: a path or URL to a favicon. `cress init` drops a placeholder at
  `static/favicon.png`; replace it the same way.
- `image`: a lead image for every page that names none of its own. Neither
  chrome like `logo` nor content like a picture you write into a page: the theme
  decides where it goes and whether to show one at all.
- `image_alt`: describes `image` for a reader who cannot see it.
- `image_caption`: the line rendered under `image`, for a credit, a license, or
  a disclosure. See [Lead images](#lead-images).
- `accent`: the accent color, a highlight the theme uses on links and other
  touches. A CSS hex value, for example `accent = "#4f7a4a"`.
- `accent_dark`: the accent for readers whose system prefers a dark color
  scheme. Leave it out and `accent` applies to both.

```toml
[site]
logo = "/logo.svg"
logo_dark = "/logo-dark.svg"
favicon = "/favicon.png"
accent = "#4f7a4a"
accent_dark = "#9ccb8f"
```

Leaving `logo` or `favicon` empty simply omits it. An accent that is not a hex
color fails the build with a clear message, and so does a `logo_dark` with no
`logo` beside it: the dark variant replaces the light one and has nothing to
replace on its own.

**Most logos do not need `logo_dark`.** Artwork in a mid-tone reads against
both backgrounds, which is how the scaffolded placeholder gets away with a
single fixed color, and an SVG drawn with `currentColor` adapts on its own.
Reach for the second file when the artwork can do neither: a raster logo, or a
wordmark fixed in near-black or near-white, which disappears into one of the
two backgrounds. Setting both adds a dark-scheme `<source>` to the logo, so the
right one is chosen before anything is fetched, with no flash of the wrong
image and no JavaScript.

A theme decides whether to honor this. All three built-in themes do. A custom theme
that never reads `logo_dark` shows the light logo to everyone, and cress does
not warn.

**Leaving the accents out is a real choice, not a shortcut.** cress has no
default accent of its own; unset, the theme picks one, and the built-in themes
pick a different value per color scheme. That matters more than it sounds: no
single color reaches the WCAG AA contrast ratio of 4.5:1 against both a light
and a dark background, so a brand color that reads beautifully in one scheme
can be close to invisible in the other. Set `accent` alone only if your color
survives both; otherwise set `accent_dark` as well.

### Footer text and copyright

Two more `[site]` fields fill the footer:

- `footer`: your own message, replacing the "Made with Cress" credit. Leave it
  out and the credit stays.
- `copyright`: a copyright notice. Leave it out and none is rendered.

```toml
[site]
footer = "A small corner of the internet."
copyright = "(c) {year} Cozy Badger"
```

In `copyright`, the token `{year}` becomes the year the site was built, so the
notice does not quietly go stale every January. Write the year yourself when you
want a fixed one or a range: `"(c) 2019-{year} Cozy Badger"` renders the start
year literally and expands only the end. Everything else in either field is
taken as written.

Both fields are **plain text**. Markup in them is escaped rather than rendered,
so `footer = "<b>bold</b>"` shows the angle brackets. That is deliberate: the
theme is the only place styling lives, and a config value that could inject
HTML into every page on the site would be a hole in that rule.

### Colored code blocks

Cress writes a fenced code block as `<pre><code class="language-go">` with the
code inside it as plain text. There is nothing there for a stylesheet to color,
so a theme can style the block as a whole and no further. Turn that on in the
`[markdown]` table:

```toml
[markdown]
highlight = true
```

Cress then reads each block in the language its fence names and wraps every
keyword, string, and comment in its own `<span>`. The spans carry a class and no
color, because the colors are the theme's to choose. All three built-in themes ship a
set for the light and the dark scheme. A theme that ships none
renders the code exactly as it did before, so turning this on can make a site
look unchanged.

A fence that names no language, or names one Cress does not know, is written as
it always was. Nothing else on the page moves either way.

They keep those colors in their own `highlight.css`, and a file in your
`static/` wins over a file the theme ships under the same name. So you can
replace the whole color scheme without writing a theme:

```text
static/
  highlight.css      your colors, used instead of the theme's
```

The classes are [chroma](https://github.com/alecthomas/chroma)'s, so a
stylesheet generated for chroma works as-is. The [theme
guide](./THEME.md#coloring-code) lists the classes worth styling.

The key defaults to off, and leaving it out is the same as writing `false`.
Turning it on rewrites every code block on the site, which is a change to your
published HTML rather than a display setting, so it is yours to make rather than
one Cress makes for you.

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
- `description`: a one-line summary of this page, used for the page's own
  metadata. Falls back to the site's `description` when absent.
- `image`: this page's lead image, falling back to the site's `image`.
- `image_alt`, `image_caption`: the description and the caption for this page's
  `image`. See [Lead images](#lead-images).
- `language`: this page's language tag, overriding the site's `language` for
  this page alone. Set it on the one page you wrote in another language; leave
  it out everywhere else.
- `layout`: the name of an alternative layout in your theme. Leave it out and
  the page gets the theme's standard layout, which is what nearly every page
  wants. See [Page layouts](#page-layouts).

Any other field is passed through to the theme untouched, so a theme can read
its own keys. Unlike `cress.toml`, front matter is deliberately permissive: an
unknown key is never an error, so notes written in Obsidian or another editor
that adds front-matter fields of its own build without being edited first. The
trade-off is that a misspelled `title` is ignored rather than reported.

### Image captions

An image that is the whole of its paragraph becomes a `<figure>`, and the
quoted title after the path becomes its `<figcaption>`:

```markdown
![Boats at sunset](/harbor.webp "Photo: Jane Doe, CC BY-SA 4.0")
```

```html
<figure>
<img src="/harbor.webp" alt="Boats at sunset">
<figcaption>Photo: Jane Doe, CC BY-SA 4.0</figcaption>
</figure>
```

This is plain Markdown, not a Cress extension, so the same file renders
everywhere else too. Other tools show the title as a tooltip instead.

Four things to know:

- **Alt text and a caption are different strings.** Alt text stands in for the
  picture and only somebody who cannot see it ever meets the text. A caption is
  text everybody reads. Write both, and do not repeat one in the other.
- **A standalone image is wrapped whether or not it has a title.** Without one
  it is a `<figure>` with no `<figcaption>`, so a theme has a single shape to
  style.
- **Only a paragraph that holds nothing but the image is wrapped.** An image
  inside a sentence stays inline and keeps its title as a tooltip, and so does
  an image wrapped in a link.
- **A caption is text, not Markdown.** Links and emphasis inside a title are
  printed literally. Write the `<figure>` by hand when a caption needs a link;
  Cress passes raw HTML through untouched.

Captions are worth reaching for whenever an image has to say something out
loud, such as who made it, what license it carries, or that it was generated by
a machine. Cress does not check any of that, and the wording stays yours. It
gives you somewhere visible to put it.

### Lead images

A page can name one image that is not part of its text, in front matter:

```markdown
---
title: A cozy afternoon
image: /cress-on-a-windowsill.webp
image_alt: A pot of cress on a sunny windowsill
image_caption: "Photo: Jane Doe, CC BY-SA 4.0"
---
```

The theme places it: where it goes and how large it is are the theme's
decisions, not yours. What it says is still yours, which is what the other two
keys are for.

`image_alt` is the description somebody gets instead of the picture. Leave it
out and the image is decorative, which is a real choice for an illustration and
the wrong one for a diagram or a screenshot.

`image_caption` is the line under the picture, which everybody reads. Use it for
whatever the image is obliged to say out loud: who made it, what license it
carries, or that it was generated by a machine. What those obligations are
depends on where you are, and Cress knows none of it. It gives you the place to
put the statement and leaves the wording to you. Leave the key out and no
caption is rendered.

Set one for the whole site and let pages override it:

```toml
[site]
image = "/social.webp"
image_alt = "Cress growing in a wooden tray"
image_caption = "AI-generated illustration."
```

**A page's own `image` takes its own alt and caption, never the site's.** If a
page names an image and no caption, it gets no caption, rather than inheriting
one written about a different picture. A credit on the wrong image is worse than
no credit at all.

With neither set, no lead image is rendered and nothing is missing. Both
built-in themes render one, caption included. Poppy makes it the photograph
taped to the top of the page, which is the shape a lead image was made for. A
theme that ignores the key shows
nothing, which is a valid choice for it to make: check before you rely on one if
you have a statement you must display.

Paths work like `logo` and `favicon`: write them rooted at the site (`/x.webp`)
and Cress handles a site published to a subdirectory for you.

### Cress trusts your content

Raw HTML in a Markdown file is passed through to the page exactly as written,
and the page body is handed to the theme without escaping. That is what lets you
drop in an embed, a `<figure>` with a linked caption, or any markup Markdown has
no syntax for.

It also means a Markdown file is as trusted as a template. A `<script>` tag in
`content/` is a `<script>` tag on the published page, so:

- **Do not build a site from Markdown you have not read.** Treat a pull request
  that touches `content/` as a change to the site's code, not to its copy.
- **The same goes for a theme.** Its templates run at build time and its
  `static/` directory is copied into the output verbatim.

Config values are the exception, and deliberately so: `footer` and `copyright`
are escaped, and `accent` must be a hex color, because a key that could inject
markup into every page would undercut the theme being the only place styling
lives. The Markdown body is the one place markup is trusted, because that is
where an author is writing the page.

## Writing good content

Cress does not police what you write. None of the following is enforced, and
none of it is a rule the builder checks. They are the conventions that keep a
site readable as it grows, collected here because each one is easy to get wrong
once and then work around forever.

The starter site puts them into practice in `content/guides/writing.md`, which
you can read on your own site and then delete.

**File and folder names.** The name becomes the URL, so it outlives the file.
Use lowercase and hyphens, avoid spaces, and pick the name before you start
writing. Renaming later breaks every inbound link, and Cress does not warn you.

**Headings.** Use one `# ` per page. Cress falls back to it for the page title,
which makes the first heading structural rather than decorative. Below it, go
down through `##` and `###` in order without skipping a level: headings are the
page's outline, not a way to change type size.

**Links.** Write internal links root-relative, starting with `/`, so they
resolve identically from every page regardless of folder depth. Make the link
text describe the destination; "here" is useless to anyone scanning the page or
listening to it. Cress reroots these links when the site is published to a
subdirectory, so `/about.html` keeps working wherever the site lives. Links you
write as raw HTML are the exception: Cress passes that markup through untouched,
so a `<a href="/about.html">` in a content file stays exactly as written.

**Images.** They live in `static/` and are linked from the site root, as
described under "Site layout" above. Cress has no asset pipeline, so
whatever you commit is what visitors download, at full size: resize before
committing. Choose the format for the artwork, WebP or JPEG for photographs,
SVG for line art and diagrams, PNG when you need transparency or a lossless
original. Always write alt text, and write what somebody would need instead of
the image rather than a repeat of the sentence around it. Add a caption when
the picture has something to say to every reader, as described under "Image
captions" above.

**Front matter.** `title` and `draft` are the only keys Cress reads. Everything
else is passed to the theme, so treat front matter as a place to serve the
theme's needs, not a schema to satisfy.

**Drafts.** `draft: true` keeps a page out of the build. It does not retract a
page that has already been published, because a build never deletes: the old
HTML stays in the output until you remove it yourself.

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

A theme owns the templates and the styling. Three are built into the binary and
need no setup, so switching is one line of config and nothing to download:

- **`cress`**, the default. Warm paper, a serif body, a muted sage accent, and
  a lot of air: headings take their weight from size and space rather than
  bold. Quiet, for a small personal site.
- **`birch`**. A developer-tool look: crisp and sans-serif, blue links, ruled
  headings, striped tables, and syntax colors a developer already reads code
  in. For a page about technical work, where the code is what you want read.
- **`poppy`**. A hand-made recipe book: warm paper, a bookish serif for the
  headings, every picture taped in slightly crooked, and quotes on scraps torn
  out of another page. For recipes, a craft log, a reading diary, notes to
  friends.

```toml
[site]
theme = "poppy"
```

All three honor `accent` and `accent_dark`, and all three are light- and
dark-scheme aware. Each picks a different default: cress a green, birch its link
blue, poppy the red of the flower it is named after, which it uses on links, on
its first heading and on the flower watermarked into its cover.

To start your own, scaffold one:

```bash
cress theme init mine
```

That writes `themes/mine/`: `page.html`, a `landing` layout so the scaffolded
home page keeps working, the partials the two share, and a stylesheet. Every
file is commented, so the starter doubles as a tour of what a template receives.
Name it in the config and it renders your site:

```toml
[site]
theme = "mine"
```

The command takes the same flags as `cress init`. Use `--source`/`-s` to name a
site root other than the current directory, and `--allow-existing` to scaffold
into a theme directory that already has files. Like `cress init`, it never
overwrites: a file already there is left alone and reported.

A theme is a directory under `themes/`, whether scaffolded or written by hand:

```text
themes/
  mine/
    theme.toml       optional: who made the theme, and what it needs
    templates/       the templates that render your pages
    static/          CSS and other assets, copied to the site root
```

A theme somebody sent you goes in the same place. Drop the directory under
`themes/`, name it in `theme =`, and it renders your site. A name that is
neither a directory nor a built-in theme is an error naming the ones that are
built in, so a typo fails the build rather than quietly falling back to the
default.

A directory under `themes/` wins over a built-in theme of the same name, so
`themes/birch/` replaces the built-in `birch` rather than clashing with it. The
build says when that happens, because nothing in the finished page would:

```console
$ cress build
warning: theme "birch" was loaded from themes/birch, which shadows the built-in
theme of the same name; rename the directory to use the built-in one
```

Rename the directory and pick the name up in `theme =` if you meant to write
your own. Keep it if you meant to replace the shipped one.

Open its `theme.toml`, if it has one, to see who wrote the theme and under what
license. It may also name the version of Cress the theme was built for, and if
that does not match the Cress you are running, the build says so:

```console
$ cress build
warning: theme "borrowed": theme.toml declares cress = "1.4", but this is cress 1.1.0; the theme may read fields this version does not provide
built 7 page(s) into public
```

That is a warning and the site still builds. A theme reading something your
Cress does not have gets an empty value rather than an error, so pages may come
out missing a piece. Upgrading Cress is the fix; see the
[operator guide](./OPERATOR.md).

To check a theme against the rules Cress expects of it, without building a site:

```bash
cress theme validate mine
```

With no name it checks the theme your config already points at. It prints what
is wrong and where, exits non-zero when it finds anything, and takes `--quiet`
to print nothing and leave only the exit status. It is most useful while writing
a theme, and it works just as well on one somebody sent you.

Writing a theme, rather than dropping one in, is its own guide. The
[theme guide](./THEME.md) covers the directory in full, the data a template
receives, the rules a theme has to follow, and what `validate` checks.

### Page layouts

Most pages should look alike, and by default they do: every page is rendered
through the theme's `page.html`. A page that needs a different shape names a
layout in its front matter, and the theme renders it through the matching
template instead.

```markdown
---
title: Welcome
layout: landing
---

# Welcome
```

That page renders through `templates/landing.html`. All three built-in themes ship
one alternative layout, `landing`, which opens the page with a panel filling the
first screen, the site's name set large in it, and the page's own content
following below. An arrow at the foot of the panel links
down to that content, so the first screen does not read as the whole page. The
header and footer are the site's usual ones, so navigation works exactly as it
does everywhere else. The site `cress init` creates uses it on the home page,
so you can see the difference immediately.

What sits behind the name is the theme's own: `cress` draws fine vertical
lines, `birch` a faint grid, `poppy` the flower it is named after.

Write the page as you would any other. The panel takes the site's name from
`title` in `cress.toml` and its line of text from the page's `description`,
falling back to the site's, so the Markdown below is only the content: give it
its own `# ` heading the way every other page has one.

Layout names belong to the theme, not to Cress. There is no fixed list of
names: a theme offers whatever layouts it documents, and `landing` means
whatever the theme in use decides it means. Two consequences are worth knowing
before you reach for the key:

- **Swapping themes can change a page.** If your new theme defines no
  `landing`, the page falls back to the standard layout and the build says so:

  ```text
  content/index.md names layout "landing", which theme "birch" does not define
  ```

  The build still succeeds and the page is still published. Remove the key, or
  rename it to something the new theme offers.
- **A typo behaves the same way.** `layout: lading` is a name no theme defines,
  so you get the standard layout and that warning. Read the warnings a build
  prints and a mistyped layout is obvious; ignore them and it looks like the
  key did nothing.

Use `layout` on the handful of pages that genuinely differ, such as a home
page. Setting it on every page means repeating it in every file, and Cress has
no way to set a layout for a whole directory at once.

## The 404 page

Every build writes a `404.html` at the root of `public/`. GitHub Pages, GitLab
Pages, Netlify, and most other static hosts serve that file for an address that
does not exist, so you get a "page not found" carrying your own navigation and
styling without doing anything.

Three sources can supply it, and the most specific one wins:

1. `content/404.md`, if you wrote one. It renders like any other page, so the
   words and the front matter are yours.
2. `templates/404.html` in the theme, if it defines one. It receives the same
   data `page.html` does, so a theme can give the 404 its own layout.
3. Otherwise cress writes one for you: a short "Page not found" with a link
   home, rendered through the theme's `page.html`.

Write `content/404.md` when you want your own wording. Nothing else needs
changing, and the page you write replaces the generated one:

```markdown
---
title: Lost
---

# Nothing here

That page moved or never existed. Try the [home page](/).
```

`cress serve` answers a missing address with the same file, so the preview shows
what a visitor to the published site sees. Some hosts need the 404 wired up in
their own configuration; check your host's documentation.

## Preview

Preview locally while you write:

```bash
cress serve
```

This builds the site, serves it at `http://localhost:1313`, and rebuilds when a
source file changes. Reload the browser to see updates; the preview does not
refresh it for you. Stop the server with Ctrl-C.

Listen somewhere else with `--addr`:

```bash
cress serve --addr localhost:8080
```

A site with a path in its `base_url` is previewed at that path too, so what you
see locally is what gets published.

## Build

Render the site into `public/`:

```bash
cress build
```

Run the command from inside the site directory, or pass `--source` (`-s`)
pointing at the directory that holds `cress.toml`. Two more flags:

- `--output DIR` (`-o`): write somewhere other than `public/`.
- `--drafts`: include pages marked `draft`. `cress serve` takes this one too.

The build never deletes. It creates the output directory if needed, then writes
its own files over whatever is already there and leaves everything else alone.
That is what makes it safe to run unattended, and it is why a page you rename or
remove leaves its old HTML behind, still served and indistinguishable from a
live page. The build names each one:

```text
warning: public/guides/styleguide.html was not written by this build
```

## Clean

Empty the output directory:

```bash
cress clean
```

It takes `--source` and `--output` the way the build does, so it empties
whatever directory that build wrote into. Because deleting is the point, it asks
first:

```console
$ cress clean
remove 13 file(s) from /home/you/my-site/public? [y/N]:
```

The absolute path in that question is the part worth reading. `--output` makes
the target whatever you last typed, and the prompt is where a mistyped path
shows itself before anything goes.

Two flags skip the question:

- `--dry-run`: list every file that would go and remove none.
- `--force` (`-f`): remove without asking. This is the one for CI, a Taskfile,
  or anywhere no one is watching. Without a terminal to answer from, `cress
  clean` refuses rather than guessing.

`--force` skips the question, never the checks. `cress clean` refuses outright,
flag or no flag, when the target is the site root, a directory containing it, or
`content/`, `static/`, or `themes/`. It also refuses to run in a directory with
no `cress.toml`, which is what catches a clean aimed at the wrong place.

Clean means empty. Everything in the output directory goes, including dotfiles,
and the directory itself stays behind. There is no keep-list, because a list of
survivors is a promise that grows with every host and never shrinks. Two
consequences are worth planning for:

- **Files your site needs, such as `.nojekyll` or `.well-known/`, belong in
  `static/`.** The build copies that directory into the output every time, so
  they come back on the next build instead of needing to survive the clean.
- **Do not point `cress clean` at a git worktree**, the way a `gh-pages`
  checkout in `public/` is one. Its `.git` goes with everything else. The repo
  itself is unharmed and `git worktree repair` relinks it, but the run is not
  what you wanted.

## Publishing

The `public/` directory is a complete static site, ready for any static host.
The [publishing guide](./PUBLISH.md) covers GitHub Pages, GitLab Pages, and your
own server, along with what `base_url` has to say for each.

## Troubleshooting

- **`config not found`**: run the command from the site directory, or pass
  `--source` pointing at the directory that holds `cress.toml`.
- **`unknown key(s)`**: a key in `cress.toml` is misspelled or unsupported. The
  message names the offending keys.
- **`theme ... not found`**: the `theme` name is neither a built-in theme nor a
  directory under `themes/`. The message names the built-in ones.
- **A nav entry is missing**: a warning like `nav.main entry ... points at
  missing content` names the group and means the path does not match a file
  under `content/`.
- **The site renders unstyled after publishing**: the links point one level too
  high, which means `base_url` does not match where the site really lives. See
  [Publishing to a subdirectory](#publishing-to-a-subdirectory).

For problems installing Cress itself, see the
[operator guide](./OPERATOR.md#troubleshooting-the-install).
