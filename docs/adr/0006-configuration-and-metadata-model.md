# Use central site configuration with explicit page metadata overrides

Status: accepted.

Date: 2026-10-13.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

People configuring a site need understandable settings without depending
on theme internals. Page metadata needs predictable defaults without
hidden directory-level cascades.

These concerns are independent of the configuration's serialization format.

## Decision drivers

- Keep central site concerns understandable and distinct.
- Separate site settings, page metadata, and theme presentation.
- Let pages override documented site metadata defaults.
- Avoid directory-level inheritance and folder-dependent page behavior.
- Resolve content behavior before presentation.

## Considered options

- Central site configuration with explicit page metadata overrides.
- Site-wide defaults, directory-level defaults, and page overrides.

## Decision outcome

Chosen option: "Central site configuration with explicit page metadata
overrides", because it keeps site settings and metadata resolution
predictable without directory-level cascades.

The filenames below illustrate the roles, not required naming conventions.

### Site configuration, for example `config.toml`

- Holds site-wide settings: identity, language, brand, and documented
  metadata defaults.
- Keeps generator-owned Markdown options separate from theme customizations.
- Does not define individual pages' content, type, or draft status.

Navigation is site data, not theme customization.
[The navigation decision](0013-explicit-site-navigation.md) owns its behavior.

### Page content, for example `about.md`

- The Markdown body contains the page's content.
- Frontmatter supplies page-specific metadata and overrides documented
  inheritable site defaults.
- Page type and draft status belong to that page. They do not inherit from
  site defaults or depend on directory names.

The generator resolves these values before passing them to themes.
There is no directory-level metadata inheritance. Themes present resolved
values rather than implement inheritance or Markdown rendering policy.
The [PRD](../prd.md#authoring-and-content) records agreed fields and
behavior. Remaining field definitions and defaults belong to the authoring
implementation.

### Consequences

- Good, because ordinary site settings do not depend on template expertise.
- Good, because metadata does not depend on directory cascades.
- Good, because themes receive consistent resolved site and page data.
- Bad, because authors cannot use directory-level metadata defaults.
- Neutral, because field-level inheritance still belongs to the concrete
  authoring contract.

## Confirmation

When configuration and metadata are implemented, confirm documented site
defaults and explicit page overrides without directory inheritance.
Check that site settings, Markdown policy, page type, and draft behavior
remain independent of theme internals and business folder names.

## More information

See [authoring formats](0005-toml-authoring-formats.md),
[shared contracts](0004-shared-contract-policy.md),
[content organization](0003-content-tree-and-routes.md),
[the rendering boundary](0008-rendering-and-template-trust-boundary.md),
and [public URL policy](0014-public-origin-and-deployment-path.md).
