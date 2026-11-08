# Highlight code at build time with theme-owned colors

Status: accepted.

Date: 2026-10-20.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Business sites sometimes show code, such as configuration or embed
snippets. Highlighting helps readers, but Cress sites ship no JavaScript,
and colors must fit each theme and its contrast.

## Decision drivers

- Keep generated pages free of JavaScript.
- Let themes own colors and contrast.
- Never guess a language or alter code.
- Use an established library.

## Considered options

- Highlight at build time with Chroma and CSS classes.
- Highlight at build time with Chroma and inline styles.
- Leave highlighting to a client-side script.

## Decision outcome

Chosen option: "Highlight at build time with Chroma and CSS classes",
because it needs no JavaScript and leaves colors to the theme.

[Chroma](https://github.com/alecthomas/chroma) highlights a fenced code
block only when its info string names a known language or alias exactly.
Other blocks stay plain and escaped. Each block keeps a `language-X`
class. The `[syntax]` table in `cress.toml` can turn highlighting off.

### Consequences

- Good, because pages stay static and work without JavaScript.
- Good, because each theme controls highlight colors and contrast.
- Good, because wrong guesses cannot mislabel code.
- Bad, because themes must style the highlight classes.
- Bad, because unlabeled code blocks stay plain.

## Confirmation

Unit tests cover known languages, aliases, unknown and missing languages,
escaping, and disabled highlighting. Each shipped theme styles the
highlight classes in its default pages.

## More information

See [theme integration](0007-theme-packages-and-integration.md) and
[rendering responsibilities](0008-rendering-and-template-trust-boundary.md).
