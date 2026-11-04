# Cress architecture

This guide describes Cress's agreed structure: its boundary, main
responsibilities, data flow, and failure boundaries. The
[PRD](PRD.md) owns product behavior. The [ADRs](adr/) own decision
rationale. This guide links to both instead of repeating them.

Status: implemented generator structure. Poppy, Clover, Anemone, site/theme
initialization, complete generation, recoverable output replacement, and
isolated preview are available. [Release tooling](RELEASE.md) is separate
from the offline executable.

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

A build runs these stages in order. Preparation failures leave previous
output untouched. Publication failures restore verified output where possible;
cleanup errors after confirmed publication leave the new result in place.

1. **Acquire.** Take the site-wide OS advisory writer lock and recover only
   verified interrupted transactions. Hold ownership for the whole build.
2. **Load.** Read the site configuration, Markdown files, page assets,
   and shared static files. Parse TOML configuration and frontmatter.
3. **Validate.** Reject malformed input and invalid known fields. Warn
   about unknown fields. Keep open theme customizations as custom data.
4. **Resolve.** Apply site metadata defaults and explicit page overrides,
   page types, and draft visibility. Derive directory-index routes and
   resolve generator-owned URLs beneath the deployment path.
5. **Prepare.** Convert Markdown to HTML, resolve Markdown links and
   images, highlight code, render captions, build list entries, and
   prepare navigation and publication metadata.
6. **Select theme.** Use a same-name local package when one exists, and
   a built-in theme only when it is absent. An invalid local package
   fails. Check compatibility, required page-type support, and labels
   before rendering.
7. **Render.** Compose each page from the render model and theme
   templates.
8. **Stage.** Write pages, sitemap, `robots.txt`, page assets, shared
   static files, and public theme assets into a staging area. Reject
   destination collisions and name the competing sources.
9. **Publish.** Verify ownership of the target and replace the previous
   output as a complete result. Restore it on
   failure where possible.

Current content preparation first inventories pages and designated assets,
then resolves visible sources through a complete source-to-public-URL map.
Excluded drafts and their assets never enter that map. Explicit
draft-inclusive preview prepares included drafts normally.
Only after resolution and
Markdown preparation does the generator compose pages and submit all page
and asset destinations to the output collision checker.

List preparation selects logical direct children from visible ordinary
pages after image and caption URLs are resolved. It supplies metadata-only
entries in their final order. Publication preparation creates canonical
URLs and XML sitemap/robots files without inspecting filesystem clocks;
generated files share the same collision checker as local assets.

Content preparation can run and be tested without templates. The renderer
can be tested with a hand-built render model.

### Output transaction

`internal/output.Writer` binds to one confined output target at construction.
A separate private site-lock owner holds an OS `flock` on the persistent
`.cress/writer.lock` file. Normal builds retain that lock with their writer;
preview attempts retain it separately while writing to their session target.
Canonical site aliases share that inode.
A private versioned record pairs the current output token with
`public/.cress-output.json`; the marker's path is reserved by the collision
checker. New and empty destinations need no previous ownership token.

Before staging, the record names a new random token, the previous token,
and whether a public directory exists. A complete stage advances from
`staging` to `ready`. Publication moves previous output to the recorded
backup and the stage to `public/`. Verified promotion advances to
`published` before backup cleanup. Each state update writes an exclusive
`output.next` and renames it over `output.json`.

Recovery checks real directories, regular marker files, matching tokens,
and allowed path combinations. It restores interrupted previous output or
retains confirmed new output, never promotes a merely staged directory.
Unexpected paths or an uncommitted next record stop with preserved
evidence. Cleanup failure remains an error without rolling back published
output. The [user reference](USER.md#output-protection) owns migration and
manual-recovery guidance.

## Preview flow

Preview reuses the full build flow for every rebuild; incremental
rendering is not required.

- `cress site preview` fixes draft visibility for the running session.
  Native fsnotify events cover the configuration and source trees, not
  deployable output or private writer state.
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

### Preview implementation

`internal/output.Session` creates owner-only targets in a private registry
under OS temporary storage. The registry name includes the user ID and a
hash of the canonical site path. Each randomly named session carries a
site-bound identity record and an OS advisory lifetime lock. Registry
changes and reclamation happen under the shared site writer lock. Only
recognized, inactive sessions with verified output/recovery state are
removed; unknown or ambiguous evidence is preserved and reported.

A preview build takes the same site writer lock as a normal build, but
opens only its session's output state, not deployable recovery state.
Generation prepares the same pages, assets, and templates, with an explicit
draft visibility setting. Writer outcomes include a verified output token
so a failed attempt can distinguish restored or confirmed output from an
unusable target. Output errors remain errors even when a token is usable.

`internal/preview` owns loopback HTTP serving, watcher lifetime, and rebuild
scheduling. Directory watches detect atomic saves and are updated when
source directories are created, moved, or removed. A buffered dirty signal
retains changes during a synchronous build. Debouncing coalesces edits;
busy writers schedule timed retries, while other failures await an edit.
Watcher failure stops rather than silently losing edits.

An HTTP read gate prevents requests from using output while publication
or restoration runs. Serving holds no site writer lock. A changed output
token reopens the verified root; restored output retains the old root.
Serving honors the configured deployment prefix, hides the ownership
marker, and disables directory listings. Signal cancellation stops the
watcher and server before verified session cleanup.

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
| `internal/content` | Site/frontmatter validation, metadata, navigation, source and asset discovery, routes, draft exclusion, URL resolution, captions, and Markdown highlighting. |
| `internal/page` | Prepared page data shared by content preparation and rendering. |
| `internal/render` | Theme-template parsing and page composition. |
| `internal/theme` | Package discovery, manifests, compatibility, labels, public assets, and offline initialization. |
| `internal/output` | Destination checks, confined staging, ownership, OS writer locking, replacement, and verified recovery. |
| `internal/preview` | Loopback serving, native source watching, serialized rebuilds, retries, and preview lifecycle coordination. |
| `internal/scaffold` | Non-overwriting site initialization from embedded example sources. |

Dependencies point from the CLI and generator toward content preparation,
rendering, and output handling. The renderer depends only on the render
model, not on content preparation internals. Interfaces appear at a real
test seam or a second implementation.

Prefer the Go standard library, including `html/template` for rendering.
The theme skeleton, all three complete built-in packages, and site scaffold are
embedded in the executable. The [user reference](USER.md) describes usage,
and the [theme reference](THEMES.md) owns the render-model fields.
See the [developer guide](DEVELOPER.md) for commands and layout.

Release hooks use go-licenses to generate CSV dependency-license reports
and notice text. Outputs are release artifacts, not checked-in catalogues
or application code. Cosign signs the checksum manifest covering archives
and reports.
See [release maintenance](RELEASE.md).

## Architecture decision records

The [decision index](PRD.md#decision-records) lists all accepted ADRs.
The most relevant records for this structure are:

- [0003: Free-form content and directory-index routes](adr/0003-content-tree-and-routes.md).
- [0006: Configuration and metadata model](adr/0006-configuration-and-metadata-model.md).
- [0007: Theme packages and integration](adr/0007-theme-packages-and-integration.md).
- [0008: Generation pipeline and page renderer](adr/0008-rendering-and-template-trust-boundary.md).
- [0010: Complete output, ownership, and recovery](adr/0010-staged-output-and-ownership.md).
- [0014: Public origin and deployment path](adr/0014-public-origin-and-deployment-path.md).
