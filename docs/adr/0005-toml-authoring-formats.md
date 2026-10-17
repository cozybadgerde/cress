# Use TOML for site configuration and Markdown frontmatter

Status: accepted.

Date: 2026-10-06.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Other Cozy Badger products use TOML configuration. YAML frontmatter was an
initial candidate because it is common in Markdown tooling, not because
Cress needs YAML-specific features.

Supporting several authoring formats adds parsing, documentation,
validation, and compatibility obligations without an established need.

## Decision drivers

- Keep author-facing configuration consistent with Cozy Badger products.
- Minimize the number of supported authoring formats.
- Preserve structured page metadata directly in Markdown.
- Avoid another input-format dependency without a product need.

## Considered options

- TOML configuration and TOML frontmatter.
- TOML configuration and YAML frontmatter.
- TOML configuration and both YAML and TOML frontmatter.

## Decision outcome

Chosen option: "TOML configuration and TOML frontmatter", because no
accepted requirement needs YAML and one authoring format reduces the
supported surface.

Use TOML for site configuration and delimited frontmatter in Markdown.
This selects the input format, not configuration fields, metadata
inheritance, delimiters, or parser tooling.

Theme manifests are a separate package interface governed by
[the theme integration decision](0007-theme-packages-and-integration.md),
not another site configuration format.

If a future product requirement genuinely needs YAML, reconsider this
decision and support for both frontmatter formats. Development tools that
use YAML do not create a site-authoring requirement.

### Consequences

- Good, because site and page metadata share an authoring format.
- Good, because YAML support is unnecessary in the current scope.
- Bad, because YAML-based content may need conversion before use.
- Bad, because external validation tools need to parse TOML or extract it
  from Markdown frontmatter.

## Confirmation

When authoring is implemented, confirm TOML-only site configuration and
frontmatter with actionable input diagnostics.

## More information

See [configuration and metadata](0006-configuration-and-metadata-model.md)
and [shared contracts](0004-shared-contract-policy.md).
