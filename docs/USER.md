# User guide

Cress turns Markdown files into a static website. It runs locally, needs no
internet connection, and writes plain HTML and CSS that any host can serve.
Your site's `cress.toml` explains every setting. The
[content guide](CONTENT.md) explains how to write pages. The
[build and preview document](OUTPUT.md) explains output handling.

## Install Cress

Follow the [installation guide](INSTALL.md), then check that it works:

```sh
cress --version
```

## Create your site

From the folder where you keep projects:

```sh
cress site init my-site
cd my-site
```

This creates sample pages, a `cress.toml` configuration file, placeholder
images, and a `.gitignore`. Cress never writes into a folder that already has
files in it, so your existing work is safe.

Poppy is the default theme. To start with another one, use
`cress site init --theme clover my-site` or
`cress site init --theme anemone my-site`.

## Write your first page

Open `content/about.md` and replace the example:

```markdown
+++
title = "About us"
+++

## Work with people who care

We help small businesses build a useful presence online.
```

The block between the `+++` lines holds the page's settings. The theme shows
the `title` as the main heading, so start your own headings at level two.
This page appears at `/about/`. Add another `.md` file under `content/` for
each new page.

The sample site includes a content tour at `content/tour.md`. Open it next to
its preview to see examples of images, captions, code, tables, and task lists.

## Make the site yours

Open `cress.toml`. Every setting has a comment explaining what it does.

- Change `[site] title` to your name or your organization's name, and update
  the description.
- Adjust `[navigation]` when you add pages. Use published addresses such as
  `/about/`, not file names.
- Replace the placeholder logo and favicon in `static/images/`, or remove
  their settings from `[brand]`.
- Set `[site] host` to your public address, such as `https://example.org`,
  before you publish. Until then, Cress reminds you that it can't create
  canonical URLs, link previews, structured data, feeds, or a sitemap.
- Describe who is behind the site in `[profile]`, such as your address and
  opening hours. Search engines can then show them in their results.

## Preview your changes

```sh
cress site preview --port 0
```

Open the address Cress prints. Each time you save a file, Cress rebuilds the
site. Refresh the browser to see the change. If something is wrong, the last
working version stays visible and the terminal tells you what to fix.

To see draft pages too, stop preview with Ctrl-C and run:

```sh
cress site preview --drafts --port 0
```

Preview never touches your published output.

## Build and publish

Stop preview with Ctrl-C, then build:

```sh
cress site build
```

Cress writes the finished site to `public/`. Before publishing, check your
content, contact details, and navigation, and remove the content tour once
you no longer need it. Fill in the sample legal notice and privacy pages,
or remove them if you don't need them. Then upload the contents of `public/` with your
hosting provider's tools.

Cress manages `public/` and the hidden `.cress/` folder for you. Don't edit
or delete them by hand. Read [keeping your output safe](#keeping-your-output-safe)
to learn why.

## Choosing a theme

All three built-in themes work without JavaScript or external resources.
They support every page type and include English and German text.

- **Poppy** uses an off-white background, dark green accents, and a single
  reading column. It suits freelancers.
- **Clover** uses white and slate, dark blue accents, and a solid title
  block. It suits established small businesses.
- **Anemone** uses paper tones, serif headings, and framed photos that tilt
  slightly on wide screens. It suits makers, shops, and cafes.

Switch themes any time by changing `[theme] name` in `cress.toml`. To compare
them, create a separate test site for each. To build your own, see the
[theme developer guide](THEMES.md).

## Organizing your pages

Most sites need only ordinary pages. Two other page types help:

- A **landing page** greets visitors with a title image and a short
  introduction. The home page is usually one.
- A **list page** collects the pages in its folder, such as journal entries
  or projects, and shows them as linked entries with dates and images.

Set the type in the page settings with `type = "landing"` or
`type = "list"`. The sample site has one of each.

Keep a page's images next to it: `about.md` can use files in
`about.assets/`. Put shared files such as your logo in `static/`.
See [add images](CONTENT.md#add-images) for details.

## Writing good content

Use these practices to make your content easier to read and navigate:

- Use headings in order, and write link text that makes sense on its own.
- Describe every meaningful image with `alt` text. Use `alt = ""` for purely
  decorative images.
- Explain important changes in words, not just with strikethrough.
- Check that text is still easy to read after changing the brand color.

The [accessibility guide](ACCESSIBILITY.md) explains why these matter and
what Cress already takes care of.

Raw HTML in your Markdown is published as written. Only include HTML you
trust, and review anything copied from elsewhere.

## Keeping your output safe

Cress manages `public/` and replaces it completely on each build, including
removing manual additions. Keep everything you want to publish in `content/`
or `static/`.

To protect your other files, Cress refuses to write into a `public/` folder it
didn't create. Builds also happen in a separate staging area first, so a
failed build leaves your previous site untouched. If something goes wrong
halfway, the next build tidies up automatically when it safely can.

If Cress stops with an error about output, don't delete `.cress/` or edit
`public/` to get past it. Read the message and see
[output protection](OUTPUT.md#output-protection).

## Experiment safely

Try new ideas in a separate test site rather than your real one:

```sh
cress site init /tmp/cress-test
```

Never initialize over a real project or keep source files inside `public/`.

## Editor support

Editors that understand JSON Schema can check your settings as you type. The
schemas are in the `schemas/` folder of the Cress source, at the same version
as your `cress` executable:

- `site.schema.json` for `cress.toml`
- `frontmatter.schema.json` for page settings between the `+++` lines
- `theme.schema.json` for a theme's `theme.json`

Keep the three files together. Cress checks more than a schema can, such as
links and theme compatibility, so a build remains the final check.

## Getting help

Use `--help` to see available commands and flags:

```sh
cress --help
cress site preview --help
```

Site commands use the current folder unless you pass another one, as in
`cress site build my-site`.
