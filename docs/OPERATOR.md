# Operator guide

This guide covers getting Cress onto a machine: installing it, pinning a
version, upgrading, and removing it again. It is about the tool, not about the
sites you build with it.

Once Cress is installed, the [user guide](./USER.md) covers creating and
building a site, and the [publishing guide](./PUBLISH.md) covers getting the
output onto a host.

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

Confirm the binary:

```bash
cress version
```

The script resolves the latest release, downloads the archive for your platform,
and verifies it against the checksums published with that release before
installing anything. A mismatch aborts the install rather than warning about it.

## Pin a version

The examples above track the latest release. Pin an exact one with `-t`:

```bash
curl -sSfL https://raw.githubusercontent.com/cozybadgerde/cress/trunk/scripts/install.sh \
  | sh -s -- -t v1.0.0
```

Pin wherever a rebuild has to produce the same site twice: CI, a container
image, a machine somebody else maintains. Version 1.0.0 fixes the public
surface, so `cress.toml`, the front matter, the CLI flags, the output layout,
and the theme template contract do not change within 1.x. Pinning is what makes
that a guarantee rather than an expectation.

Both flags also read an environment variable, which is often easier to set in a
CI job than to thread through a pipe: `CRESS_BINDIR` for `-b` and
`CRESS_VERSION` for `-t`.

## Install in CI

CI runners are ephemeral, so the install is part of the job. The same script
works, with two changes: pin the version, and install somewhere already on
`PATH`.

```yaml
- name: Install cress
  run: |
    curl -sSfL https://raw.githubusercontent.com/cozybadgerde/cress/trunk/scripts/install.sh \
      | sh -s -- -b /usr/local/bin -t v1.0.0
```

Pinning matters more here than anywhere else. An unpinned install means a
release published between two runs changes what the second one builds, without
a commit to point at.

The [publishing guide](./PUBLISH.md) has complete workflows for GitHub Pages and
GitLab Pages, with this step in place.

## Upgrade

Run the install command again. It overwrites the existing binary with the newer
one, so an upgrade is the same operation as the first install:

```bash
curl -sSfL https://raw.githubusercontent.com/cozybadgerde/cress/trunk/scripts/install.sh \
  | sh -s -- -b "$HOME/.local/bin"
```

Check what you have first with `cress version`, and read the
[release notes](https://github.com/cozybadgerde/cress/releases) when the major
version changes. Within 1.x nothing in a site needs migrating, for the reason
given under [Pin a version](#pin-a-version).

## Uninstall

Cress is a single binary and keeps no state, no cache, and no configuration
outside the site directories you created. Removing it is removing the file:

```bash
rm ~/.local/bin/cress        # or /usr/local/bin/cress for a system-wide install
```

Your sites are untouched. They are ordinary directories of Markdown and TOML,
and they outlive the tool that renders them.

## Troubleshooting the install

- **`cress: command not found` after installing.** The install directory is not
  on your `PATH`. For the rootless install, add `~/.local/bin` to it in your
  shell profile.
- **`checksum mismatch`.** The download did not match the published checksum.
  Retry once in case of a truncated transfer; if it persists, do not install
  the file, and open an issue.
- **No binary for your platform.** Cress publishes Linux and macOS builds for
  `amd64` and `arm64`. On anything else, build from source with the
  [developer guide](./DEVELOPER.md). Windows support is tracked in
  [#22](https://github.com/cozybadgerde/cress/issues/22).

For problems with a site rather than with the tool, see the troubleshooting
section of the [user guide](./USER.md#troubleshooting).
