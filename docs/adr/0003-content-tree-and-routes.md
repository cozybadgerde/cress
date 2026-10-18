# Use a free-form content tree and directory-index routes

Status: accepted.

Date: 2026-10-15.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Small-business sites organize content around services, menus, information,
and occasional articles. Requiring separate pages and articles trees would
impose categories that do not describe every site's needs.

Clean public routes must work on static hosts without custom extensionless
file rewrites. Content assets also need an understandable ownership model.

## Decision drivers

- Let authors use business-relevant directory names.
- Make source-to-output mapping predictable.
- Separate site content from configuration and theme sources.
- Support domain-root and URL-prefix hosting.
- Keep page behavior explicit rather than inferred from special folders.

## Considered options

- Require separate pages and articles trees.
- Use a free-form content tree with directory-index output.
- Generate sibling HTML files and depend on host URL rewriting.

## Decision outcome

Chosen option: "Use a free-form content tree with directory-index output",
because it supports varied business sites while keeping portable routes.

Content organization guides routes without prescribing business categories.
Section index pages own their section's root route; other pages have
their own routes. A section index does not imply a landing-page type.
Explicit page metadata selects content behavior independently of folder
names.

List pages use logical direct children of their containing section, not a
filesystem-depth scan. Equivalent source layouts for a page must not change
its membership. List and landing types are not restricted to section index
files. [The list-page requirements](../PRD.md#list-pages) own eligibility,
ordering, visibility, and introduction behavior.

Keep page-owned assets identifiable and separate from content sections.
Publish them with their owning pages, preserving relative organization.
Shared static assets are copied without transformation and without
exposing a source-only directory prefix.

Resolve source-relative Markdown links and images to their published
destinations.
Generated links must work at both a domain root and a configured prefix.
[The URL policy](0014-public-origin-and-deployment-path.md) owns public
identity, deployment paths, and site-root reference semantics.

Reject duplicate destinations and file/directory conflicts across pages,
generated metadata, page assets, shared static files, and public theme
assets. Diagnostics identify the competing sources; precedence and build
order must not silently resolve collisions.

Remaining source conventions, normalization rules, and path details belong
to the authoring and routing implementation. They must preserve the agreed
routes, list membership, ownership, and portability boundaries.

### Minimal example

An illustrative content tree, not a prescribed layout:

```text
content/
  index.md
  about.md
  services/
    consulting.md
```

The consulting page can produce `services/consulting/index.html`, served
at `/services/consulting/`. The `services` folder is an ordinary content
directory, not an implicit page type.

### Consequences

- Good, because business-relevant directories remain ordinary content.
- Good, because directory indexes avoid provider-specific rewrite rules.
- Good, because asset ownership is distinct from content organization.
- Good, because list membership follows page hierarchy rather than file
  layout.
- Bad, because sources can claim the same destination and need collision
  diagnostics.

## Confirmation

When routing and assets are implemented, check free-form content and
page-owned assets at both a domain root and a URL prefix.
Confirm predictable routes, explicit page behavior, source protection,
and collision errors that identify both sources.
Check equivalent page layouts retain the same logical list membership,
without recursively including deeper descendants.

## More information

See [the product requirements](../PRD.md) and
[output ownership and recovery](0010-staged-output-and-ownership.md).
