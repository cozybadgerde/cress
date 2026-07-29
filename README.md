# Cress

[![CI](https://github.com/cozybadgerde/cress/actions/workflows/ci.yml/badge.svg?branch=trunk)](https://github.com/cozybadgerde/cress/actions/workflows/ci.yml)
[![Latest release](https://img.shields.io/github/v/release/cozybadgerde/cress?include_prereleases&label=release)](https://github.com/cozybadgerde/cress/releases)
[![Open issues](https://img.shields.io/github/issues/cozybadgerde/cress?label=issues)](https://github.com/cozybadgerde/cress/issues)
![Languages](https://img.shields.io/github/languages/count/cozybadgerde/cress)

A cozy static site generator. Fresh little sites, fast.

## Motivation

Most static site generators are either too much or too little. The big ones
carry taxonomies, shortcodes, data pipelines, and a config surface you grow into
for years. The tiny ones stop just short of being useful and leave you writing
the HTML plumbing yourself.

Cress aims for the middle. Give it Markdown, a small config file, and a theme,
and it renders a calm little website. The content stays clean, the theme owns
all the styling, and there is not much else to learn.

## Getting started

Install the latest release (rootless, into `~/.local/bin`) and scaffold a site:

```bash
curl -sSfL https://raw.githubusercontent.com/cozybadgerde/cress/trunk/scripts/install.sh \
  | sh -s -- -b "$HOME/.local/bin"
cress init my-site
cd my-site
cress serve
```

Open the printed address, edit the files under `content/`, and reload. When you
are ready to publish, run `cress build` and deploy the `public/` directory. For
a host-wide install and other options, see the [operator guide](docs/OPERATOR.md).

## Description

Cress turns a source tree into a static website:

```text
content/ (Markdown)  +  cress.toml (config)  +  theme  =  public/ (HTML)
```

Every Markdown file under `content/` becomes a page: `content/about.md` renders
to `public/about.html`, and `content/index.md` is the home page. Navigation is a
short, explicit list in `cress.toml`. A theme supplies the templates and the
CSS, so styling lives in exactly one place. The built-in theme works with no
setup; drop a directory under `themes/` when you want your own.

## Features

- **Markdown in, HTML out.** CommonMark plus GitHub-flavored extensions, via
  goldmark. Your content files stay free of layout.
- **Folders are your site map.** The content tree maps straight to the output
  tree. No sections model to learn.
- **Explicit navigation.** Flat `[nav.main]` and `[nav.footer]` tables in
  `cress.toml`, mkdocs-style. What you list is what shows up, in the order you
  write it.
- **One place for styling.** The theme owns every template and stylesheet. The
  built-in theme is cozy and light or dark aware out of the box.
- **White-label ready.** Set a logo, favicon, accent color, footer message, and
  copyright line in `cress.toml`. `cress init` drops placeholder logo and
  favicon files in `static/` for you to replace, so a fresh site already looks
  finished.
- **A 404 page you never have to write.** Every build emits `404.html` carrying
  your nav and styling. Override it with `content/404.md` when you want your
  own words.
- **Preview as you write.** `cress serve` builds the site, serves it, and
  rebuilds when a source file changes.
- **A single binary.** Rootless install, no runtime dependencies.

## Comparison

Cress aims for the middle ground: a single binary like Hugo or Zola, but a much
smaller surface to learn. The table sketches where it sits among established
tools.

|                 | Cress               | Hugo                | Zola                | Eleventy               | Jekyll               |
| --------------- | ------------------- | ------------------- | ------------------- | ---------------------- | -------------------- |
| Language        | Go                  | Go                  | Rust                | Node.js                | Ruby                 |
| Install         | single binary       | single binary       | single binary       | npm + Node             | gem + Ruby           |
| Config format   | TOML                | TOML, YAML, JSON    | TOML                | JS, JSON               | YAML                 |
| Templating      | Go templates        | Go templates        | Tera                | Nunjucks, Liquid, JS   | Liquid               |
| Front matter    | YAML                | TOML, YAML, JSON    | TOML, YAML          | YAML, JSON, JS         | YAML                 |
| Feature surface | minimal             | extensive           | rich                | flexible, plugin-based | mature, plugin-based |

Hugo and Cress build on the same Go template engine, so the syntax is identical.
The difference is scale: Hugo wraps it in hundreds of functions and a layout
system, while Cress uses the plain `html/template` standard library with a
minimal function set.

Cress deliberately ships less. If you need taxonomies, multilingual content,
shortcodes, or an asset pipeline today, Hugo or Zola will serve you better.
Cress is for small sites where simplicity is the point.

## Documentation

The documentation lives in the `docs/` directory:

- [User guide](docs/USER.md): configure and author a Cress site
- [Operator guide](docs/OPERATOR.md): install Cress and publish the output
- [Developer guide](docs/DEVELOPER.md): build the project and its layout

## Contributing

First off, thank you for considering contributing to this project. This document
assumes that you read it to understand how to contribute, raise issues, and
submit pull requests.

Please read our
[contributing guidelines](https://github.com/cozybadgerde/.github/blob/trunk/.github/CONTRIBUTING.md)
and [code of conduct](https://github.com/cozybadgerde/.github/blob/trunk/.github/CODE_OF_CONDUCT.md)
for more information.

## Contact

In case of any questions or feedback, please contact the project maintainers or
network with us via:

- website: [cozybadger.de](https://cozybadger.de)
- LinkedIn: [@cozybadgerde](https://www.linkedin.com/company/cozybadgerde/)
- e-mail: [hello@cozybadger.de](mailto:hello@cozybadger.de)

## License

[BSD 3-Clause](LICENSE) © Daniel Schier
