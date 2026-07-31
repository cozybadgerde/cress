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

The target directory has to be empty, though dotfiles do not count, so
`git init` first and then `cress init` works. To scaffold into a directory that
already holds other files, add `--force` (`-f`). Like the build, `cress init`
never overwrites: a starter file that is already there is left alone and
counted in the summary, so re-running it only fills in what is missing.

## Build

Render the site into `public/`:

```bash
cress build --source my-site
```

Run from inside the site directory to omit `--source`. Useful flags:

- `--output DIR` (`-o`): write somewhere other than `public/`.
- `--drafts`: include pages marked `draft`.

The build never deletes. It creates the output directory if needed, then writes
its own files over whatever is already there, leaving everything else alone. A
page you rename or remove therefore leaves its old HTML behind, and the build
names it:

```text
warning: public/guides/styleguide.html was not written by this build
```

Use `cress clean` to clear them out; `cress build` will not do it for you.

## Clean

Empty the output directory:

```bash
cress clean --source my-site
```

It takes `--output` (`-o`) the same way the build does, so it empties whatever
directory that build wrote into. Because deleting is the point, it asks first:

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
flag or no flag, when the target is the site root, a directory containing it,
or `content/`, `static/`, or `themes/`. It also refuses to run in a directory
with no `cress.toml`, which is what catches a clean aimed at the wrong place.

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

Every build also writes a `404.html` at the root of `public/`, carrying the
site's own navigation and styling. Most static hosts serve it for an address
that does not exist, but some need it named in their own configuration, so check
your host's documentation. It is not counted in the page total the build
reports, because you did not write it. Add `content/404.md` to replace it with
your own wording.

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
