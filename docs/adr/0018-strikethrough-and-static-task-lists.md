# Use built-in Markdown extensions for strikethrough and static task lists

Status: accepted.

Date: 2026-10-30.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Authors need familiar syntax for deleted text and checklists without raw
HTML. Goldmark already provides both features, but its task-list renderer
does not associate each checkbox with the item's text.

## Decision drivers

- Prefer existing parser capabilities without new dependencies or settings.
- Preserve content behavior across themes and without browser JavaScript.
- Associate task text with the native checkbox for assistive technology.
- Keep static publication separate from interactive task management.

## Considered options

- Enable individual extensions and label the native task-list checkboxes.
- Enable the whole GFM bundle, including bare-URL linkification.
- Render custom checkbox controls or interactive task lists.

## Decision outcome

Chosen option: "Enable individual extensions and label the native task-list
checkboxes", because it meets the authoring need without unrelated parsing
changes or interactive behavior.

Enable Goldmark's strikethrough and task-list extensions individually.
Preserve its semantic `<del>` output and native disabled checkboxes.
Content preparation wraps each checkbox and its first paragraph's inline
content in a `<label>`, excluding subsequent paragraphs and nested lists.
This needs no generated IDs, translated status labels, or custom renderer.
Themes preserve this markup; the browser supplies native checkbox styling.

The [product requirements](../PRD.md#authoring-and-content) own behavior.
The [authoring reference](../USER.md#strikethrough-and-task-lists) owns syntax
and publisher guidance. Raw HTML remains publisher-controlled under
[the content trust policy](0015-trusted-markdown-and-raw-html.md).

### Consequences

- Good, because authors get familiar syntax without another dependency.
- Good, because native checkboxes expose completion state and their labels.
- Good, because links remain ordinary links without task-state persistence.
- Bad, because disabled checkbox presentation varies between browsers.
- Bad, because strikethrough announcement and checkbox reading behavior
  still require the documented human accessibility assessment.

## Confirmation

Test checked and unchecked items, both completion-marker cases, ordered
and nested lists, loose paragraphs, links, captions, and code/raw-HTML
boundaries. Verify label associations and unchanged ordinary lists.
Build the content tour with both themes and retain the human-review gate.
