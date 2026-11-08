# Expose standalone links as structure, not presentation

Status: accepted.

Date: 2026-10-30.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Poppy and Clover style the link-only paragraph after a landing page's
introduction as a primary action. CSS cannot distinguish that paragraph
from one containing a link with surrounding text: `:only-child` counts
elements, not text nodes.

Fixing this must preserve the boundary between prepared content and theme
presentation without introducing a generator-owned button convention.

## Considered options

- Add a neutral structural class to standalone-link paragraphs.
- Mark primary actions in content preparation.
- Require theme-specific HTML or browser JavaScript.

## Decision outcome

Chosen option: "Add a neutral structural class to standalone-link
paragraphs", because content preparation can reliably identify structure
while each theme retains control over presentation.

Add `standalone-link` to a Markdown paragraph whose only child is a direct
link, including a CommonMark autolink. Apply this regardless of page type,
position, or theme. Formatting inside the link is allowed; surrounding
text, additional links, and wrappers around the link do not qualify.
Only actual paragraphs receive the class, not tight-list text blocks.

Poppy and Clover consume the class only in their existing landing-page
position. Other themes may use it differently or ignore it entirely.
Do not add primary-action metadata, change link semantics, inject styles,
or interpret authored HTML.

The [theme developer guide](../THEMES.md#markup-inside-the-body) owns the
markup contract; the [built-in themes](../../internal/theme/builtin/) own
bundled presentation behavior. The broader content/presentation review remains
separate work.

## Confirmation

Test inline and reference links, autolinks, formatted labels, surrounding
text, multiple links, code, raw HTML, and nested paragraphs. Verify both
themes retain the intended standalone action without styling sentence
links as buttons.
