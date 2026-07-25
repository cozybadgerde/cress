# Operator guide

This guide covers installing Cress, building a site, and publishing the output.

## Prerequisites

- A shell with `curl` or `wget`, plus `tar` and `sha256sum` (or `shasum`).
- Cress ships binaries for Linux and macOS, on `amd64` and `arm64`.

To build from source instead, see the [developer guide](./DEVELOPER.md).

## Install

Install the latest release rootless, into `~/.local/bin`:

```bash
curl -sSfL https://raw.githubusercontent.com/cozybadgerde/cress/trunk/scripts/install.sh \
  | sh -s -- -b "$HOME/.local/bin"
```

Install system-wide, into `/usr/local/bin`:

```bash
curl -sSfL https://raw.githubusercontent.com/cozybadgerde/cress/trunk/scripts/install.sh | sudo sh
```

Pin a specific version with `-t`:

```bash
curl -sSfL https://raw.githubusercontent.com/cozybadgerde/cress/trunk/scripts/install.sh \
  | sh -s -- -t v1.0.0
```

The script verifies the release checksum before installing. Confirm the binary:

```bash
cress version
```

## Create a site

Scaffold a new site into a directory:

```bash
cress init my-site
```

This writes `cress.toml`, a `content/` directory with a handful of starter
pages, and a `static/` directory holding the default logo and favicon. For what
these files contain and how to author them, see the [user guide](./USER.md).

## Build

Render the site into `public/`:

```bash
cress build --source my-site
```

Run from inside the site directory to omit `--source`. Useful flags:

- `--output DIR` (`-o`): write somewhere other than `public/`.
- `--drafts`: include pages marked `draft`.

The build replaces the output directory's contents on each run.

## Preview

Serve the site locally and rebuild on change:

```bash
cress serve --source my-site
```

By default it listens on `localhost:1313`. Change it with `--addr`:

```bash
cress serve --addr localhost:8080
```

Stop the server with Ctrl-C. The preview does not refresh the browser for you;
reload the page after a change.

## Publish

The `public/` directory is a complete static site. Deploy it to any static host
or web server. For example, copy it to a server's document root:

```bash
cress build --source my-site
rsync -a --delete my-site/public/ user@host:/var/www/my-site/
```

Cress writes root-relative links, so the site expects to be served from the root
of its domain.

## Troubleshooting

- **`config not found`**: run the command from the site directory, or pass
  `--source` pointing at the directory that holds `cress.toml`.
- **`unknown key(s)`**: a key in `cress.toml` is misspelled or unsupported. The
  message names the offending keys.
- **`theme ... not found`**: the `theme` name has no matching directory under
  `themes/`. Use `cress` for the built-in theme, or add the theme directory.
- **A nav entry is missing**: a warning like `nav.main entry ... points at
  missing content` names the group and means the path does not match a file
  under `content/`.
