# Distribute external themes through documentation and archives

Status: accepted.

Date: 2026-11-11.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Cozy Badger maintains themes outside the executable in
[cress-themes](https://github.com/cozybadgerde/cress-themes). Maintainers
need to find these themes and install them into a site's `themes/` folder.

Installing a theme now means cloning the collection and copying a folder.
Common mistakes are symlinked folders, which Cress rejects, and updates that
merge files instead of replacing the package. Compatibility is checked only
when the site builds.

[ADR 0007](0007-theme-packages-and-integration.md) states that maintainers
obtain and update local packages separately and that Cress does not download
them. [ADR 0008](0008-rendering-and-template-trust-boundary.md) treats theme
templates as trusted code.

## Decision drivers

- Keep generation and preview offline.
- Keep installation simple for non-technical maintainers.
- Avoid network code, package indexes, and release processes in Cress
  without a demonstrated need.
- Don't suggest that Cress vouches for code it installs.

## Considered options

- Document the collection and publish one archive per theme.
- Add a local `cress theme add` command that validates and copies a folder or
  archive.
- Add a `cress theme install` command that downloads a named theme from a
  Cozy Badger index and verifies its checksum.

## Decision outcome

Chosen option: "Document the collection and publish one archive per theme",
because it removes the need for Git without adding code or maintenance to
Cress, and leaves ADR 0007 unchanged.

The [user guide](../USER.md#choosing-a-theme) links to the collection.
The collection owns its catalogue, installation steps, and releases. Each
release publishes one archive per theme that contains the complete theme
folder, so maintainers download and extract one theme without Git.
Updates replace the whole folder.

Cress gets no install, download, or theme index command. Build and preview
continue to read only local themes and check compatibility before rendering.

### Consequences

- Good, because Cress has no network code or index to maintain.
- Good, because the collection can release themes independently of Cress.
- Good, because maintainers review and place theme files themselves, which
  matches the trust model for templates.
- Bad, because maintainers still copy folders by hand and can make mistakes
  that Cress reports only at build time.
- Bad, because the catalogue lives outside Cress, so maintainers must follow
  a link to find it.

A local add command remains the next step if copying causes repeated
problems. Downloading by name would need a new ADR that supersedes the
download rule in ADR 0007.

## Confirmation

Confirm that the user guide links to the collection and that each
collection release includes a complete, extractable archive per theme that
builds with a supported Cress release.

## More information

See [theme packages](0007-theme-packages-and-integration.md),
[offline generation](0002-offline-generation-and-deployment-boundary.md),
and [the CLI structure](0009-workflow-cli.md).
