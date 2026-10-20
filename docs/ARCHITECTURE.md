# Cress architecture

This guide describes Cress's agreed structure: its boundary, main
responsibilities, data flow, and failure boundaries. The
[PRD](PRD.md) owns product behavior. The [ADRs](adr/) own decision
rationale. This guide links to both instead of repeating them.

Status: agreed target structure, with phases 1 and 2 implemented.
The current build handles site contracts, metadata, navigation, ordinary
and landing pages, and new or empty output. Later pipeline stages below
remain target behavior. Package layouts and interfaces emerge with
implementation.

## System context

Cress is a single offline executable. A maintainer runs it locally or in
CI to turn a site project into static files. Deployment, hosting, and
theme downloads stay outside Cress.

```text
                 site project
    configuration, content, assets, local themes
                       |
                       v
  maintainer --> cress executable --> deployable output --> external deployment
                 (built-in themes,
                  skeleton embedded)
                       |
                       v
               private preview output --> local browser
```

| Actor or system | Relationship |
|-----------------|--------------|
| Maintainer | Runs `cress site` and `cress theme` commands and owns the site project. |
| Business owner | Supplies Markdown content through the maintainer. |
| Theme author | Builds a local theme package against the theme contract. |
| Deployment tooling | Copies deployable output to a host. Not part of Cress. |
| Site visitor | Uses the deployed static site. |

Generation reads only local inputs and embedded themes. It never uploads,
fetches referenced resources, or executes embedded code. See
[ADR 0002](adr/0002-offline-generation-and-deployment-boundary.md).

## Responsibilities

Cress is one generator with four responsibilities. They describe ownership
and test boundaries, not required packages or interfaces
([ADR 0008](adr/0008-rendering-and-template-trust-boundary.md)).

| Responsibility | Owns | Does not own |
|----------------|------|--------------|
| Generator | Coordinates a build or preview, collects diagnostics, and propagates failures. | Content rules, template logic, or filesystem replacement rules. |
| Content preparation | Loading and validating configuration and frontmatter, metadata resolution, drafts, routes, URLs, navigation, lists, Markdown-to-HTML conversion, captions, and asset ownership. | Page layouts or presentation. |
| Page renderer | Selecting templates by resolved page type and composing the render model with theme templates into page HTML. | Discovering, filtering, or reordering content. |
| Output handling | Staging, collision checks, ownership checks, writer locking, complete replacement, recovery, and preview session cleanup. | Deciding what content exists. |

Two supporting areas serve these responsibilities:

- **Theme integration** resolves the selected theme package, checks its
  manifest and runtime compatibility, and loads its templates, public
  assets, and labels ([ADR 0007](adr/0007-theme-packages-and-integration.md)).
- **CLI** maps `cress <scope> <action>` commands onto the generator and
  reports outcomes and exit status
  ([ADR 0009](adr/0009-workflow-cli.md)).

### Render model

The render model is the only data a theme sees. It holds prepared site
data, resolved navigation, the current page's metadata and rendered body,
list entries, title-image captions, theme labels, and the open theme
customizations. All URLs in it are already resolved for the deployment
path.

Themes do not receive raw configuration, the content tree, or generator
internals. Rendered HTML stays distinct from plain values, which keep
contextual escaping. The render model is a documented theme contract and
follows the [compatibility policy](adr/0004-shared-contract-policy.md).

## Build flow

A build runs these stages in order. Any error stops the build before
existing output changes.

1. **Load.** Read the site configuration, Markdown files, page assets,
   and shared static files. Parse TOML configuration and frontmatter.
2. **Validate.** Reject malformed input and invalid known fields. Warn
   about unknown fields. Keep open theme customizations as custom data.
3. **Resolve.** Apply site metadata defaults and explicit page overrides,
   page types, and draft visibility. Derive directory-index routes and
   resolve generator-owned URLs beneath the deployment path.
4. **Prepare.** Convert Markdown to HTML, resolve Markdown links and
   images, highlight code, render captions, build list entries, and
   prepare navigation and publication metadata.
5. **Select theme.** Use a same-name local package when one exists, and
   a built-in theme only when it is absent. An invalid local package
   fails. Check compatibility, required page-type support, and labels
   before rendering.
6. **Render.** Compose each page from the render model and theme
   templates.
7. **Stage.** Write pages, sitemap, `robots.txt`, page assets, shared
   static files, and public theme assets into a staging area. Reject
   destination collisions and name the competing sources.
8. **Publish.** Take the writer lock, verify ownership of the target,
   and replace the previous output as a complete result. Restore it on
   failure where possible.

Content preparation can run and be tested without templates. The renderer
can be tested with a hand-built render model.

## Preview flow

Preview reuses the full build flow for every rebuild; incremental
rendering is not required.

- Each session owns a private temporary output directory and a fixed
  draft choice. Sessions never share output or touch deployable output.
- The session serves its last verified output. A failed rebuild keeps
  that output and reports the failure.
- Source edits schedule a rebuild. Edits during a rebuild schedule
  another one, so the latest source is never lost. Output changes never
  trigger rebuilds.
- A rebuild blocked by another writer stays pending and retries. Other
  failures do not retry.
- If no verified output remains, the session stops serving and exits
  unsuccessfully with recovery guidance.
- Orderly shutdown removes the session's output. Later sessions reclaim
  only verifiably inactive sessions.

See [ADR 0010](adr/0010-staged-output-and-ownership.md) and
[the maintainer workflow](PRD.md#maintainer-workflow).

## Failure and trust boundaries

| Boundary | Rule | Source |
|----------|------|--------|
| Inputs | Malformed input and invalid known fields fail. Unknown fields warn. | [ADR 0004](adr/0004-shared-contract-policy.md) |
| Output ownership | Never replace nonempty output that Cress cannot prove it owns. Refuse symlinks and non-directories. | [ADR 0010](adr/0010-staged-output-and-ownership.md) |
| Writers | One writer per site through OS advisory locking. Explicit builds and preview startup fail fast. | [ADR 0010](adr/0010-staged-output-and-ownership.md) |
| Recovery | Restore verified output or stop with guidance. Never guess destructive repairs. | [ADR 0010](adr/0010-staged-output-and-ownership.md) |
| Drafts | Draft visibility is per build or session and never leaks into deployable output. | [ADR 0010](adr/0010-staged-output-and-ownership.md) |
| Themes | Templates are trusted, not sandboxed. Helpers expose no filesystem, network, environment, or shell access. | [ADR 0008](adr/0008-rendering-and-template-trust-boundary.md) |
| Authored content | Markdown and raw HTML are trusted and not sanitized. Authored HTML URLs pass through unchanged. | [ADR 0015](adr/0015-trusted-markdown-and-raw-html.md) |
| Publication | Only page output, generated metadata, assets, and designated public theme assets are published. | [ADR 0003](adr/0003-content-tree-and-routes.md) |

Failures are reported with an unsuccessful exit status. Cress does not
hide failures or substitute silent fallbacks.

## Development structure

The executable entry point is `cmd/cress`. Current implementation packages
follow the responsibilities above:

| Package | Responsibility |
|---------|----------------|
| `internal/command` | urfave/cli command definitions and terminal reporting. |
| `internal/generator` | Build coordination. |
| `internal/content` | Site/frontmatter validation, metadata and navigation preparation, local discovery, routes, draft exclusion, and Markdown preparation. |
| `internal/page` | Prepared page data shared by content preparation and rendering. |
| `internal/render` | Embedded skeleton loading and page composition. |
| `internal/output` | Destination checks and confined new-or-empty output writes. |

Dependencies point from the CLI and generator toward content preparation,
rendering, and output handling. The renderer depends only on the render
model, not on content preparation internals. Interfaces appear at a real
test seam or a second implementation.

Prefer the Go standard library, including `html/template` for rendering.
The theme skeleton is embedded in the executable; built-in themes follow
in phase 8. The [user reference](USER.md) describes current limitations,
and the [theme reference](THEMES.md) owns the initial render-model fields.
See the [developer guide](DEVELOPER.md) for commands and layout.

## Architecture decision records

The [decision index](PRD.md#decision-records) lists all accepted ADRs.
The most relevant records for this structure are:

- [0003: Free-form content and directory-index routes](adr/0003-content-tree-and-routes.md).
- [0006: Configuration and metadata model](adr/0006-configuration-and-metadata-model.md).
- [0007: Theme packages and integration](adr/0007-theme-packages-and-integration.md).
- [0008: Generation pipeline and page renderer](adr/0008-rendering-and-template-trust-boundary.md).
- [0010: Complete output, ownership, and recovery](adr/0010-staged-output-and-ownership.md).
- [0014: Public origin and deployment path](adr/0014-public-origin-and-deployment-path.md).
