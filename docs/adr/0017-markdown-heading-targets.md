# Prepare Markdown heading targets independently of themes

Status: accepted.

Date: 2026-10-30.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Business pages need links to individual services, menu sections, and FAQs.
These targets should work across themes without requiring raw HTML or
coupling content links to a theme's presentation.

Goldmark provides heading attributes and automatic ID generation, but its
default source-order allocation does not give later explicit IDs priority
or reserve IDs used by theme and generated markup.

## Decision drivers

- Keep heading links independent of the selected theme.
- Prefer existing parser capabilities without another dependency.
- Let authors keep stable links when heading wording changes.
- Prevent ambiguous Markdown targets without inspecting trusted raw HTML.

## Considered options

- Built-in Goldmark ID generation with explicit-ID priority and reservations.
- A custom or third-party Unicode-aware slug generator.
- Only author-supplied raw HTML targets.

## Decision outcome

Chosen option: "Built-in Goldmark ID generation with explicit-ID priority
and reservations", because it adds useful links with a small implementation
and no dependency or configuration growth.

Content preparation parses heading attributes, validates and reserves all
explicit IDs, then uses Goldmark's ID generator for remaining headings.
The [product requirements](../PRD.md#heading-links) own observable behavior;
the [authoring reference](../USER.md#heading-links) owns syntax and exact
allocation rules.

The reserved `main` and `page-title` IDs preserve existing theme targets.
The `cress-*` namespace protects generated captions and provides a namespace
for additional theme structure. Themes must not introduce IDs that collide
with prepared headings outside those reservations.

Keep authored HTML opaque under
[the content trust policy](0015-trusted-markdown-and-raw-html.md).
Publishers must prevent collisions involving its IDs; Cress does not parse
or rewrite HTML to guarantee complete-document uniqueness.

### Consequences

- Good, because heading links survive a theme change.
- Good, because explicit IDs stay stable and take priority regardless of order.
- Good, because no new dependency, field, or browser interaction is needed.
- Bad, because automatic IDs drop non-ASCII characters and depend on wording.
- Bad, because arbitrary theme and authored HTML IDs still require coordination.

## Confirmation

Test automatic IDs, explicit overrides in either order, repeated headings,
reserved names, non-ASCII and empty headings, and source-aware failures.
Check both heading syntaxes and existing links, captions, and raw HTML.
Build initialized sites with both themes and verify their heading targets.
