# Follow accessibility best practices for generated pages

Status: accepted. Revised on 2026-11-07: a lighter review replaces the
original WCAG assessment gate.

Date: 2026-10-15.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Cress produces markup that themes combine with publisher content. Visitors
with different abilities and devices should be able to use the result.
The project needs a clear accessibility expectation that a one-person
open-source project can keep up with, and a clear split of who owns what.

## Decision drivers

- Make accessibility part of every release, not a one-off effort.
- Keep the effort proportionate to a small project.
- Preserve generator, theme, and publisher ownership.
- Combine automated checks with a short human review.

## Considered options

- Follow best practices, with WCAG 2.2 Level AA as a reference.
- Assess default pages against every WCAG 2.2 A and AA criterion, with a
  documented assessment as a release gate.
- Rely on automated accessibility checks alone.

## Decision outcome

Chosen option: "Follow best practices, with WCAG 2.2 Level AA as a
reference", because it keeps accessibility a release concern at a cost the
project can sustain. The first version of this record chose the full
assessment gate; it proved too heavy for a one-person project.

Default pages from each shipped theme follow accessibility best practices.
The focus is keyboard use, visible focus, meaningful structure, contrast,
resizing, and narrow screens.

Apply the [existing ownership boundary](0008-rendering-and-template-trust-boundary.md):

- Cress owns the markup it generates from Markdown and preserves
  author-supplied accessibility information, such as image alternatives.
- Each theme, bundled or third-party, owns its page structure, styling,
  focus presentation, contrast, and responsive behavior.
- Publishers own their content, colors, custom HTML, and styles. Authored
  HTML stays [trusted and unchanged](0015-trusted-markdown-and-raw-html.md).

Before each minor or major release, automated checks and a short human
review cover the shipped themes. Fix barriers found before tagging.

### Consequences

- Good, because accessibility is checked on every relevant release.
- Good, because the review fits into the release routine.
- Good, because ownership tells publishers and theme authors what is theirs.
- Bad, because a short review can miss issues a full assessment would find.
- Neutral, because publishers remain responsible for their finished sites.

## Confirmation

The [release guide](../RELEASE.md#2-go-through-the-pre-tag-checklist)
includes the accessibility checks. The
[accessibility guide](../ACCESSIBILITY.md) explains what Cress and its
themes take care of and what publishers can do.

## More information

See [the accessibility product requirements](../PRD.md#accessibility) and
[theme integration](0007-theme-packages-and-integration.md).
[Evaluating web accessibility](https://www.w3.org/WAI/test-evaluate/)
explains why automated checks need human review.
