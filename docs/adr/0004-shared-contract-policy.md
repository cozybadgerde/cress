# Keep supported contracts stable and schemas aligned with releases

Status: accepted.

Date: 2026-10-16.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Configuration, frontmatter, and theme interfaces need documented contracts
that authors and tooling can rely on. Validation must distinguish mistakes
from supported customization, and upgrades must not silently invalidate
existing documented workflows.

New releases may assign meaning to previously unknown fields; authors need
to know whether earlier warning-only acceptance is protected.

Authors should also be able to check inputs with tools of their choice.
This shared policy does not define domain data models or theme runtime
requirements.

## Decision drivers

- Make supported inputs discoverable to authors and tooling.
- Enable schema-based validation without prescribing an editor or workflow.
- Keep documentation, schemas, and runtime behavior aligned.
- Distinguish invalid known fields from unknown fields and open extensions.
- Preserve documented contracts across stable-major upgrades.
- Avoid a separate schema release lifecycle.

## Considered options

Contract descriptions:

- Human-readable references only.
- JSON Schema and human-readable references coupled to Cress releases.

Validation:

- Reject every unknown field.
- Reject invalid known fields, warn about unknown fields, and preserve
  designated open customizations.

Compatibility:

- Preserve documented contracts within a stable major release.
- Preserve warning-only acceptance of unknown fields until a new major.
- Permit breaking changes in minor releases with separate contract versions.

## Decision outcome

| Topic | Decision |
|-------|----------|
| Documentation | Publish JSON Schema and human-readable references for configuration, frontmatter, and theme manifests. |
| Versions | Schema versions match Cress releases and describe what that release validates. No separate schema lifecycle. |
| Validation | Invalid known fields fail generation with actionable errors. Unknown fields warn without failing an otherwise valid build. |
| External checks | Make published schemas usable by compatible validation tools, without requiring a particular editor or workflow. |
| Customizations | Keep a documented open area for theme customizations; their own schemas are not required. |
| Compatibility | Minor and patch releases within 1.x preserve supported fields, interfaces, behavior, and defaults. Breaking these contracts requires a new major, even for deprecated features. |
| New features | Older executables do not have to support features added in newer releases. |

Unknown fields outside the documented open customization area are
accepted with a warning, not treated as supported extensions. A minor
release may recognize such a field and validate it as a known field.
Previously warning-only inputs may then fail or produce different output.
This is not a breaking change to a supported contract. Patch releases do
not introduce newly recognized fields.

This boundary allows feature growth without freezing the meaning of every
unknown key. New fields must preserve documented behavior and defaults
when absent. Within a stable major, Cress must not claim keys inside the
documented open customization area as core fields.

Announce newly recognized fields and their validation rules in release
notes, with matching release schemas and human-readable references.

Invalid syntax or document structure is an error, not an unknown-field
warning. Keys inside the documented open customization area are custom
data, not unknown core fields.

Schemas validate parsed structures and values. They must accept otherwise
valid inputs with unknown fields and allow documented open customizations.
Cress produces unknown-field warnings during generation; external tools
are not required to produce the same warnings.

External tools must parse TOML or extract Markdown frontmatter as needed.
Runtime validation remains required for conditions that schemas cannot
check, such as theme compatibility and output ownership.
A schema version does not declare a theme's runtime compatibility.

Specific editor integrations, inline completion, and custom language servers
are not required. This policy does not add a validation subcommand.
Configuration and theme integration have their own domain ADRs.
Field definitions not already agreed in the PRD, schema dialect,
publication paths, and validation tools are implementation choices.

### Consequences

- Good, because authors and tools can discover supported inputs.
- Good, because authors can choose their own validation tools.
- Good, because schema identity tracks the corresponding generator.
- Good, because stable-major upgrades preserve documented workflows.
- Good, because minor releases can add fields without treating unknown keys
  as an extension API.
- Bad, because schemas, documentation, and runtime validation must agree.
- Bad, because incompatible contract improvements require a major release.
- Bad, because previously warning-only inputs can fail or change behavior
  after a minor upgrade.
- Bad, because custom values without schemas rely on theme documentation.
- Neutral, because integration depends on the chosen tool's format support.

## Confirmation

During implementation and release validation, confirm published schemas
can validate representative parsed configuration, frontmatter, and theme
manifest data. Check schema/runtime agreement, diagnostic behavior, and
compatibility of documented sites and themes across minor and patch upgrades.
Check recognition of previously unknown fields, including new behavior and
validation errors, minor-only introduction, protected open customizations,
and matching release notes.

## More information

See [authoring formats](0005-toml-authoring-formats.md),
[configuration and metadata](0006-configuration-and-metadata-model.md),
and [theme integration](0007-theme-packages-and-integration.md).
See [JSON Schema's additional-property rules](https://json-schema.org/understanding-json-schema/reference/object#additionalproperties)
for accepting fields outside the known schema.
