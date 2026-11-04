# Organize CLI commands by scope

Status: accepted.

Date: 2026-10-06.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Site publishers and theme developers have different workflows. A site
publisher mainly works on a site, while a theme developer regularly works
on theme packages. Both may need actions such as initialization.

The CLI needs an understandable way to organize these operations as its
capabilities grow. Users should find their usual tasks without navigating
unrelated work.

## Decision drivers

- Make each audience's usual tasks easy to discover.
- Keep related operations and their help together.
- Give commands a predictable structure.
- Allow additional scopes without reorganizing existing workflows.
- Keep general information accessible without choosing a scope.

## Considered options

- Scope-first commands, such as `cress site init` and `cress theme init`.
- Action-first commands, such as `cress init site` and `cress init theme`.
- Flat commands without scope groups.

## Decision outcome

Chosen option: "Scope-first commands", because it groups work by what users
are working on rather than by actions shared across different workflows.

Use the structure `cress <scope> <action>`. Site operations belong under
`cress site`; theme operations belong under `cress theme`. Common help and
version information remain accessible at the root.

| Example | Scope and purpose |
|---------|-------------------|
| `cress site init` | Initialize a site for publishing. |
| `cress theme init` | Initialize a theme for development. |
| `cress config validate` | Illustrate a possible future configuration scope. |

These examples explain the grouping, not a complete command catalog.
A configuration scope and standalone validation commands are outside the
current 1.0 scope. This decision does not introduce either.

Action-first grouping puts site and theme initialization under the same
`init` group, mixing publisher and theme-development workflows.
Scope-first grouping lets users focus on the operations relevant to their
work. Scopes are not exclusive user roles; one person can use several.

Supported operations and their behavior belong to the PRD. Flags and other
invocation details belong to the CLI implementation and user reference.

### Consequences

- Good, because publishing and theme development have clear command groups.
- Good, because help can focus on the user's current scope.
- Good, because new scopes fit the same command structure.
- Bad, because shared actions appear under more than one scope.

## Confirmation

When the CLI is implemented, check that commands and help follow the
scope-first structure. Confirm site and theme operations stay in their own
groups and common help and version information remain accessible at the
root.

## More information

See [the product requirements](../PRD.md)
and [output protection](0010-staged-output-and-ownership.md).
