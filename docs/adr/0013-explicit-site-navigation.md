# Configure site navigation explicitly in display order

Status: accepted.

Date: 2026-10-15.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

People configuring a site need to choose which links visitors see and in
what order. The content tree alone does not express the intended main and
footer navigation, and changing themes should not redefine that choice.

## Decision drivers

- Make navigation selection and order predictable.
- Keep navigation independent of theme customization.
- Avoid inferring menus from content organization.
- Support internal page links and external web links.
- Keep the initial navigation model simple.

## Considered options

- Explicit navigation entries in configuration order.
- Navigation discovered automatically from the content tree.
- Explicit entries ordered through separate weights.

## Decision outcome

Chosen option: "Explicit navigation entries in configuration order",
because people configuring a site should control its menus without
content-discovery rules or a separate ordering mechanism.

- Navigation is site data, not theme customization.
- Use flat main and footer groups. Configuration order is display order.
- Support internal page links and external HTTP/HTTPS links.
- The generator resolves internal links beneath the deployment prefix;
  external links are not rewritten.
- Each theme renders populated groups and omits empty groups.

Themes own navigation markup and styling, not selection or ordering of
the configured entries.

[The URL policy](0014-public-origin-and-deployment-path.md) defines the
deployment path and site-root reference semantics.

Invalid known field types and values follow the shared contract policy.
Concrete fields and configuration syntax belong to implementation.

### Consequences

- Good, because menu contents and ordering remain explicit.
- Good, because navigation does not depend on folder names or theme settings.
- Bad, because adding a content page does not automatically add a menu entry.
- Bad, because internal navigation targets are not checked for existence.

## Confirmation

When navigation is implemented, confirm explicit selection, configuration
order, main and footer rendering, and omission of empty groups.
Exercise the same navigation contract with bundled and local themes.
Check internal links beneath a deployment prefix and unchanged external
links, with actionable diagnostics for invalid known values.

## More information

See [configuration and metadata](0006-configuration-and-metadata-model.md),
[shared contracts](0004-shared-contract-policy.md), and
[the rendering boundary](0008-rendering-and-template-trust-boundary.md).
