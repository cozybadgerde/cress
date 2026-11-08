# Add Anemone as the third and final bundled theme for 1.0

Status: accepted.

Date: 2026-10-30.

Decision-maker: Daniel Schier, Product Owner of Cozy Badger.

## Context and problem statement

Poppy and Clover serve freelancers and small businesses with largely
single-column, sans-serif layouts. A third theme should broaden the
out-of-the-box experience for makers, independent shops, cafes, and creative
businesses rather than offer another palette over the same composition.

The approved direction is a quiet paper/Scandi editorial theme. It must
remain useful without photographs and preserve the content/presentation
boundary established by [ADR 0008](0008-rendering-and-template-trust-boundary.md).

## Considered options

- Add one paper/Scandi theme.
- Add one glass-like theme.
- Bundle both additional themes.
- Keep only Poppy and Clover.

## Decision outcome

Chosen option: "Add one paper/Scandi theme", named Anemone after the
Japanese anemone, because it adds a distinct editorial presentation without
requiring imagery, scripts, or additional configuration.

Bundle exactly three themes for this scope: Poppy, Clover, and Anemone.
Poppy remains the default. This replaces only the two-theme counts in
[ADR 0007](0007-theme-packages-and-integration.md) and
[ADR 0012](0012-release-artifacts.md); complete-package, local-first
integration and embedded release packaging remain unchanged. The development
skeleton is not a selectable theme.

Anemone uses warm paper surfaces, dark ink, editorial system-serif headings,
and system-sans body text. A desktop side masthead sits beside the content;
narrow layouts use one column. Framed images have subtle offset shadows and
restrained rotation of images only, never text or captions. Narrow layouts
remove rotation. Title images keep their intrinsic proportions instead of
using a fixed crop. Image-free pages retain the same usable hierarchy.

Use the prepared brand color for underlined links and accents. Do not rely
on an arbitrary accent as a text background. Consume no custom settings.
Ship exact English and German labels, local CSS, no browser JavaScript,
and no downloaded fonts or other remote resources.

The [theme developer guide](../THEMES.md) and
[user guide](../USER.md#choosing-a-theme) own current package and
presentation details. Product scope belongs to the [PRD](../PRD.md#themes).

## Confirmation

Extend initialized-site, complete-page, localization, and default contrast
checks to all three themes. Verify ordinary, landing, and list pages,
the content tour, title and inline images with captions, image-free pages,
long content, deployment prefixes, and same-name local overrides.

Inspect rendered desktop and narrow pages. Record the new theme in the
existing accessibility assessment without treating visual inspection or
automated contrast tests as a complete human WCAG assessment.
