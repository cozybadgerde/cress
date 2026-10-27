# Integrate complete local theme packages with explicit compatibility

Status: accepted.

Date: 2026-10-15.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Cress needs curated built-in themes and third-party themes that work
offline. Package selection must be predictable, and public assets must
remain distinct from private sources.

Themes also need explicit runtime compatibility and labels for the site's
configured language. These are theme integration concerns, not shared
schema policy or template execution mechanics.

## Decision drivers

- Build from complete built-in or local packages without downloads.
- Support local customization without implicit file merging.
- Keep ordinary brand settings separate from expert customizations.
- Publish only designated public assets.
- Reject incompatible packages before rendering.
- Keep theme labels consistent with the configured site language.

## Considered options

- Complete local theme packages with separate sources and public assets.
- Only built-in themes with fixed presentation.
- Remote themes downloaded during generation.
- Copy entire theme packages into public output.

## Decision outcome

Chosen option: "Complete local theme packages with separate sources and
public assets", because it supports offline customization without
accidental source publication.

Embed the two built-in themes and a minimal renderable theme skeleton in
the executable. Built-in rendering and theme initialization need no
separate runtime files or downloads. The skeleton is developer scaffolding,
not another selectable built-in theme.

Use a consistent package contract for built-in themes, local themes, and
the skeleton. Separate the JSON manifest, templates, designated public
assets, and theme-owned TOML translations.
Publish only public assets in a theme-specific namespace, subject to
the same collision checks as other output.

A bundled theme remains a named presentation package, with the same
responsibilities as a third-party theme. Embedding changes how the package
is distributed, not its role or access to generator internals.
[The rendering boundary](0008-rendering-and-template-trust-boundary.md)
defines generator and presentation ownership.

Themes own styling, including responsive layout and overflow behavior.
Content preparation supplies semantic HTML and useful class hooks without
prescribing presentation or injecting styles. Structural wrappers may
provide keyboard access; themes decide how to present them.

Ordinary site and brand settings remain separate from the open theme
customization area established by the shared contract policy.
Themes explicitly consume and document the custom values they support.

### Complete local-first selection

Resolve the selected name locally before using a built-in theme.
A same-name local package replaces the complete built-in package.
Do not merge packages or inherit missing files.

Use a built-in theme only when the local package is absent. Invalid,
incompatible, incomplete, or unreadable local packages fail without
fallback. A lookup error is not evidence of absence.

Maintainers obtain and update local packages separately. Cress does not
download them or update them with built-in theme releases.

### Runtime compatibility

Each theme declares a minimum Cress version and one supported major.
A stable runtime must match that major and meet the minimum version.
The minimum must belong to the declared major; arbitrary compatibility
ranges are not part of this model.

Validate compatibility before rendering. Missing, invalid, or incompatible
declarations fail with actionable diagnostics.
A manifest's schema release identity does not replace the theme's runtime
compatibility declaration.

### Theme-owned labels

Use the theme's plain-text label map for the configured site language,
with exact language matching and no default-language or regional fallback.

An unavailable required language or a requested missing label fails
clearly. Do not substitute another language, an empty value, or the key
name. Themes own their label keys; no universal catalogue is required.
A theme that requests no translated labels needs no translation file.

Exact manifest fields, package layouts, label access, and prerelease or
development-runtime semantics belong to implementation specifications.

### Consequences

- Good, because local customization remains complete and predictable.
- Good, because source files stay separate from public output.
- Good, because compatibility failures occur before rendering.
- Good, because missing labels fail instead of silently substituting.
- Bad, because maintainers manage third-party packages separately.
- Bad, because unschematized custom values rely on theme documentation.
- Bad, because unsupported languages require matching theme labels.

## Confirmation

When theme integration is implemented, confirm offline initialization,
built-in and local selection, public/private boundaries, runtime
compatibility, and explicit translation failures.

## More information

See [shared contracts](0004-shared-contract-policy.md),
[configuration and metadata](0006-configuration-and-metadata-model.md),
[the rendering boundary](0008-rendering-and-template-trust-boundary.md),
and [release distribution](0012-release-artifacts.md).
